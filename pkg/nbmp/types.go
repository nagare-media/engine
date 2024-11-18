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

package nbmp

import (
	"context"

	"github.com/nagare-media/models.go/base"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

const (
	// NBMP brand v1 for nagare media engine.
	BrandNagareMediaEngineV1 = base.URI("urn:nagare-media:engine:schema:nbmp:v1")
)

const (
	ReportTypeEngineCloudEvents = "engine.nagare.media/cloud-events"
)

const (
	EngineWorkflowIDParameterKey          = "engine.nagare.media/workflow-id"
	EngineWorkflowNameParameterKey        = "engine.nagare.media/workflow-name"
	EngineWorkflowDescriptionParameterKey = "engine.nagare.media/workflow-description"
	EngineTaskIDParameterKey              = "engine.nagare.media/task-id"
	EngineTaskNameParameterKey            = "engine.nagare.media/task-name"
	EngineTaskDescriptionParameterKey     = "engine.nagare.media/task-description"
)

const (
	TypicalDelayFlowcontrolQueryParameterKey    = "typical-delay"
	MinDelayFlowcontrolQueryParameterKey        = "min-delay"
	MaxDelayFlowcontrolQueryParameterKey        = "max-delay"
	MinThroughputFlowcontrolQueryParameterKey   = "min-throughput"
	MaxThroughputFlowcontrolQueryParameterKey   = "max-throughput"
	AveragingWindowFlowcontrolQueryParameterKey = "averaging-window"
)

// Function interface.
type Function interface {
	// Execute this function.
	Exec(context.Context) error
}

// FunctionFunc adapts Go functions to NBMP functions.
type FunctionFunc func(context.Context) error

// Exec adapts FunctionFunc to the Function's Exec method.
func (f FunctionFunc) Exec(ctx context.Context) error {
	return f(ctx)
}

// TaskBuilder configures a function based on the given NBMP task description.
type TaskBuilder func(context.Context, *nbmpv2.Task) (Function, error)
