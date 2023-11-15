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

package svc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"text/template"

	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/internal/task-shim/actions"
	metaaction "github.com/nagare-media/engine/internal/task-shim/actions/meta"
	"github.com/nagare-media/engine/pkg/nbmp"
	nbmpsvcv2 "github.com/nagare-media/engine/pkg/nbmp/svc/v2"
	"github.com/nagare-media/engine/pkg/nbmp/utils"
	"github.com/nagare-media/engine/pkg/strobj"
	"github.com/nagare-media/engine/pkg/tplfuncs"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

type taskService struct {
	cfg *enginev1.TaskShimTaskServiceConfiguration

	rootCtx       context.Context
	terminateFunc func(error)

	mtx          sync.RWMutex
	tsk          *nbmpv2.Task       // protected by mtx
	tskErr       error              // protected by mtx
	tskDone      chan struct{}      // protected by mtx
	tskCtx       context.Context    // protected by mtx
	tskCtxCancel context.CancelFunc // protected by mtx
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
		l := log.FromContext(s.rootCtx)

		// wait for cancellation
		<-s.rootCtx.Done()

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

	// run onCreate actions
	ctx = log.IntoContext(ctx, l.WithValues("reason", "create"))
	return s.execEventActions(ctx, s.cfg.OnCreateActions, t)
}

func (s *taskService) Update(ctx context.Context, t *nbmpv2.Task) error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	// update is only allowed if Task was created, IDs match...
	if s.tsk == nil || s.tsk.General.ID != t.General.ID {
		return nbmp.ErrNotFound
	}
	// ...and it has not terminated
	if s.hasTerminated() {
		return nbmp.ErrInvalid
	}

	l := log.FromContext(s.rootCtx)
	l.Info("update task")

	// run onUpdate actions
	ctx = log.IntoContext(ctx, l.WithValues("reason", "update"))
	return s.execEventActions(ctx, s.cfg.OnUpdateActions, t)
}

func (s *taskService) Delete(ctx context.Context, t *nbmpv2.Task) error {
	l := log.FromContext(s.rootCtx)
	l.Info("delete task")

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
	if s.tsk == nil || s.tsk.General.ID != t.General.ID {
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
				s.startTask(log.IntoContext(ctx, l), currentTask)
				continue

			case metaaction.ConfigTypeRestartTask:
				s.restartTask(log.IntoContext(ctx, l), currentTask)
				continue

			case metaaction.ConfigTypeStopTask:
				s.stopTask(log.IntoContext(ctx, l))
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

func (s *taskService) startTask(ctx context.Context, currentTask *nbmpv2.Task) {
	// assumption: s.mtx is already locked by caller
	l := log.FromContext(s.rootCtx)

	// create new task context
	s.tsk = currentTask
	s.tskDone = make(chan struct{})
	s.tskCtx, s.tskCtxCancel = context.WithCancel(s.rootCtx)

	// start task
	l.Info("starting task")
	s.tsk.General.State = &nbmpv2.RunningState
	go func() {
		s.tskErr = s.execTask(s.tskCtx, currentTask)

		s.mtx.Lock()
		if s.tskErr != nil {
			l.Error(s.tskErr, "task failed")
			s.tsk.General.State = &nbmpv2.InErrorState
		} else {
			s.tsk.General.State = &nbmpv2.DestroyedState
		}
		l.Info("task finished")
		s.mtx.Unlock()

		close(s.tskDone)
	}()
}

func (s *taskService) restartTask(ctx context.Context, currentTask *nbmpv2.Task) {
	// assumption: s.mtx is already locked by caller
	log.FromContext(s.rootCtx)
	s.stopTask(ctx)
	s.startTask(ctx, currentTask)
}

func (s *taskService) stopTask(ctx context.Context) {
	// assumption: s.mtx is already locked by caller
	l := log.FromContext(s.rootCtx)
	l.Info("stopping task")

	// tskDone is nil => task has not started yet
	if s.tskDone == nil {
		l.Info("task was already stopped")
		return
	}

	// cancel task context and wait for termination
	s.tskCtxCancel()
	l.Info("waiting for task termination")
	<-s.tskDone
	l.Info("task stopped")
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
		utils.MustDeepCopyInto(currentTask, tskCopy)
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
