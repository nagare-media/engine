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

package functions

import (
	"context"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/nagare-media/engine/pkg/starter"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

// TaskController for task execution.
type TaskController struct {
	FunctionName    string
	TaskDescription *nbmpv2.Task
}

var _ starter.Starter = &TaskController{}

// Start task controller.
func (c *TaskController) Start(ctx context.Context) error {
	l := log.FromContext(ctx)
	l.Info("starting task controller")

	// get task builder
	builder, ok := TaskBuilders.Get(c.FunctionName)
	if !ok {
		err := fmt.Errorf("function named %s not registered", c.FunctionName)
		l.Error(err, "finding task builder failed")
		return err
	}

	// build task
	l.Info("building task")
	tsk, err := builder(ctx, c.TaskDescription)
	if err != nil {
		l.Error(err, "building task failed")
		return err
	}

	// execute task
	l.Info("starting task")
	err = tsk.Exec(ctx)
	if err != nil {
		l.Error(err, "task failed")
	}
	l.Info("task finished")
	return err
}
