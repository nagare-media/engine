/*
Copyright 2022-2023 The nagare media authors

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
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log"

	backoff "github.com/cenkalti/backoff/v4"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/nagare-media/engine/internal/functions"
	"github.com/nagare-media/engine/internal/pkg/mime"
	"github.com/nagare-media/engine/internal/pkg/uuid"
	"github.com/nagare-media/engine/pkg/events"
	"github.com/nagare-media/engine/pkg/nbmp"
	nbmputils "github.com/nagare-media/engine/pkg/nbmp/utils"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

// Function description
const (
	Name = "mmsys-test-encode"

	TestParameterKey         = "mmsys-test-encode.engine.nagare.media/test"
	ChunkSecondsParameterKey = "mmsys-test-encode.engine.nagare.media/chunk-seconds"
)

const (
	// * baseline 1: simple encoding
	// * baseline 2: split+merge encoding
	// * baseline 3: split+merge encoding distributed
	//
	// with two disruptions:
	//   no event sourcing:
	//     * test 1: simple encoding
	//     * test 2: split+merge encoding
	//     * test 3: split+merge encoding distributed
	//   event sourcing:
	//     * test 4: split+merge encoding
	//     * test 5: split+merge encoding distributed

	BaselineSimple                = "baseline-simple"
	BaselineSplitMerge            = "baseline-split-merge"
	BaselineSplitMergeDistributed = "baseline-split-merge-distributed"

	TestNoRecoverySimple                = "test-no-recovery-simple"
	TestNoRecoverySplitMerge            = "test-no-recovery-split-merge"
	TestNoRecoverySplitMergeDistributed = "test-no-recovery-split-merge-distributed"

	TestRecoverySplitMerge            = "test-recovery-split-merge"
	TestRecoverySplitMergeDistributed = "test-recovery-split-merge-distributed"
)

const (
	MediaEncodedEventType  = "media.nagare.engine.v1alpha1.functions.mmsys-test-encode.media-encoded"
	MediaPackagedEventType = "media.nagare.engine.v1alpha1.functions.mmsys-test-encode.media-packaged"
)

type MediaEventData struct {
	Name string
	URL  string
}

// function mmsys-test-encode implements various video encoding test for the evaluation portion of an MMSys paper.
type function struct {
	// meta
	task       *nbmpv2.Task
	workflowID string
	taskID     string
	instanceID string

	// config
	test         string
	chunkSeconds int

	// input
	inVideoURL string
	inVideoFPS int
	inVideoDur time.Duration

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

// Exec mmsys-test-encode function.
func (f *function) Exec(ctx context.Context) error {
	l := log.FromContext(ctx, "test", f.test).WithName(Name)
	ctx = log.IntoContext(ctx, l)

	l.Info(fmt.Sprintf("executing %s test", f.test))

	switch f.test {
	case BaselineSimple:
		return f.baselineSimple(ctx)

	case BaselineSplitMerge:
		return f.baselineSplitMerge(ctx)

	case BaselineSplitMergeDistributed:
		return f.baselineSplitMergeDistributed(ctx)

	case TestNoRecoverySimple:
		return f.testNoRecoverySimple(ctx)

	case TestNoRecoverySplitMerge:
		return f.testNoRecoverySplitMerge(ctx)

	case TestNoRecoverySplitMergeDistributed:
		return f.testNoRecoverySplitMergeDistributed(ctx)

	case TestRecoverySplitMerge:
		return f.TestRecoverySplitMerge(ctx)

	case TestRecoverySplitMergeDistributed:
		return f.TestRecoverySplitMergeDistributed(ctx)

	default:
		return fmt.Errorf("unknown test '%s'", f.test)
	}
}

func (f *function) baselineSimple(ctx context.Context) error {
	l := log.FromContext(ctx)

	// encode
	encObjKey := path.Join("tmp", f.workflowID, f.taskID, f.instanceID, "out.hevc")
	encObjURL, err := f.encode(ctx, false, 0, f.inVideoURL, encObjKey)
	if err != nil {
		l.Error(err, "encoding failed")
		return err
	}

	// package to CMAF
	_, err = f.packageToCMAF(ctx, []string{encObjURL}, f.s3ObjectKey)
	if err != nil {
		l.Error(err, "packaging failed")
		return err
	}

	return nil
}

func (f *function) baselineSplitMerge(ctx context.Context) error {
	l := log.FromContext(ctx)

	var (
		chunkObjURLs        = make([]string, 0)
		chunkObjKeyPrefix   = path.Join("tmp", f.workflowID, f.taskID, f.instanceID)
		chunkObjNamePattern = "out%d-%d.hevc"
	)

	// encode
	inVideoSeconds := int(f.inVideoDur / time.Second)
	for seekSecond := 0; seekSecond < inVideoSeconds; seekSecond += f.chunkSeconds {
		chunkObjName := fmt.Sprintf(chunkObjNamePattern, seekSecond, seekSecond+f.chunkSeconds)
		chunkObjKey := path.Join(chunkObjKeyPrefix, chunkObjName)
		chunkObjURL, err := f.encode(ctx, true, seekSecond, f.inVideoURL, chunkObjKey)
		if err != nil {
			l.Error(err, "encoding failed")
			return err
		}
		chunkObjURLs = append(chunkObjURLs, chunkObjURL)
	}

	// package to CMAF
	_, err := f.packageToCMAF(ctx, chunkObjURLs, f.s3ObjectKey)
	if err != nil {
		l.Error(err, "packaging failed")
		return err
	}

	return nil
}

func (f *function) baselineSplitMergeDistributed(ctx context.Context) error {
	panic("TODO: implement")
}

func (f *function) testNoRecoverySimple(ctx context.Context) error {
	panic("TODO: implement")
}

func (f *function) testNoRecoverySplitMerge(ctx context.Context) error {
	panic("TODO: implement")
}

func (f *function) testNoRecoverySplitMergeDistributed(ctx context.Context) error {
	panic("TODO: implement")
}

func (f *function) TestRecoverySplitMerge(ctx context.Context) error {
	panic("TODO: implement")
}

func (f *function) TestRecoverySplitMergeDistributed(ctx context.Context) error {
	panic("TODO: implement")
}

func (f *function) encode(ctx context.Context, chunk bool, chunkSeekSecond int, inUrl, outObjKey string) (string, error) {
	l := log.FromContext(ctx)

	outObjURL := strings.Join([]string{f.s3.EndpointURL().String(), f.s3Bucket, outObjKey}, "/")
	if chunk {
		l.Info(fmt.Sprintf("encode chunk %s to %s", inUrl, outObjURL))
	} else {
		l.Info(fmt.Sprintf("encode video %s to %s", inUrl, outObjURL))
	}

	// report event if no error happened
	var err error
	defer func() {
		if err == nil {
			f.observeMediaEvent(ctx, MediaEncodedEventType, MediaEventData{
				Name: filepath.Base(outObjKey),
				URL:  outObjURL,
			})
		}
	}()

	var (
		gop                 = 2 * f.inVideoFPS
		chunkFrames         = f.chunkSeconds * f.inVideoFPS // assume constant framerate
		chunkOverlapFrames  = 2 * gop
		chunkOverlapSeconds = chunkOverlapFrames / f.inVideoFPS // = 4s
	)

	if chunk && chunkSeekSecond != 0 && chunkSeekSecond <= chunkOverlapSeconds {
		return "", fmt.Errorf("seek must be at least %ds long: %ds", chunkOverlapSeconds, chunkSeekSecond)
	}

	// FFmpeg process
	ffmpegArgs := []string{"-hide_banner", "-nostats"}
	if chunk {
		if chunkSeekSecond > 0 {
			ffmpegArgs = append(ffmpegArgs, "-accurate_seek", "-ss", strconv.Itoa(chunkSeekSecond-chunkOverlapSeconds))
		}
		ffmpegArgs = append(ffmpegArgs, "-t", strconv.Itoa(f.chunkSeconds+chunkOverlapSeconds))
	}
	ffmpegArgs = append(ffmpegArgs,
		"-i", inUrl,
		"-filter:v", "scale=1920:1920:force_original_aspect_ratio=decrease:force_divisible_by=2",
		"-f", "yuv4mpegpipe",
		"-",
	)

	ffmpegCmd := exec.CommandContext(ctx, "ffmpeg", ffmpegArgs...)
	ffmpegCmd.Stderr = os.Stdout // redirect all to stdout
	ffmpegPipe, err := ffmpegCmd.StdoutPipe()
	if err != nil {
		return "", err
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

	// x265 process
	x265Args := []string{
		"--no-progress",
		"--y4m", "-",
		"--profile", "main",
		"--preset", "medium",
		"--colorprim", "bt709",
		"--transfer", "bt709",
		"--colormatrix", "bt709",
		"--range", "limited",
		"--bframes", "3",
		"--ref", "3",
		"--hrd",
		"--crf", "26",
		"--vbv-bufsize", "4000",
		"--vbv-maxrate", "2000",
		"--no-open-gop",
		"--keyint", strconv.Itoa(gop),
		"--min-keyint", strconv.Itoa(gop),
	}
	if chunk {
		var startFrame, endFrame int
		if chunkSeekSecond == 0 {
			startFrame = 1
			endFrame = chunkFrames
		} else {
			startFrame = chunkOverlapFrames + 1
			endFrame = chunkFrames + chunkOverlapFrames
		}
		x265Args = append(x265Args,
			"--chunk-start", strconv.Itoa(startFrame),
			"--chunk-end", strconv.Itoa(endFrame),
		)
	}
	x265Args = append(x265Args, "-")

	x265Cmd := exec.CommandContext(ctx, "x265", x265Args...)
	x265Cmd.Stdin = ffmpegPipe
	x265Cmd.Stderr = os.Stdout // redirect all to stdout
	x265Pipe, err := x265Cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	x265Cmd.Cancel = func() error {
		// x265 apparently does not always terminate on SIGINT
		return x265Cmd.Process.Signal(os.Kill)
	}

	// start cmds
	l.V(1).Info(fmt.Sprintf("running command %s", ffmpegCmd))
	if err = ffmpegCmd.Start(); err != nil {
		l.Error(err, "FFmpeg failed")
		return "", err
	}
	l.V(1).Info(fmt.Sprintf("running command %s", x265Cmd))
	if err = x265Cmd.Start(); err != nil {
		l.Error(err, "x265 failed")
		return "", err
	}

	// upload to S3
	_, err = f.s3.PutObject(ctx, f.s3Bucket, outObjKey, x265Pipe, -1, minio.PutObjectOptions{ContentType: "video/mp4"})
	if err != nil {
		l.Error(err, "S3 upload failed")
		return "", err
	}

	// call Wait on cmds to clean up resources
	if err = x265Cmd.Wait(); err != nil {
		l.Error(err, "x265 failed")
		return "", err
	}
	if err = ffmpegCmd.Wait(); err != nil {
		l.Error(err, "FFmpeg failed")
		return "", err
	}

	return outObjURL, nil
}

func (f *function) packageToCMAF(ctx context.Context, inURLs []string, outObjKey string) (string, error) {
	if len(inURLs) == 0 {
		return "", errors.New("no inputs for CMAF packaging")
	}

	l := log.FromContext(ctx)

	outObjURL := strings.Join([]string{f.s3.EndpointURL().String(), f.s3Bucket, outObjKey}, "/")
	l.Info(fmt.Sprintf("package to %s", outObjURL))

	// report event if no error happened
	var err error
	defer func() {
		if err == nil {
			f.observeMediaEvent(ctx, MediaPackagedEventType, MediaEventData{
				Name: filepath.Base(outObjKey),
				URL:  outObjURL,
			})
		}
	}()

	// create tmp file
	tmpDir, err := os.MkdirTemp("", "")
	if err != nil {
		return "", err
	}
	tmpFile := path.Join(tmpDir, "out.mp4")

	// run MP4Box
	args := make([]string, 0, 2*len(inURLs)+5)
	args = append(args, "-add", inURLs[0])
	for _, in := range inURLs[1:] {
		args = append(args, "-cat", in)
	}
	args = append(args,
		"-frag", "1000",
		"--cmaf=cmf2",
		"-new", tmpFile,
	)

	mp4boxCmd := exec.CommandContext(ctx, "MP4Box", args...)
	mp4boxCmd.Stdout = os.Stdout
	mp4boxCmd.Stderr = os.Stdout // redirect all to stdout
	mp4boxCmd.Cancel = func() error {
		if runtime.GOOS == "windows" {
			return mp4boxCmd.Process.Signal(os.Kill)
		}
		go func() {
			time.Sleep(30 * time.Second)
			_ = mp4boxCmd.Process.Signal(os.Kill)
		}()
		return mp4boxCmd.Process.Signal(os.Interrupt)
	}

	l.V(1).Info(fmt.Sprintf("running command %s", mp4boxCmd))
	if err := mp4boxCmd.Run(); err != nil {
		l.Error(err, "MP4Box failed")
		return "", err
	}

	// upload to S3
	_, err = f.s3.FPutObject(ctx, f.s3Bucket, outObjKey, tmpFile, minio.PutObjectOptions{ContentType: "video/mp4"})
	if err != nil {
		l.Error(err, "S3 upload failed")
		return "", err
	}

	return outObjURL, nil
}

func (f *function) observeMediaEvent(ctx context.Context, t string, me MediaEventData) {
	l := log.FromContext(ctx)

	e := cloudevents.NewEvent()
	e.SetID(uuid.UUIDv4())
	e.SetType(t)
	e.SetSource("/engine.nagare.media/functions/mmsys-test-encode")
	e.SetSubject(fmt.Sprintf("/engine.nagare.media/workflow(%s)/task(%s)/instance(%s)", f.workflowID, f.taskID, f.instanceID))
	e.SetTime(time.Now())

	if err := e.SetData(mime.ApplicationJSON, me); err != nil {
		l.Error(err, "failed to encode data for event report")
		return
	}

	if err := e.Validate(); err != nil {
		panic(fmt.Sprintf("Event creation results in invalid CloudEvent; fix implementation! error: %s", err))
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := f.reportClient.Send(ctx, e); err != nil {
		l.Error(err, "failed to report event")
		return
	}
}

// BuildTask from mmsys-test-encode function.
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

	chunkSecondsStr, ok := nbmputils.GetStringParameterValue(t.Configuration.Parameters, ChunkSecondsParameterKey)
	if !ok {
		return nil, fmt.Errorf("missing %s config", ChunkSecondsParameterKey)
	}
	f.chunkSeconds, err = strconv.Atoi(chunkSecondsStr)
	if err != nil {
		return nil, errors.New("chunk-seconds is not an integer")
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

	fpsStr, ok := nbmputils.GetStringParameterValue(in.VideoFormat, nbmp.VideoFormatFrameRateAverage)
	if !ok {
		return nil, fmt.Errorf("missing video format %s parameter", nbmp.VideoFormatFrameRateAverage)
	}
	f.inVideoFPS, err = strconv.Atoi(fpsStr)
	if err != nil {
		return nil, errors.New("only integer frame rates are supported")
	}

	durationStr, ok := nbmputils.GetStringParameterValue(in.VideoFormat, nbmp.FormatFrameDuration)
	if !ok {
		return nil, fmt.Errorf("missing video format %s parameter", nbmp.FormatFrameDuration)
	}
	f.inVideoDur, err = time.ParseDuration(durationStr)
	if err != nil {
		return nil, fmt.Errorf("video format %s parameter is not a duration", nbmp.FormatFrameDuration)
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
