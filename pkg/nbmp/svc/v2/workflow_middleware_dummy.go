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

package v2

import (
	"context"

	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

type WorkflowDummyMiddleware struct {
	Next WorkflowService
}

var _ WorkflowService = &WorkflowDummyMiddleware{}

func (m *WorkflowDummyMiddleware) Create(ctx context.Context, w *nbmpv2.Workflow) error {
	return m.Next.Create(ctx, w)
}

func (m *WorkflowDummyMiddleware) Update(ctx context.Context, w *nbmpv2.Workflow) error {
	return m.Next.Update(ctx, w)
}

func (m *WorkflowDummyMiddleware) Delete(ctx context.Context, w *nbmpv2.Workflow) error {
	return m.Next.Delete(ctx, w)
}

func (m *WorkflowDummyMiddleware) Retrieve(ctx context.Context, w *nbmpv2.Workflow) error {
	return m.Next.Retrieve(ctx, w)
}
