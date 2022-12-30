/*
Copyright 2022 The nagare media authors

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

package svc

import (
	"context"

	"github.com/nagare-media/engine/internal/gateway-nbmp/store"
	"github.com/nagare-media/engine/internal/pkg/uuid"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type WorkflowService interface {
	Create(ctx context.Context, wf *nbmpv2.Workflow) error
	Update(ctx context.Context, wf *nbmpv2.Workflow) error
	Delete(ctx context.Context, wf *nbmpv2.Workflow) error
	Retrieve(ctx context.Context, wf *nbmpv2.Workflow) error
}

type workflowService struct{}

var _ WorkflowService = &workflowService{}

func NewWorkflowService() *workflowService {
	return nil
}

func NewWorkflowServiceWithStore(s store.Store) *workflowService {
	return nil
}

func (wfsvc *workflowService) Create(ctx context.Context, wf *nbmpv2.Workflow) error {
	l := log.FromContext(ctx)
	PopulateWorkflowDefaults(wf)

	// the ID should initially be empty
	if wf.General.ID != "" {
		wf.Acknowledge = &nbmpv2.Acknowledge{
			Status: nbmpv2.FailedAcknowledgeStatus,
			Failed: []string{"$.general.id"},
		}
		return ErrInvalid
	}
	wf.General.ID = uuid.UUIDv4()
	l = l.WithValues("workflowID", wf.General.ID)

	l.V(2).Info("creating workflow")

	return nil
}

func (wfsvc *workflowService) Update(ctx context.Context, wf *nbmpv2.Workflow) error {
	l := log.FromContext(ctx, "workflowID", wf.General.ID)
	l.V(2).Info("updating workflow")
	PopulateWorkflowDefaults(wf)

	return nil
}

func (wfsvc *workflowService) Delete(ctx context.Context, wf *nbmpv2.Workflow) error {
	l := log.FromContext(ctx, "workflowID", wf.General.ID)
	l.V(2).Info("deleting workflow")
	PopulateWorkflowDefaults(wf)

	return nil
}

func (wfsvc *workflowService) Retrieve(ctx context.Context, wf *nbmpv2.Workflow) error {
	l := log.FromContext(ctx, "workflowID", wf.General.ID)
	l.V(2).Info("retrieving workflow")
	PopulateWorkflowDefaults(wf)

	return nil
}

func PopulateWorkflowDefaults(wf *nbmpv2.Workflow) {
	wf.Scheme = &nbmpv2.Scheme{
		URI: nbmpv2.SchemaURI,
	}
}
