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

package function

import (
	"context"
	goio "io"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/nagare-media/engine/internal/functions"
	"github.com/nagare-media/engine/internal/functions/io"
	"github.com/nagare-media/engine/pkg/nbmp"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

// Function description
const (
	Name = "data-discard"
)

// function discards input streams.
type function struct {
	tsk *nbmpv2.Task
}

var _ nbmp.Function = &function{}

// Exec data-discard function.
func (f *function) Exec(ctx context.Context) error {
	l := log.FromContext(ctx).WithName(Name)
	ctx = log.IntoContext(ctx, l)

	mgr := io.NewManager()

	//      ┌────function─────┐
	//    ┌─┴─┐   ┌─────────┐ │
	// ──▶│ * │──▶│ discard │ │
	//    └─┬─┘   └─────────┘ │
	//      └─────────────────┘

	// input ports
	for _, p := range f.tsk.General.InputPorts {
		port, err := io.NewInputPortForInputs(p, &f.tsk.Input)
		if err != nil {
			return err
		}

		discard := io.NewConnection(goio.Discard)
		port.Connect(discard)

		if err := mgr.ManagePort(port, true); err != nil {
			return err
		}
	}

	return mgr.Start(ctx)
}

// BuildTask from data-discard function.
func BuildTask(ctx context.Context, t *nbmpv2.Task) (nbmp.Function, error) {
	f := &function{
		tsk: t,
	}
	return f, nil
}

func init() {
	functions.TaskBuilders.Register(Name, BuildTask)
}
