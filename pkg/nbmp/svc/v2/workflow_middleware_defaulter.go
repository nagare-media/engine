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

package v2

import (
	"context"

	"github.com/nagare-media/engine/internal/pkg/uuid"

	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

type workflowDefaulterMiddleware struct {
	next WorkflowService
}

// WorkflowDefaulterMiddleware sets default values for given workflow.
func WorkflowDefaulterMiddleware(next WorkflowService) WorkflowService {
	return &workflowDefaulterMiddleware{
		next: next,
	}
}

var _ WorkflowServiceMiddleware = WorkflowDefaulterMiddleware
var _ WorkflowService = &workflowDefaulterMiddleware{}

func (m *workflowDefaulterMiddleware) Create(ctx context.Context, w *nbmpv2.Workflow) error {
	if err := m.common(ctx, w); err != nil {
		return err
	}

	// General

	if w.General.ID == "" {
		w.General.ID = uuid.UUIDv4()
	}

	w.General.State = &nbmpv2.InstantiatedState

	return m.next.Create(ctx, w)
}

func (m *workflowDefaulterMiddleware) Update(ctx context.Context, w *nbmpv2.Workflow) error {
	if err := m.common(ctx, w); err != nil {
		return err
	}
	return m.next.Update(ctx, w)
}

func (m *workflowDefaulterMiddleware) Delete(ctx context.Context, w *nbmpv2.Workflow) error {
	if err := m.common(ctx, w); err != nil {
		return err
	}
	return m.next.Delete(ctx, w)
}

func (m *workflowDefaulterMiddleware) Retrieve(ctx context.Context, w *nbmpv2.Workflow) error {
	if err := m.common(ctx, w); err != nil {
		return err
	}
	return m.next.Retrieve(ctx, w)
}

func (m *workflowDefaulterMiddleware) common(ctx context.Context, w *nbmpv2.Workflow) error {
	// Scheme

	if w.Scheme == nil {
		w.Scheme = &nbmpv2.Scheme{
			URI: nbmpv2.SchemaURI,
		}
	}

	// Acknowledge

	// reset Acknowledge for response
	w.Acknowledge = &nbmpv2.Acknowledge{}
	if w.Acknowledge.Unsupported == nil {
		w.Acknowledge.Unsupported = make([]string, 0)
	}
	if w.Acknowledge.Failed == nil {
		w.Acknowledge.Failed = make([]string, 0)
	}
	if w.Acknowledge.Partial == nil {
		w.Acknowledge.Partial = make([]string, 0)
	}

	return nil
}
