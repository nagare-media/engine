/*
Copyright 2022-2024 The nagare media authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package sleep

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/log"

	backoff "github.com/cenkalti/backoff/v4"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/internal/functions"
	enginehttp "github.com/nagare-media/engine/internal/pkg/http"
	"github.com/nagare-media/engine/internal/pkg/mime"
	"github.com/nagare-media/engine/internal/pkg/uuid"
	"github.com/nagare-media/engine/pkg/events"
	"github.com/nagare-media/engine/pkg/nbmp"
	nbmputils "github.com/nagare-media/engine/pkg/nbmp/utils"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

// Function description
const (
	Name = "mmsys-test-scene-detection"

	TestParameterKey                        = "mmsys-test-scene-detection.engine.nagare.media/test"
	MaxNumberOfSimulatedCrashesParameterKey = "mmsys-test-scene-detection.engine.nagare.media/max-number-of-simulated-crashes"
	SimulatedCrashWaitDurationParameterKey  = "mmsys-test-scene-detection.engine.nagare.media/simulated-crash-wait-duration"

	DefaultMaxNumberOfSimulatedCrashes = 1
	DefaultSimulatedCrashWaitDuration  = 20 * time.Second
)

const (
	// * baseline:         scene detection
	// * test-no-recovery: scene detection with n disruptions
	// * test-recovery:    scene detection with n disruptions and recovery

	Baseline       = "baseline"
	TestNoRecovery = "test-no-recovery"
	TestRecovery   = "test-recovery"
)

var (
	ffmpegScdetRegex = regexp.MustCompile(`.*scdet.*lavfi\.scd\.score:\D+(?P<score>\d+\.\d+),.*lavfi\.scd\.time:\D+(?P<time>\d+\.\d+)`)
	scoreIndex       = ffmpegScdetRegex.SubexpIndex("score")
	timeIndex        = ffmpegScdetRegex.SubexpIndex("time")
)

const (
	SceneDetectedEventType = "media.nagare.engine.v1alpha1.functions.mmsys-test-scene-detection.scene-detected"
)

type SceneDetectionEventData struct {
	Time float64
}

// function mmsys-test-scene-detection implements various video encoding test for the evaluation portion of an MMSys paper.
type function struct {
	// meta
	task           *nbmpv2.Task
	workflowID     string
	taskID         string
	instanceID     string
	instanceNumber int

	// config
	test                        string
	maxNumberOfSimulatedCrashes int
	simulatedCrashWaitDuration  time.Duration

	// input
	inVideoURL string

	// output
	awsAccessID        string
	awsSecretAccessKey string
	s3Insecure         bool
	s3Endpoint         string
	s3Bucket           string
	s3ObjectKey        string

	s3           *minio.Client
	reportClient events.Client
}

var _ nbmp.Function = &function{}

// Exec mmsys-test-scene-detection function.
func (f *function) Exec(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	l := log.FromContext(ctx, "test", f.test).WithName(Name)
	ctx = log.IntoContext(ctx, l)

	l.Info(fmt.Sprintf("executing %s", f.test))

	switch f.test {
	case Baseline:
		return f.baseline(ctx)

	case TestNoRecovery:
		return f.testNoRecovery(ctx)

	case TestRecovery:
		return f.testRecovery(ctx)

	default:
		return fmt.Errorf("unknown test '%s'", f.test)
	}
}

func (f *function) baseline(ctx context.Context) error {
	l := log.FromContext(ctx)

	scenes, err := f.detectScenes(ctx, 0, f.inVideoURL)
	if err != nil {
		l.Error(err, "scene detection failed")
		return err
	}

	err = f.outputScenes(ctx, scenes, f.s3ObjectKey)
	if err != nil {
		l.Error(err, "writing scenes to S3 failed")
		return err
	}

	return nil
}

func (f *function) testNoRecovery(ctx context.Context) error {
	l := log.FromContext(ctx)

	// count number of previous task instances for simulated crash
	_ = f.syncSceneDetectionEvents(ctx)
	l.Info(fmt.Sprintf("previous task instances: %d", f.instanceNumber))
	f.setupSimulatedCrash(ctx)

	return f.baseline(ctx)
}

func (f *function) testRecovery(ctx context.Context) error {
	l := log.FromContext(ctx)

	// sync with previous instances of this task
	l.Info("recover from previous task instances")
	scenes := f.syncSceneDetectionEvents(ctx)
	if f.instanceNumber > 0 {
		l.Info(fmt.Sprintf("synced with %d previous task instance(s)", f.instanceNumber))
	} else {
		l.Info("no previous task instances found")
	}

	// simulate crash
	f.setupSimulatedCrash(ctx)

	// detect scenes
	seekSecond := 0.0
	if n := len(scenes); n > 0 {
		seekSecond = scenes[n-1]
	}

	newScenes, err := f.detectScenes(ctx, seekSecond, f.inVideoURL)
	if err != nil {
		l.Error(err, "scene detection failed")
		return err
	}

	// merge with previously detected scenes
	scenes = append(scenes, newScenes...)

	// output scenes
	err = f.outputScenes(ctx, scenes, f.s3ObjectKey)
	if err != nil {
		l.Error(err, "writing scenes to S3 failed")
		return err
	}

	return nil
}

func (f *function) detectScenes(ctx context.Context, seekSeconds float64, inUrl string) ([]float64, error) {
	l := log.FromContext(ctx)
	l.Info(fmt.Sprintf("detect scenes starting at %.3f in %s", seekSeconds, inUrl))

	if seekSeconds < 0 {
		return nil, fmt.Errorf("seek cannot be negative: %.3fs", seekSeconds)
	}

	var (
		err    error
		scenes []float64
	)

	// open /dev/null
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0755)
	if err != nil {
		return nil, err
	}

	// FFmpeg process
	ffmpegArgs := []string{"-hide_banner", "-nostats"}
	if seekSeconds > 0 {
		ffmpegArgs = append(ffmpegArgs, "-ss", fmt.Sprintf("%.3f", seekSeconds))
	}
	ffmpegArgs = append(ffmpegArgs,
		"-copyts",
		"-reconnect", "1",
		"-reconnect_on_network_error", "1",
		"-reconnect_on_http_error", "1",
		"-reconnect_delay_max", "5",
		"-i", inUrl,
		"-filter:v", "scdet=threshold=14.0",
		"-an",
		"-f", "null",
		"-",
	)

	ffmpegCmd := exec.CommandContext(ctx, "ffmpeg", ffmpegArgs...)
	ffmpegCmd.Stdout = devNull // ignore stdout
	ffmpegPipe, err := ffmpegCmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	ffmpegCmd.Cancel = func() error {
		if runtime.GOOS == "windows" {
			return ffmpegCmd.Process.Signal(os.Kill)
		}
		go func() {
			time.Sleep(30 * time.Second)
			_ = ffmpegCmd.Process.Signal(os.Kill)
		}()
		return ffmpegCmd.Process.Signal(os.Interrupt)
	}

	// start FFmpeg
	l.V(1).Info(fmt.Sprintf("running command %s", ffmpegCmd))
	if err = ffmpegCmd.Start(); err != nil {
		l.Error(err, "FFmpeg failed")
		return nil, err
	}

	// scan output for detected scenes
	s := bufio.NewScanner(ffmpegPipe)
	for s.Scan() {
		matches := ffmpegScdetRegex.FindStringSubmatch(s.Text())
		if matches == nil {
			continue
		}

		// found a scene
		time, err := strconv.ParseFloat(matches[timeIndex], 64)
		if err != nil {
			l.Error(err, "parsing scene detection failed; fix implementation")
			return nil, err
		}
		scenes = append(scenes, time)

		f.observeSceneDetectionEvent(ctx, SceneDetectedEventType, SceneDetectionEventData{
			Time: time,
		})

		l.Info("detected a scene", "time", time, "score", matches[scoreIndex])
	}

	// call Wait on cmds to clean up resources
	if err = ffmpegCmd.Wait(); err != nil {
		l.Error(err, "FFmpeg failed")
		return nil, err
	}

	return scenes, nil
}

func (f *function) outputScenes(ctx context.Context, scenes []float64, outObjKey string) error {
	l := log.FromContext(ctx)

	outObjURL := strings.Join([]string{f.s3.EndpointURL().String(), f.s3Bucket, outObjKey}, "/")
	l.Info(fmt.Sprintf("output to %s", outObjURL))

	file := &bytes.Buffer{}
	for _, time := range scenes {
		fmt.Fprintf(file, "%.2f\n", time)
	}

	_, err := f.s3.PutObject(ctx, f.s3Bucket, outObjKey, file, int64(file.Len()), minio.PutObjectOptions{ContentType: "text/plain"})
	if err != nil {
		l.Error(err, "S3 upload failed")
		return err
	}

	return nil
}

func (f *function) syncSceneDetectionEvents(ctx context.Context) []float64 {
	l := log.FromContext(ctx)

	var (
		scenes                 = make([]float64, 0)
		sceneDetectionEventsCh = make(chan cloudevents.Event)
		synceDone              = make(chan struct{})
	)

	go f.execEventAPIServer(ctx, sceneDetectionEventsCh, synceDone)

	for {
		select {
		case <-synceDone:
			close(sceneDetectionEventsCh)
			return scenes

		case me := <-sceneDetectionEventsCh:
			if me.Type() != SceneDetectedEventType {
				continue
			}

			sded := &SceneDetectionEventData{}
			if err := me.DataAs(sded); err != nil {
				l.Error(err, "failed to decode scene detection event data")
				continue
			}

			scenes = append(scenes, sded.Time)
		}
	}
}

func (f *function) execEventAPIServer(ctx context.Context, sceneDetectionEventsCh chan<- cloudevents.Event, synceDone chan struct{}) {
	l := log.FromContext(ctx)

	srv := enginehttp.NewServer(&enginev1.WebserverConfig{
		BindAddress:   ptr.To(":8080"),
		ReadTimeout:   &metav1.Duration{Duration: 1 * time.Minute},
		WriteTimeout:  &metav1.Duration{Duration: 1 * time.Minute},
		IdleTimeout:   &metav1.Duration{Duration: 1 * time.Minute},
		Network:       ptr.To("tcp"),
		PublicBaseURL: ptr.To("http://127.0.0.1:8080"),
	})

	eventCh := make(chan cloudevents.Event)
	events.API(eventCh).MountTo(srv.App)

	l.Info("start event API server")
	var (
		srvErr  error
		srvDone = make(chan struct{})
	)
	go func() { srvErr = srv.Start(ctx); close(srvDone) }()

	synced := false
	subj := fmt.Sprintf("/engine.nagare.media/workflow(%s)/task(%s)/instance(%s)", f.workflowID, f.taskID, f.instanceID)

	for {
		select {
		case <-ctx.Done():
			l.Info("terminate event API server")
			return

		case <-srvDone:
			if srvErr != nil {
				panic(fmt.Sprintf("event API server failed: %s", srvErr))
			}

		case e := <-eventCh:
			// ignore events after synced is done
			if synced {
				continue
			}

			switch e.Type() {
			case SceneDetectedEventType:
				sceneDetectionEventsCh <- e

			case events.TaskCreated:
				if e.Subject() != subj {
					f.instanceNumber++
					continue
				}

				// we are synced up to this instance
				synced = true
				close(synceDone)
			}
		}
	}
}

func (f *function) setupSimulatedCrash(ctx context.Context) {
	l := log.FromContext(ctx)

	if f.instanceNumber >= f.maxNumberOfSimulatedCrashes {
		l.Info("reached max number of simulated crashes: disable simulated crash")
		return
	}

	// simulate hard crash
	l.Info("enable simulated crash")
	go func() {
		time.Sleep(f.simulatedCrashWaitDuration)
		l.Info("simulate crash now")
		os.Exit(1)
	}()
}

func (f *function) observeSceneDetectionEvent(ctx context.Context, t string, me SceneDetectionEventData) {
	l := log.FromContext(ctx)

	e := cloudevents.NewEvent()
	e.SetID(uuid.UUIDv4())
	e.SetType(t)
	e.SetSource("/engine.nagare.media/functions/mmsys-test-scene-detection")
	e.SetSubject(fmt.Sprintf("/engine.nagare.media/workflow(%s)/task(%s)/instance(%s)", f.workflowID, f.taskID, f.instanceID))
	e.SetTime(time.Now())

	if err := e.SetData(mime.ApplicationJSON, me); err != nil {
		l.Error(err, "failed to encode data for event report")
		return
	}

	if err := e.Validate(); err != nil {
		panic(fmt.Sprintf("Event creation results in invalid CloudEvent; fix implementation! error: %s", err))
	}

	// we start a new context because rootCtx might have been canceled
	// TODO: make configurable
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := f.reportClient.Send(ctx, e); err != nil {
		l.Error(err, "failed to report event")
		return
	}
}

// BuildTask from mmsys-test-scene-detection function.
func BuildTask(ctx context.Context, t *nbmpv2.Task) (nbmp.Function, error) {
	var (
		ok  bool
		err error
		f   = &function{}
	)

	// meta

	f.task = t

	f.workflowID, ok = nbmputils.GetStringParameterValue(t.Configuration.Parameters, nbmp.EngineWorkflowIDParameterKey)
	if !ok {
		return nil, fmt.Errorf("missing %s config", nbmp.EngineWorkflowIDParameterKey)
	}

	f.taskID, ok = nbmputils.GetStringParameterValue(t.Configuration.Parameters, nbmp.EngineTaskIDParameterKey)
	if !ok {
		return nil, fmt.Errorf("missing %s config", nbmp.EngineTaskIDParameterKey)
	}

	f.instanceID = t.General.ID

	// config

	f.test, ok = nbmputils.GetStringParameterValue(t.Configuration.Parameters, TestParameterKey)
	if !ok {
		return nil, fmt.Errorf("missing %s config", TestParameterKey)
	}

	f.maxNumberOfSimulatedCrashes = DefaultMaxNumberOfSimulatedCrashes
	maxNumberOfSimulatedCrashesStr, ok := nbmputils.GetStringParameterValue(t.Configuration.Parameters, MaxNumberOfSimulatedCrashesParameterKey)
	if ok {
		f.maxNumberOfSimulatedCrashes, err = strconv.Atoi(maxNumberOfSimulatedCrashesStr)
		if err != nil {
			return nil, fmt.Errorf("%s is not an integer", MaxNumberOfSimulatedCrashesParameterKey)
		}
	}

	f.simulatedCrashWaitDuration = DefaultSimulatedCrashWaitDuration
	simulatedCrashWaitDurationStr, ok := nbmputils.GetStringParameterValue(t.Configuration.Parameters, SimulatedCrashWaitDurationParameterKey)
	if ok {
		f.simulatedCrashWaitDuration, err = time.ParseDuration(simulatedCrashWaitDurationStr)
		if err != nil {
			return nil, fmt.Errorf("%s is not an integer", SimulatedCrashWaitDurationParameterKey)
		}
	}

	// input

	if len(t.Input.MediaParameters) == 0 {
		return nil, errors.New("no media input")
	}

	// TODO: use port bindings to determine input
	in := t.Input.MediaParameters[0]

	f.inVideoURL = string(in.CachingServerURL)
	if in.CachingServerURL == "" {
		return nil, errors.New("missing caching-server-url for media input")
	}

	// output

	if len(t.Output.MediaParameters) == 0 {
		return nil, errors.New("no media output")
	}

	// TODO: use port bindings to determine output
	out := t.Output.MediaParameters[0]

	if out.CachingServerURL == "" {
		return nil, errors.New("missing caching-server-url for media output")
	}

	url, err := out.CachingServerURL.URL()
	if err != nil {
		return nil, err
	}

	if url.Scheme != "s3" {
		return nil, errors.New("output is not in an S3 bucket")
	}

	seg := strings.SplitN(url.Path[1:], "/", 2)
	if len(seg) != 2 {
		return nil, errors.New("output is missing object name")
	}

	f.s3Endpoint = url.Host
	f.s3Bucket = seg[0]
	f.s3ObjectKey = seg[1]

	q, ok := url.Query()["AWS_ACCESS_KEY_ID"]
	if !ok {
		return nil, errors.New("output is missing AWS_ACCESS_KEY_ID")
	}
	f.awsAccessID = q[0]

	q, ok = url.Query()["AWS_SECRET_ACCESS_KEY"]
	if !ok {
		return nil, errors.New("output is missing AWS_SECRET_ACCESS_KEY")
	}
	f.awsSecretAccessKey = q[0]

	f.s3Insecure = true
	q, ok = url.Query()["INSECURE"]
	if ok {
		f.s3Insecure, _ = strconv.ParseBool(q[0])
	}

	// initialize NBMP reporting
	// TODO: DRY
	if t.Reporting == nil {
		return nil, errors.New("reporting description missing")
	}
	if t.Reporting.ReportType != nbmp.ReportTypeEngineCloudEvents {
		return nil, fmt.Errorf("unsupported report type '%s'", t.Reporting.ReportType)
	}
	if t.Reporting.DeliveryMethod != nbmpv2.HTTP_POSTDeliveryMethod {
		return nil, fmt.Errorf("unsupported delivery method '%s'", t.Reporting.ReportType)
	}
	c := &events.HTTPClient{
		URL:    string(t.Reporting.URL),
		Client: http.DefaultClient,
	}
	expBackOff := &backoff.ExponentialBackOff{
		InitialInterval:     500 * time.Millisecond,
		RandomizationFactor: 0.25,
		Multiplier:          1.5,
		MaxInterval:         2 * time.Second,
		MaxElapsedTime:      0, // = indefinitely (we use contexts for that)
		Stop:                backoff.Stop,
		Clock:               backoff.SystemClock,
	}
	f.reportClient = events.ClientWithBackoff(c, expBackOff)

	// initialize S3
	f.s3, err = minio.New(f.s3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(f.awsAccessID, f.awsSecretAccessKey, ""),
		Secure: !f.s3Insecure,
	})
	if err != nil {
		return nil, err
	}

	return f, nil
}

func init() {
	functions.TaskBuilders.Register(Name, BuildTask)
}
