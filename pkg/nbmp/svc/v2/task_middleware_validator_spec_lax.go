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

	nbmputils "github.com/nagare-media/engine/pkg/nbmp/utils"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

type taskValidatorSpecLaxMiddleware struct {
	next TaskService
}

// TaskValidatorSpecLaxMiddleware validates given task according the NBMP specification allowing for some deviation.
func TaskValidatorSpecLaxMiddleware(next TaskService) TaskService {
	return &taskValidatorSpecLaxMiddleware{
		next: next,
	}
}

var _ TaskServiceMiddleware = TaskValidatorSpecLaxMiddleware
var _ TaskService = &taskValidatorSpecLaxMiddleware{}

func (m *taskValidatorSpecLaxMiddleware) Create(ctx context.Context, t *nbmpv2.Task) error {
	m.common(ctx, t)

	if err := nbmputils.AcknowledgeStatusToErr(nbmputils.UpdateAcknowledgeStatus(t.Acknowledge)); err != nil {
		return err
	}

	return m.next.Create(ctx, t)
}

func (m *taskValidatorSpecLaxMiddleware) Update(ctx context.Context, t *nbmpv2.Task) error {
	m.common(ctx, t)

	// task must have an ID
	if t.General.ID == "" {
		t.Acknowledge.Failed = append(t.Acknowledge.Failed, "$.general.id")
	}

	if err := nbmputils.AcknowledgeStatusToErr(nbmputils.UpdateAcknowledgeStatus(t.Acknowledge)); err != nil {
		return err
	}

	return m.next.Update(ctx, t)
}

func (m *taskValidatorSpecLaxMiddleware) Delete(ctx context.Context, t *nbmpv2.Task) error {
	// no call to common as we only care about the workflow ID

	// task must have an ID
	if t.General.ID == "" {
		t.Acknowledge.Failed = append(t.Acknowledge.Failed, "$.general.id")
	}

	if err := nbmputils.AcknowledgeStatusToErr(nbmputils.UpdateAcknowledgeStatus(t.Acknowledge)); err != nil {
		return err
	}

	return m.next.Delete(ctx, t)
}

func (m *taskValidatorSpecLaxMiddleware) Retrieve(ctx context.Context, t *nbmpv2.Task) error {
	// no call to common as we only care about the workflow ID

	// task must have an ID
	if t.General.ID == "" {
		t.Acknowledge.Failed = append(t.Acknowledge.Failed, "$.general.id")
	}

	if err := nbmputils.AcknowledgeStatusToErr(nbmputils.UpdateAcknowledgeStatus(t.Acknowledge)); err != nil {
		return err
	}

	return m.next.Retrieve(ctx, t)
}

func (m *taskValidatorSpecLaxMiddleware) common(ctx context.Context, t *nbmpv2.Task) {
	// TODO: implement task validation
}
