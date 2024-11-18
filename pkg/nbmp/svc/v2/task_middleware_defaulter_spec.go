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
	"k8s.io/utils/ptr"

	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

type taskDefaulterSpecMiddleware struct {
	next TaskService
}

// TaskDefaulterSpecMiddleware sets default values for given task.
func TaskDefaulterSpecMiddleware(next TaskService) TaskService {
	return &taskDefaulterSpecMiddleware{
		next: next,
	}
}

var _ TaskServiceMiddleware = TaskDefaulterSpecMiddleware
var _ TaskService = &taskDefaulterSpecMiddleware{}

func (m *taskDefaulterSpecMiddleware) Create(ctx context.Context, t *nbmpv2.Task) error {
	if err := DefaultTask(t); err != nil {
		return err
	}

	if t.General.ID == "" {
		t.General.ID = uuid.UUIDv4()
	}
	t.General.State = ptr.To(nbmpv2.InstantiatedState)

	return m.next.Create(ctx, t)
}

func (m *taskDefaulterSpecMiddleware) Update(ctx context.Context, t *nbmpv2.Task) error {
	if err := DefaultTask(t); err != nil {
		return err
	}
	return m.next.Update(ctx, t)
}

func (m *taskDefaulterSpecMiddleware) Delete(ctx context.Context, t *nbmpv2.Task) error {
	if err := DefaultTask(t); err != nil {
		return err
	}
	return m.next.Delete(ctx, t)
}

func (m *taskDefaulterSpecMiddleware) Retrieve(ctx context.Context, t *nbmpv2.Task) error {
	if err := DefaultTask(t); err != nil {
		return err
	}
	return m.next.Retrieve(ctx, t)
}
