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

	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

type TaskDummyMiddleware struct {
	next TaskService
}

var _ TaskService = &TaskDummyMiddleware{}

func (m *TaskDummyMiddleware) Create(ctx context.Context, w *nbmpv2.Task) error {
	return m.next.Create(ctx, w)
}

func (m *TaskDummyMiddleware) Update(ctx context.Context, w *nbmpv2.Task) error {
	return m.next.Update(ctx, w)
}

func (m *TaskDummyMiddleware) Delete(ctx context.Context, w *nbmpv2.Task) error {
	return m.next.Delete(ctx, w)
}

func (m *TaskDummyMiddleware) Retrieve(ctx context.Context, w *nbmpv2.Task) error {
	return m.next.Retrieve(ctx, w)
}
