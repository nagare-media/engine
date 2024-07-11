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

type taskDefaulterMiddleware struct {
	next TaskService
}

// TaskDefaulterMiddleware sets default values for given task.
func TaskDefaulterMiddleware(next TaskService) TaskService {
	return &taskDefaulterMiddleware{
		next: next,
	}
}

var _ TaskServiceMiddleware = TaskDefaulterMiddleware
var _ TaskService = &taskDefaulterMiddleware{}

func (m *taskDefaulterMiddleware) Create(ctx context.Context, t *nbmpv2.Task) error {
	if err := m.common(ctx, t); err != nil {
		return err
	}

	//// General

	if t.General.ID == "" {
		t.General.ID = uuid.UUIDv4()
	}

	t.General.State = &nbmpv2.InstantiatedState

	return m.next.Create(ctx, t)
}

func (m *taskDefaulterMiddleware) Update(ctx context.Context, t *nbmpv2.Task) error {
	if err := m.common(ctx, t); err != nil {
		return err
	}
	return m.next.Update(ctx, t)
}

func (m *taskDefaulterMiddleware) Delete(ctx context.Context, t *nbmpv2.Task) error {
	if err := m.common(ctx, t); err != nil {
		return err
	}
	return m.next.Delete(ctx, t)
}

func (m *taskDefaulterMiddleware) Retrieve(ctx context.Context, t *nbmpv2.Task) error {
	if err := m.common(ctx, t); err != nil {
		return err
	}
	return m.next.Retrieve(ctx, t)
}

func (m *taskDefaulterMiddleware) common(ctx context.Context, t *nbmpv2.Task) error {
	//// Scheme

	if t.Scheme == nil {
		t.Scheme = &nbmpv2.Scheme{
			URI: nbmpv2.SchemaURI,
		}
	}

	//// Acknowledge

	// reset Acknowledge for response
	t.Acknowledge = &nbmpv2.Acknowledge{}
	if t.Acknowledge.Unsupported == nil {
		t.Acknowledge.Unsupported = make([]string, 0)
	}
	if t.Acknowledge.Failed == nil {
		t.Acknowledge.Failed = make([]string, 0)
	}
	if t.Acknowledge.Partial == nil {
		t.Acknowledge.Partial = make([]string, 0)
	}

	// TODO: implement task defaulter

	return nil
}
