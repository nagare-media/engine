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

package svc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"text/template"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	backoff "github.com/cenkalti/backoff/v4"
	cloudevents "github.com/cloudevents/sdk-go/v2"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/internal/pkg/uuid"
	"github.com/nagare-media/engine/internal/task-shim/actions"
	metaaction "github.com/nagare-media/engine/internal/task-shim/actions/meta"
	"github.com/nagare-media/engine/pkg/events"
	"github.com/nagare-media/engine/pkg/nbmp"
	nbmpsvcv2 "github.com/nagare-media/engine/pkg/nbmp/svc/v2"
	nbmputils "github.com/nagare-media/engine/pkg/nbmp/utils"
	"github.com/nagare-media/engine/pkg/strobj"
	"github.com/nagare-media/engine/pkg/tplfuncs"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

const (
	// TODO: make configurable
	ReportTimeout = 10 * time.Second
)

type taskService struct {
	cfg *enginev1.TaskShimTaskServiceConfiguration

	rootCtx       context.Context
	terminateFunc func(error)

	mtx              sync.RWMutex
	engineWorkflowID string             // protected by mtx, but constant once Create was called
	engineTaskID     string             // protected by mtx, but constant once Create was called
	engineInstanceID string             // protected by mtx, but constant once Create was called
	tsk              *nbmpv2.Task       // protected by mtx
	tskErr           error              // protected by mtx
	tskDone          chan struct{}      // protected by mtx
	tskCtx           context.Context    // protected by mtx
	tskCtxCancel     context.CancelFunc // protected by mtx

	reportClient events.Client
}

var _ nbmpsvcv2.TaskService = &taskService{}

func NewTaskService(ctx context.Context, terminateFunc func(error), cfg *enginev1.TaskShimTaskServiceConfiguration) *taskService {
	s := &taskService{
		cfg:           cfg,
		rootCtx:       ctx,
		terminateFunc: terminateFunc,
	}
	s.init()
	return s
}

func (s *taskService) init() {
	// Go routine to handle root context cancellation
	go func() {
		// wait for cancellation
		<-s.rootCtx.Done()

		l := log.FromContext(s.rootCtx)

		// check if task is running and wait for it to terminate
		s.mtx.Lock()
		tskDone := s.tskDone
		s.mtx.Unlock()
		if tskDone != nil {
			<-tskDone
		}

		// terminate process
		l.Info("terminate process")
		s.terminateFunc(s.rootCtx.Err())
	}()
}

func (s *taskService) Create(ctx context.Context, t *nbmpv2.Task) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	// create is only allowed if Task does not exist
	if s.tsk != nil {
		return nbmp.ErrAlreadyExists
	}

	l := log.FromContext(s.rootCtx)
	l.Info("create task")

	s.engineInstanceID = t.General.ID
	if t.Configuration != nil {
		s.engineWorkflowID, _ = nbmputils.GetStringParameterValue(t.Configuration.Parameters, nbmp.EngineWorkflowIDParameterKey)
		s.engineTaskID, _ = nbmputils.GetStringParameterValue(t.Configuration.Parameters, nbmp.EngineTaskIDParameterKey)
	}
	l = l.WithValues(
		"workflow", s.engineWorkflowID,
		"task", s.engineTaskID,
		"instance", s.engineInstanceID,
	)
	s.rootCtx = log.IntoContext(s.rootCtx, l)

	// create reportClient if Task includes Reporting descriptors
	if err := s.createReportClient(t); err != nil {
		l.Error(err, "disable event reporting")
	}
	s.observeEvent(events.TaskCreated)

	// run onCreate actions
	ctx = log.IntoContext(ctx, l.WithValues("reason", "create"))
	return s.execEventActions(ctx, s.cfg.OnCreateActions, t)
}

func (s *taskService) createReportClient(t *nbmpv2.Task) error {
	// validate
	if t.Reporting == nil {
		return errors.New("not configured")
	}
	if t.Reporting.ReportType != nbmp.ReportTypeEngineCloudEvents {
		return fmt.Errorf("unsupported report type '%s'", t.Reporting.ReportType)
	}
	if t.Reporting.DeliveryMethod != nbmpv2.HTTP_POSTDeliveryMethod {
		return fmt.Errorf("unsupported delivery method '%s'", t.Reporting.ReportType)
	}

	// create report client
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
	s.reportClient = events.ClientWithBackoff(c, expBackOff)

	return nil
}

func (s *taskService) Update(ctx context.Context, t *nbmpv2.Task) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	// update is only allowed if Task was created, IDs match...
	if s.tsk == nil || s.engineInstanceID != t.General.ID {
		return nbmp.ErrNotFound
	}
	// ...and it has not terminated
	if s.hasTerminated() {
		return nbmp.ErrInvalid
	}

	l := log.FromContext(s.rootCtx)
	l.Info("update task")
	s.observeEvent(events.TaskUpdated)

	// run onUpdate actions
	ctx = log.IntoContext(ctx, l.WithValues("reason", "update"))
	return s.execEventActions(ctx, s.cfg.OnUpdateActions, t)
}

func (s *taskService) Delete(ctx context.Context, t *nbmpv2.Task) error {
	l := log.FromContext(s.rootCtx)
	l.Info("delete task")
	s.observeEvent(events.TaskDeleted)

	s.mtx.Lock()
	defer s.mtx.Unlock()

	// call to Delete should always result in termination of process
	defer l.Info("terminate process")
	defer s.terminateFunc(s.tskErr)

	// if Task has never been created => just exit
	if s.tsk == nil {
		l.Info("no task running")
		return nil
	}
	*t = *s.tsk

	// Task is running or has terminated => run onDelete actions
	ctx = log.IntoContext(ctx, l.WithValues("reason", "delete"))
	if err := s.execEventActions(ctx, s.cfg.OnDeleteActions, t); err != nil {
		return err
	}

	// if Task has terminated => just exit
	if s.hasTerminated() {
		l.Info("task has already terminated")
		return nil
	}

	// Task is running => stop it and wait for termination
	l.Info("terminate task")
	s.tskCtxCancel()
	<-s.tskDone
	return nil
}

func (s *taskService) Retrieve(ctx context.Context, t *nbmpv2.Task) error {
	l := log.FromContext(s.rootCtx)
	l.V(1).Info("retrieve task")

	if !s.mtx.TryRLock() {
		return nbmp.ErrRetryLater
	}
	defer s.mtx.RUnlock()

	// retrieve is only allowed if Task was created and IDs match
	if s.tsk == nil || s.engineInstanceID != t.General.ID {
		return nbmp.ErrNotFound
	}

	*t = *s.tsk
	return nil
}

func (s *taskService) execEventActions(ctx context.Context, al []enginev1.TaskServiceAction, currentTask *nbmpv2.Task) error {
	// assumption: s.mtx is already locked by caller
	l := log.FromContext(s.rootCtx)
	l.Info("execute event actions")

	for _, a := range al {
		l := l.WithValues(
			"name", a.Name,
			"action", a.Action,
		)

		// check for meta actions
		if a.Action == metaaction.Name {
			cfg, err := evaluateConfigToString(a, s.tsk, currentTask)
			if err != nil {
				return err
			}

			l := l.WithValues("config", cfg)
			l.Info("execute action")

			switch metaaction.ConfigType(cfg) {
			case metaaction.ConfigTypeStartTask:
				s.startTask(currentTask)
				continue

			case metaaction.ConfigTypeRestartTask:
				s.restartTask(currentTask)
				continue

			case metaaction.ConfigTypeStopTask:
				s.stopTask()
				continue
			}
		}

		// execute action
		l.Info("execute action")
		var err error
		currentTask, err = execAction(log.IntoContext(ctx, l), a, s.tsk, currentTask)
		if err != nil {
			l.Error(err, "action failed")
			return err
		}
	}

	l.Info("event actions finished")
	s.tsk = currentTask
	return nil
}

func (s *taskService) startTask(currentTask *nbmpv2.Task) {
	// assumption: s.mtx is already locked by caller
	l := log.FromContext(s.rootCtx)

	// create new task context
	s.tsk = currentTask
	s.tskDone = make(chan struct{})
	s.tskCtx, s.tskCtxCancel = context.WithCancel(s.rootCtx)

	// start task
	l.Info("starting task")
	s.tsk.General.State = &nbmpv2.RunningState
	s.observeEvent(events.TaskStarted)
	go func() {
		s.tskErr = s.execTask(s.tskCtx, currentTask)

		// Although we touch s.tsk.General.State, we don't lock s.mtx as this will result in a deadlock during parallel NBMP
		// service requests. A race is possible, but probably not critical as the task will either terminate or restart
		// which will set the correct s.tsk.General.State.
		if s.tskErr != nil {
			l.Error(s.tskErr, "task failed")
			s.tsk.General.State = &nbmpv2.InErrorState
			s.observeEvent(events.TaskFailed)
		} else {
			s.tsk.General.State = &nbmpv2.DestroyedState
			s.observeEvent(events.TaskSucceeded)
		}
		l.Info("task finished")

		close(s.tskDone)
	}()
}

func (s *taskService) restartTask(currentTask *nbmpv2.Task) {
	// assumption: s.mtx is already locked by caller
	l := log.FromContext(s.rootCtx)
	l.Info("restarting task")
	s.stopTask()
	s.startTask(currentTask)
}

func (s *taskService) stopTask() {
	// assumption: s.mtx is already locked by caller
	l := log.FromContext(s.rootCtx)
	l.Info("stopping task")

	// check if stop is required
	if s.tskDone == nil {
		l.Info("task has never started")
		return
	}
	if s.hasTerminated() {
		l.Info("task has already stopped")
		return
	}

	// cancel task context and wait for termination
	s.tskCtxCancel()
	l.Info("waiting for task termination")
	<-s.tskDone
	l.Info("task stopped")
	s.observeEvent(events.TaskStopped)
}

func (s *taskService) execTask(ctx context.Context, currentTask *nbmpv2.Task) error {
	// This is the main task execution function. It will run in its own Goroutine.

	l := log.FromContext(s.rootCtx)

	// remember pointer to last set task
	s.mtx.Lock()
	lastSetTsk := s.tsk
	s.mtx.Unlock()

	for _, a := range s.cfg.Actions {
		// set new task if ctx is not done
		select {
		case <-ctx.Done():
			// we should terminate
			l.Info("termination requested; terminate task")
			return nil

		default:
			s.mtx.Lock()
			if s.tsk == lastSetTsk {
				// there has been no update to the task during last action
				s.tsk = currentTask
				lastSetTsk = s.tsk
			} else {
				// task has been updated => ignore our currentTask and update pointers
				currentTask = s.tsk
				lastSetTsk = s.tsk
			}
			s.mtx.Unlock()
		}

		l := l.WithValues(
			"name", a.Name,
			"action", a.Action,
			"reason", "task",
		)

		// check for meta actions
		if a.Action == metaaction.Name {
			cfg, err := evaluateConfigToString(a, nil, currentTask)
			if err != nil {
				return err
			}

			l := l.WithValues("config", cfg)
			l.Info("execute action")

			switch metaaction.ConfigType(cfg) {
			case metaaction.ConfigTypeStartTask:
				return errors.New("svc: failed to start task: cannot start task from running task")
			case metaaction.ConfigTypeRestartTask:
				return errors.New("svc: failed to restart task: cannot restart task from running task")
			case metaaction.ConfigTypeStopTask:
				return errors.New("svc: failed to stop task: cannot stop task from running task")
			}
		}

		// execute action
		l.Info("execute action")
		var err error
		tskCopy := &nbmpv2.Task{}
		nbmputils.MustDeepCopyInto(currentTask, tskCopy)
		currentTask, err = execAction(log.IntoContext(ctx, l), a, nil, tskCopy)
		if err != nil {
			l.Error(err, "action failed")
			return err
		}
	}

	return nil
}

func (s *taskService) hasTerminated() bool {
	// assumption: s.mtx is already locked by caller

	// tskDone is nil => task has not started yet
	if s.tskDone == nil {
		return false
	}

	select {
	case <-s.tskDone:
		// channel is closed => task has terminated
		return true
	default:
		// channel is open => task is running
		return false
	}
}

func (s *taskService) observeEvent(t string) {
	if s.reportClient == nil {
		return
	}

	// TODO: implement event filtering

	l := log.FromContext(s.rootCtx)

	e := cloudevents.NewEvent()
	e.SetID(uuid.UUIDv4())
	e.SetType(t)
	e.SetSource("/engine.nagare.media/task-shim")
	e.SetSubject(fmt.Sprintf("/engine.nagare.media/workflow(%s)/task(%s)/instance(%s)", s.engineWorkflowID, s.engineTaskID, s.engineInstanceID))
	e.SetTime(time.Now())
	if err := e.Validate(); err != nil {
		panic(fmt.Sprintf("Event creation results in invalid CloudEvent; fix implementation! error: %s", err))
	}

	// we start a new context because rootCtx might have been canceled
	ctx, cancel := context.WithTimeout(context.Background(), ReportTimeout)
	defer cancel()

	if err := s.reportClient.Send(ctx, e); err != nil {
		l.Error(err, "failed to report event")
		return
	}
}

func execAction(ctx context.Context, a enginev1.TaskServiceAction, oldTask, currentTask *nbmpv2.Task) (*nbmpv2.Task, error) {
	// evaluate configuration
	cfg, err := evaluateConfigToRawJSON(a, oldTask, currentTask)
	if err != nil {
		return nil, err
	}

	// get action builder
	builder, ok := actions.ActionBuilders.Get(a.Action)
	if !ok {
		return nil, fmt.Errorf("action named %s not registered", a.Action)
	}

	// build action
	actionCtx := actions.Context{
		Ctx:    ctx,
		Config: cfg,
		ContextData: actions.ContextData{
			OldTask: oldTask,
			Task:    currentTask,
		},
	}
	action, err := builder(actionCtx)
	if err != nil {
		return nil, err
	}

	// execute action
	return action.Exec(actionCtx)
}

func evaluateConfigToString(a enginev1.TaskServiceAction, oldTask, currentTask *nbmpv2.Task) (string, error) {
	j, err := evaluateConfigToRawJSON(a, oldTask, currentTask)
	if err != nil {
		return "", err
	}
	var str string
	if err = json.Unmarshal(j, &str); err != nil {
		return "", err
	}
	return str, nil
}

func evaluateConfigToRawJSON(a enginev1.TaskServiceAction, oldTask, currentTask *nbmpv2.Task) ([]byte, error) {
	switch {
	case a.Config == nil:
		return []byte("null"), nil

	case a.Config.Type == strobj.Object:
		return a.Config.ObjectVal.Raw, nil

	case a.Config.Type == strobj.String:
		// evaluate string as a Go template...
		tpl, err := template.
			New("config").
			Funcs(tplfuncs.DefaultFuncMap()).
			Parse(a.Config.StrVal)
		if err != nil {
			return nil, err
		}

		buf := &bytes.Buffer{}
		tplData := actions.ContextData{
			OldTask: oldTask,
			Task:    currentTask,
		}
		if err = tpl.Execute(buf, tplData); err != nil {
			return nil, err
		}

		// ...then parse it as YAML and convert it to a JSON string...
		raw, err := yaml.YAMLToJSON(buf.Bytes())
		if err != nil {
			return nil, err
		}
		return raw, nil
	default:
		return nil, fmt.Errorf("svc: impossible StringOrObject.Type")
	}
}
