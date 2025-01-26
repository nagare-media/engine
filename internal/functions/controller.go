/*
Copyright 2022-2025 The nagare media authors

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
	"encoding/json"
	"fmt"
	"os"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/nagare-media/engine/pkg/starter"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

// TaskController for task execution.
type TaskController struct {
	FunctionName string
	TDDPath      string
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

	// TODO: react to SIGHUP and reconfigure task
	tdd, err := decodeTddFile(c.TDDPath)
	if err != nil {
		l.Error(err, "unable to parse the task description document")
		return err
	}

	// build task
	l.Info("building task")
	tsk, err := builder(ctx, tdd)
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

func decodeTddFile(path string) (*nbmpv2.Task, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()

	tdd := &nbmpv2.Task{}
	err = dec.Decode(tdd)
	if err != nil {
		return nil, err
	}

	return tdd, nil
}
