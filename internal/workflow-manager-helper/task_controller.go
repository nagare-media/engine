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

package workflowmanagerhelper

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log"

	backoff "github.com/cenkalti/backoff/v4"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/pkg/starter"
	"github.com/nagare-media/models.go/base"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

const (
	DeleteTimeout = 30 * time.Second
)

type taskCtrl struct {
	cfg  *enginev1.WorkflowManagerHelperConfiguration
	data *enginev1.WorkflowManagerHelperData
	http http.Client
}

var _ starter.Starter = &taskCtrl{}

func NewTaskController(cfg *enginev1.WorkflowManagerHelperConfiguration, data *enginev1.WorkflowManagerHelperData) starter.Starter {
	return &taskCtrl{
		cfg:  cfg,
		data: data,
		http: http.Client{},
	}
}

func (c *taskCtrl) Start(ctx context.Context) error {
	l := log.FromContext(ctx).WithName("task")
	ctx = log.IntoContext(ctx, l)

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

	if err := c.createTaskPhase(ctx); err != nil {
		fmt.Fprint(terminationMsgBuf, "create phase failed: ")
		fmt.Fprintln(terminationMsgBuf, err)
		return err
	}

	// TODO: add Go routine that checks for updates to the secret resulting in PATCH requests to the Task API

	if err := c.observeTaskPhase(ctx); err != nil {
		// we ignore the error and always try to move on to the delete phase
		// we will record the error in the Kubernetes termination message
		fmt.Fprint(terminationMsgBuf, "observe phase failed: ")
		fmt.Fprintln(terminationMsgBuf, err)
	}

	if err := c.deleteTaskPhase(ctx); err != nil {
		// we ignore the error and always try to move on to the delete phase
		// we will record the error in the Kubernetes termination message
		fmt.Fprint(terminationMsgBuf, "delete phase failed: ")
		fmt.Fprintln(terminationMsgBuf, err)
	}

	return nil
}

func (c *taskCtrl) createTaskPhase(ctx context.Context) error {
	l := log.FromContext(ctx, "phase", "create-task")
	ctx = log.IntoContext(ctx, l)
	l.Info("starting new phase")

	// convert data to NBMP Task
	t := &nbmpv2.Task{}
	if err := c.data.ConvertToNBMPTask(t); err != nil {
		l.Error(err, "failed to convert passed nagare media engine data to NBMP Task")
		return err
	}
	t.Reporting = &nbmpv2.Reporting{
		ReportType:     "engine.nagare.media/cloud-events",
		URL:            base.URI(fmt.Sprintf("%s/events", *c.cfg.ReportsController.Webserver.PublicBaseURL)),
		DeliveryMethod: nbmpv2.HTTP_POSTDeliveryMethod,
	}

	buf := bytes.Buffer{}
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(t); err != nil {
		l.Error(err, "failed to encode NBMP Task as JSON")
		return err
	}

	op := func() error {
		l.Info("create task")

		ctx, cancle := context.WithTimeout(ctx, c.cfg.TaskController.RequestTimeout.Duration)
		defer cancle()

		req, err := http.NewRequestWithContext(ctx, "POST", c.cfg.TaskController.TaskAPI, &buf)
		if err != nil {
			return err
		}

		resp, err := c.http.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode == 204 || resp.StatusCode > 299 {
			// TODO: parse response body to give more infos in log
			return fmt.Errorf("unexpected HTTP status code in response: %d", resp.StatusCode)
		}

		return nil
	}

	no := func(err error, t time.Duration) {
		l.Error(err, fmt.Sprintf("failed; retrying after %s", t))
	}

	return backoff.RetryNotify(op, newBackOffWithContext(ctx), no)
}

func (c *taskCtrl) observeTaskPhase(ctx context.Context) error {
	l := log.FromContext(ctx, "phase", "observe-task")
	ctx = log.IntoContext(ctx, l)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	l.Info("starting new phase")

	var (
		observerErr, eventEmitterErr error
		observerDone                 = make(chan struct{})
		eventEmitterDone             = make(chan struct{})
	)
	go func() { observerErr = c.observerLoop(ctx); close(observerDone) }()
	go func() { eventEmitterErr = c.eventEmitterLoop(ctx); close(eventEmitterDone) }()

	select {
	case <-observerDone:
		cancel()
	case <-eventEmitterDone:
		cancel()
	}
	<-observerDone
	<-eventEmitterDone

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
				l.Error(err, "probe failed")
				failedProbes++
				continue
			}

			if tsk.General.State == nil {
				// this is unexpected and should be fixed by the Task API implementation
				l.Info("empty .general.state is assumed to be 'running'; fix Task API implementation")
				tsk.General.State = &nbmpv2.RunningState
			}

			switch *tsk.General.State {
			case nbmpv2.InErrorState:
				l.Info("task failed")
				return nil

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
	ctx, cancle := context.WithTimeout(ctx, c.cfg.TaskController.RequestTimeout.Duration)
	defer cancle()

	// TODO: allow ID to be set by Task API
	url := fmt.Sprintf("%s/%s", c.cfg.TaskController.TaskAPI, c.data.Task.ID)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 204 || resp.StatusCode > 299 {
		// TODO: parse response body to give more infos in log
		return nil, fmt.Errorf("unexpected HTTP status code in response: %d", resp.StatusCode)
	}

	t := &nbmpv2.Task{}
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(t)
	return t, err
}

func (c *taskCtrl) eventEmitterLoop(ctx context.Context) error {
	l := log.FromContext(ctx, "function", "event-emitter")
	ctx = log.IntoContext(ctx, l)
	l.Info("starting event-emitter")

	// TODO: check if there is an event input

	for {
		select {
		case <-ctx.Done():
			l.Info("termination requested")
			return nil

			// TODO: receive NATS events
		}
	}
}

func (c *taskCtrl) deleteTaskPhase(ctx context.Context) error {
	l := log.FromContext(ctx, "phase", "delete-task")
	// we start a new context because the original ctx might have been canceled
	ctx = log.IntoContext(context.Background(), l)
	l.Info("starting new phase")

	op := func() error {
		l.Info("delete task")

		ctx, cancle := context.WithTimeout(ctx, c.cfg.TaskController.RequestTimeout.Duration)
		defer cancle()

		// TODO: allow ID to be set by Task API
		url := fmt.Sprintf("%s/%s", c.cfg.TaskController.TaskAPI, c.data.Task.ID)
		req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
		if err != nil {
			return err
		}

		resp, err := c.http.Do(req)
		if err != nil {
			return err
		}

		if resp.StatusCode > 299 {
			// TODO: parse response body to give more infos in log
			return fmt.Errorf("unexpected HTTP status code in response: %d", resp.StatusCode)
		}

		return nil
	}

	no := func(err error, t time.Duration) {
		l.Error(err, fmt.Sprintf("failed; retrying after %s", t))
	}

	ctxDelete, cancle := context.WithTimeout(ctx, DeleteTimeout)
	defer cancle()
	return backoff.RetryNotify(op, newBackOffWithContext(ctxDelete), no)
}

func newBackOffWithContext(ctx context.Context) backoff.BackOff {
	expBackOff := &backoff.ExponentialBackOff{
		InitialInterval:     500 * time.Millisecond,
		RandomizationFactor: 0.25,
		Multiplier:          1.5,
		MaxInterval:         10 * time.Second,
		MaxElapsedTime:      0, // = indefinitely (we use contexts for that)
		Stop:                backoff.Stop,
		Clock:               backoff.SystemClock,
	}
	expBackOff.Reset()
	return backoff.WithContext(expBackOff, ctx)
}
