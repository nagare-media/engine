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

package v2

import (
	"context"

	"github.com/nagare-media/engine/internal/pkg/uuid"

	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

type workflowDefaulterSpecMiddleware struct {
	next WorkflowService
}

// WorkflowDefaulterSpecMiddleware sets default values for given workflow.
func WorkflowDefaulterSpecMiddleware(next WorkflowService) WorkflowService {
	return &workflowDefaulterSpecMiddleware{
		next: next,
	}
}

var _ WorkflowServiceMiddleware = WorkflowDefaulterSpecMiddleware
var _ WorkflowService = &workflowDefaulterSpecMiddleware{}

func (m *workflowDefaulterSpecMiddleware) Create(ctx context.Context, w *nbmpv2.Workflow) error {
	if err := m.common(w); err != nil {
		return err
	}

	if w.General.ID == "" {
		w.General.ID = uuid.UUIDv4()
	}
	w.General.State = &nbmpv2.InstantiatedState

	return m.next.Create(ctx, w)
}

func (m *workflowDefaulterSpecMiddleware) Update(ctx context.Context, w *nbmpv2.Workflow) error {
	if err := m.common(w); err != nil {
		return err
	}
	return m.next.Update(ctx, w)
}

func (m *workflowDefaulterSpecMiddleware) Delete(ctx context.Context, w *nbmpv2.Workflow) error {
	if err := m.common(w); err != nil {
		return err
	}
	return m.next.Delete(ctx, w)
}

func (m *workflowDefaulterSpecMiddleware) Retrieve(ctx context.Context, w *nbmpv2.Workflow) error {
	if err := m.common(w); err != nil {
		return err
	}
	return m.next.Retrieve(ctx, w)
}

func (m *workflowDefaulterSpecMiddleware) common(w *nbmpv2.Workflow) error {
	return DefaultWorkflow(w)
}
