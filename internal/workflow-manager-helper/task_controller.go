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

package workflowmanagerhelper

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/log"

	backoff "github.com/cenkalti/backoff/v4"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/nats-io/nats.go/jetstream"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/internal/pkg/mime"
	enginenats "github.com/nagare-media/engine/internal/pkg/nats"
	"github.com/nagare-media/engine/pkg/events"
	"github.com/nagare-media/engine/pkg/nbmp"
	nbmpclientv2 "github.com/nagare-media/engine/pkg/nbmp/client/v2"
	"github.com/nagare-media/engine/pkg/starter"
	"github.com/nagare-media/engine/pkg/updatable"
	"github.com/nagare-media/models.go/base"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

const (
	// TODO: make configurable
	TaskCreateTimeout = 1 * time.Minute
	TaskUpdateTimeout = 1 * time.Minute
	TaskDeleteTimeout = 1 * time.Minute
)

var (
	probeLaterErr    = errors.New("probe later")
	taskRestartedErr = errors.New("task restarted")
	taskFailedErr    = errors.New("task failed")
)

type taskCtrl struct {
	cfg  *enginev1.WorkflowManagerHelperConfig
	data updatable.Updatable[*enginev1.WorkflowManagerHelperData]

	// NBMP Task ID returned by the Task API
	tskInstanceID string

	client nbmpclientv2.TaskClient
}

var _ starter.Starter = &taskCtrl{}

func NewTaskController(cfg *enginev1.WorkflowManagerHelperConfig, data updatable.Updatable[*enginev1.WorkflowManagerHelperData]) starter.Starter {
	return &taskCtrl{
		cfg:    cfg,
		data:   data,
		client: nbmpclientv2.NewTaskClient(cfg.TaskController.TaskAPI),
	}
}

func (c *taskCtrl) Start(ctx context.Context) error {
	l := log.FromContext(ctx).WithName("task")

	// Kubernetes reads final termination messages in /dev/termination-log
	terminationMsgBuf := &bytes.Buffer{}
	defer func() {
		if terminationMsgBuf.Len() == 0 {
			return
		}

		f, err := os.OpenFile("/dev/termination-log", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
		if err != nil {
			l.Error(err, "failed to open /dev/termination-log")
			return
		}
		defer f.Close()

		if _, err := io.Copy(f, terminationMsgBuf); err != nil {
			l.Error(err, "failed to write /dev/termination-log")
			return
		}
	}()

	// subscribe to data changes
	data, dataSub := c.data.Subscribe()
	l = l.WithValues(
		"workflow", data.Value.Workflow.ID,
		"task", data.Value.Task.ID,
	)
	ctx = log.IntoContext(ctx, l)

	// create phase
	if err := c.createTaskPhase(ctx, data.Value); err != nil {
		fmt.Fprint(terminationMsgBuf, "create phase failed: ")
		fmt.Fprintln(terminationMsgBuf, err)
		return err
	}

	// observe phase
	l = l.WithValues("instance", c.tskInstanceID)
	ctx = log.IntoContext(ctx, l)

	var observeTaskErr error
	if observeTaskErr = c.observeTaskPhase(ctx, data.Value, dataSub); observeTaskErr != nil {
		// we ignore the error and always try to move on to the delete phase
		// we will record the error in the Kubernetes termination message
		fmt.Fprint(terminationMsgBuf, "observe phase failed: ")
		fmt.Fprintln(terminationMsgBuf, observeTaskErr)

		// don't run delete task phase on restarted tasks
		if observeTaskErr == taskRestartedErr {
			return observeTaskErr
		}
	}

	// delete phase
	if err := c.deleteTaskPhase(ctx); err != nil {
		// we will record the error in the Kubernetes termination message
		fmt.Fprint(terminationMsgBuf, "delete phase failed: ")
		fmt.Fprintln(terminationMsgBuf, err)
		return err
	}

	return observeTaskErr
}

func (c *taskCtrl) createTaskPhase(ctx context.Context, data *enginev1.WorkflowManagerHelperData) error {
	l := log.FromContext(ctx, "phase", "create-task")
	ctx = log.IntoContext(ctx, l)
	l.Info("starting new phase")

	// convert data to NBMP Task
	t, err := c.convertToNBMPTask(data)
	if err != nil {
		l.Error(err, "failed to convert passed nagare media engine data to NBMP Task")
		return err
	}

	op := func() error {
		l.Info("create task")

		ctx, cancel := context.WithTimeout(ctx, c.cfg.TaskController.CreateRequestTimeout.Duration)
		defer cancel()

		t, err = c.client.Create(ctx, t)
		if err != nil {
			// TODO: check if t is not nil to give more infos in log
			l.Error(err, "failed to create NBMP Task")
			return err
		}

		c.tskInstanceID = t.General.ID

		return nil
	}

	no := func(err error, t time.Duration) {
		l.Error(err, fmt.Sprintf("failed; retrying after %s", t))
	}

	ctxCreate, cancel := context.WithTimeout(ctx, TaskCreateTimeout)
	defer cancel()

	err = backoff.RetryNotify(op, newBackOffWithContext(ctxCreate), no)
	if err != nil {
		l.Error(err, "failed to create task")
	}
	return err
}

func (c *taskCtrl) convertToNBMPTask(data *enginev1.WorkflowManagerHelperData) (*nbmpv2.Task, error) {
	t := &nbmpv2.Task{}
	if err := data.ConvertToNBMPTask(t); err != nil {
		return nil, err
	}
	t.Reporting = &nbmpv2.Reporting{
		ReportType:     nbmp.ReportTypeEngineCloudEvents,
		URL:            base.URI(fmt.Sprintf("%s/events", *c.cfg.ReportsController.Webserver.PublicBaseURL)),
		DeliveryMethod: nbmpv2.HTTP_POSTDeliveryMethod,
	}
	return t, nil
}

func (c *taskCtrl) observeTaskPhase(ctx context.Context, data *enginev1.WorkflowManagerHelperData,
	dataSub <-chan updatable.VersionedValue[*enginev1.WorkflowManagerHelperData]) error {
	l := log.FromContext(ctx, "phase", "observe-task")
	ctx = log.IntoContext(ctx, l)
	l.Info("starting new phase")

	var (
		tskUpdateErr, observerErr, eventEmitterErr error
		observerDone, eventEmitterDone             chan struct{}

		ctx2   context.Context
		cancel context.CancelFunc
	)

loop:
	for {
		// start observation
		ctx2, cancel = context.WithCancel(ctx)
		observerDone = make(chan struct{})
		eventEmitterDone = make(chan struct{})

		go func() { observerErr = c.observerLoop(ctx2); close(observerDone) }()
		go func() { eventEmitterErr = c.eventEmitterLoop(ctx2, data); close(eventEmitterDone) }()

		select {
		// data update
		case dv, ok := <-dataSub:
			if !ok {
				// unexpected channel close
				tskUpdateErr = errors.New("unexpected data subscription channel closed")
				break loop
			}

			l.Info("observed data change; update task")

			// stop observation
			cancel()
			<-observerDone
			<-eventEmitterDone

			// update task
			data = dv.Value
			if tskUpdateErr = c.updateTask(ctx, data); tskUpdateErr != nil {
				break loop
			}

		// unexpected termination
		case <-observerDone:
			break loop
		case <-eventEmitterDone:
			break loop

		// requested termination
		case <-ctx.Done():
			break loop
		}
	}

	cancel()
	<-observerDone
	<-eventEmitterDone

	if tskUpdateErr != nil {
		return tskUpdateErr
	}
	if observerErr != nil {
		return observerErr
	}
	return eventEmitterErr
}

func (c *taskCtrl) observerLoop(ctx context.Context) error {
	l := log.FromContext(ctx, "function", "observer")
	ctx = log.IntoContext(ctx, l)
	l.Info("starting observer")

	t := time.NewTimer(c.cfg.TaskController.ObservePeriode.Duration)
	defer t.Stop()

	failedProbes := 0
	for {
		if failedProbes > *c.cfg.TaskController.MaxFailedProbes {
			err := fmt.Errorf("probe failed %d times", failedProbes)
			l.Error(err, "observer failed")
			return err
		}

		// reset timer
		t.Stop()
		t.Reset(c.cfg.TaskController.ObservePeriode.Duration)

		select {
		case <-ctx.Done():
			l.Info("termination requested")
			return nil

		case <-t.C:
			tsk, err := c.probeTask(ctx)
			if err != nil {
				if err == probeLaterErr {
					// we should check later
					continue
				}
				if err == taskRestartedErr {
					l.Error(err, "probe failed likely because the task has been restarted; exit workflow-manager-helper")
					return err
				}
				l.Error(err, "probe failed")
				failedProbes++
				continue
			}

			if tsk.General.State == nil {
				// this is unexpected and should be fixed by the Task API implementation
				l.Info("empty .general.state is assumed to be 'running'; fix Task API implementation")
				tsk.General.State = ptr.To(nbmpv2.RunningState)
			}

			switch *tsk.General.State {
			case nbmpv2.InErrorState:
				l.Error(taskFailedErr, "task failed")
				return taskFailedErr

			case nbmpv2.DestroyedState:
				l.Info("task finished successfully")
				return nil

			case nbmpv2.InstantiatedState:
				fallthrough
			case nbmpv2.IdleState:
				fallthrough
			case nbmpv2.RunningState:
				l.V(1).Info("task still running")
				failedProbes = 0

			default:
				l.Error(fmt.Errorf("unknown .general.status '%s'", *tsk.General.State), "probe failed")
				failedProbes++
				continue
			}
		}
	}
}

func (c *taskCtrl) probeTask(ctx context.Context) (*nbmpv2.Task, error) {
	l := log.FromContext(ctx)
	l.V(1).Info("probe task")

	// send GET request to Task API
	ctx, cancel := context.WithTimeout(ctx, c.cfg.TaskController.RetrieveRequestTimeout.Duration)
	defer cancel()

	t, err := c.client.Retrieve(ctx, c.tskInstanceID)

	switch err {
	case nil:
		return t, nil

	case nbmp.ErrRetryLater:
		return nil, probeLaterErr

	case nbmp.ErrNotFound:
		// 404 error => probably a restart of the task => restart workflow-manager-helper
		return nil, taskRestartedErr

	default:
		// TODO: check if t is not nil to give more infos in log
		return nil, err
	}
}

func (c *taskCtrl) eventEmitterLoop(ctx context.Context, data *enginev1.WorkflowManagerHelperData) error {
	l := log.FromContext(ctx, "function", "event-emitter")
	ctx = log.IntoContext(ctx, l)
	l.Info("starting event-emitter")

	// check for task event inputs
	eventClients := make(map[string]events.Client)
	for _, in := range data.Task.Inputs {
		if in.Type != enginev1.MetadataMediaType {
			continue
		}
		if in.Metadata.MimeType != nil && *in.Metadata.MimeType != mime.ApplicationCloudEventsJSON {
			continue
		}
		if in.Direction != nil && *in.Direction != enginev1.PushMediaDirection {
			l.Info(fmt.Sprintf("skipping task event input '%s' with no push direction", in.ID))
			continue
		}
		if in.URL == nil {
			l.Info(fmt.Sprintf("skipping task event input '%s' with missing URL", in.ID))
			continue
		}

		// TODO: implement event filters
		ec := &events.HTTPClient{
			URL:    *in.URL,
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
		eventClients[in.ID] = events.ClientWithBackoff(ec, expBackOff)
	}

	if len(eventClients) == 0 {
		// even when we return early, we need to wait for ctx to be canceled
		l.Info("no task event inputs; terminating")
		<-ctx.Done()
		return nil
	}

	// connect to NATS
	nc, js, err := enginenats.CreateJetStreamConn(ctx, string(data.System.NATS.URL))
	if err != nil {
		return err
	}
	defer nc.Close()

	// ensure stream exists for messages to persist
	// TODO: add support for streams created by user
	s, err := enginenats.CreateOrUpdateEngineStream(ctx, js)
	if err != nil {
		return err
	}

	subjectName := enginenats.Subject(enginenats.SubjectPrefix, data.Workflow.ID, data.Task.ID)
	con, err := s.OrderedConsumer(ctx, jetstream.OrderedConsumerConfig{
		FilterSubjects: []string{subjectName},
		DeliverPolicy:  jetstream.DeliverAllPolicy,
		ReplayPolicy:   jetstream.ReplayInstantPolicy,
	})
	if err != nil {
		return err
	}

	cc, err := con.Consume(func(msg jetstream.Msg) {
		defer func() {
			if err := msg.Ack(); err != nil {
				if ctx.Err() == nil {
					// only log if ctx is still open
					l.Error(err, "failed to acknowledge consumption of NATS message")
				}
			}
		}()

		e := cloudevents.Event{}
		if err := json.Unmarshal(msg.Data(), &e); err != nil {
			l.Error(err, "failed to decode event as CloudEvent; skipping")
			return
		}

		for name, ec := range eventClients {
			// TODO: make configurable
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			if err := ec.SendAsyncAck(ctx, e); err != nil {
				// only log if ctx is still open
				l.Error(err, fmt.Sprintf("failed to send event to task input '%s'", name))
			}
			cancel()
		}
	})
	if err != nil {
		return err
	}

	<-ctx.Done()
	l.Info("termination requested")
	cc.Stop()
	return nil
}

func (c *taskCtrl) updateTask(ctx context.Context, data *enginev1.WorkflowManagerHelperData) error {
	l := log.FromContext(ctx)

	// convert data to NBMP Task
	t, err := c.convertToNBMPTask(data)
	if err != nil {
		l.Error(err, "failed to convert passed nagare media engine data to NBMP Task")
		return err
	}
	t.General.ID = c.tskInstanceID

	op := func() error {
		l.Info("update task")

		ctx, cancel := context.WithTimeout(ctx, c.cfg.TaskController.UpdateRequestTimeout.Duration)
		defer cancel()

		t, err = c.client.Update(ctx, t)
		if err != nil {
			// TODO: check if t is not nil to give more infos in log
			l.Error(err, "failed to update NBMP Task")
			return err
		}

		if c.tskInstanceID != t.General.ID {
			return fmt.Errorf("NBMP task ID '%s' doesn't match task instance ID '%s'", t.General.ID, c.tskInstanceID)
		}

		return nil
	}

	no := func(err error, t time.Duration) {
		l.Error(err, fmt.Sprintf("failed; retrying after %s", t))
	}

	ctxUpdate, cancel := context.WithTimeout(ctx, TaskUpdateTimeout)
	defer cancel()

	err = backoff.RetryNotify(op, newBackOffWithContext(ctxUpdate), no)
	if err != nil {
		l.Error(err, "failed to update task")
	}
	return err
}

func (c *taskCtrl) deleteTaskPhase(ctx context.Context) error {
	l := log.FromContext(ctx, "phase", "delete-task")
	// we start a new context because the original ctx might have been canceled
	ctx = log.IntoContext(context.Background(), l)
	l.Info("starting new phase")

	op := func() error {
		l.Info("delete task")

		ctx, cancel := context.WithTimeout(ctx, c.cfg.TaskController.DeleteRequestTimeout.Duration)
		defer cancel()

		_, err := c.client.Delete(ctx, c.tskInstanceID)
		if err != nil {
			l.Error(err, "failed to delete NBMP Task")
			return err
		}

		return nil
	}

	no := func(err error, t time.Duration) {
		l.Error(err, fmt.Sprintf("failed; retrying after %s", t))
	}

	ctxDelete, cancel := context.WithTimeout(ctx, TaskDeleteTimeout)
	defer cancel()

	err := backoff.RetryNotify(op, newBackOffWithContext(ctxDelete), no)
	if err != nil {
		l.Error(err, "failed to delete task")
	}
	return err
}
