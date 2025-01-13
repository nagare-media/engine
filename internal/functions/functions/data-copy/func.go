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

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/nagare-media/engine/internal/functions"
	"github.com/nagare-media/engine/internal/functions/io"
	"github.com/nagare-media/engine/pkg/nbmp"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

// Function description
const (
	Name = "data-copy"

	MainInputPortName  = "in"
	MainOutputPortName = "out"
)

// function generates a test pattern input stream.
type function struct {
	tsk *nbmpv2.Task
}

var _ nbmp.Function = &function{}

// Exec data-copy function.
func (f *function) Exec(ctx context.Context) error {
	l := log.FromContext(ctx).WithName(Name)
	ctx = log.IntoContext(ctx, l)

	mgr := io.NewManager()

	//       ┌───────function────────┐
	//    ┌──┴─┐   ┌──────┐   ┌──────┴──┐
	// ──▶│ in │──▶│ main │──▶│ out.{n} │──▶
	//    └──┬─┘   └──────┘   └──────┬──┘
	//       └───────────────────────┘

	// streams
	mainStream := io.NewStream()

	// input ports
	if len(f.tsk.General.InputPorts) == 0 {
		return io.PortMissing
	}
	for _, p := range f.tsk.General.InputPorts {
		port, err := io.NewInputPortForInputs(p, &f.tsk.Input)
		if err != nil {
			return err
		}

		if port.Name() != MainInputPortName {
			return io.PortNameUnknown
		}

		port.Connect(mainStream)
		if err := mgr.ManagePort(port, true); err != nil {
			return err
		}

		break
	}

	// output ports
	if len(f.tsk.General.OutputPorts) == 0 {
		return io.PortMissing
	}
	for _, p := range f.tsk.General.OutputPorts {
		port, err := io.NewOutputPortForOutputs(p, &f.tsk.Output)
		if err != nil {
			return err
		}

		switch port.BaseName() {
		case MainOutputPortName:
			mainStream.Connect(port)
			if err := mgr.ManagePort(port, true); err != nil {
				return err
			}
		default:
			return io.PortNameUnknown
		}
	}

	// TODO: should we validate that everything is connected or is this simply a user error?

	return mgr.Start(ctx)
}

// BuildTask from data-copy function.
func BuildTask(ctx context.Context, t *nbmpv2.Task) (nbmp.Function, error) {
	f := &function{
		tsk: t,
	}
	return f, nil
}

func init() {
	functions.TaskBuilders.Register(Name, BuildTask)
}
