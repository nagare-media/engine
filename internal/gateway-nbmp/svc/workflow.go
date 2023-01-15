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

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/internal/pkg/uuid"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

//+kubebuilder:rbac:groups=engine.nagare.media,resources=workflows,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=engine.nagare.media,resources=tasks,verbs=get;list;watch;create;update;patch;delete

type WorkflowService interface {
	Create(ctx context.Context, wf *nbmpv2.Workflow) error
	Update(ctx context.Context, wf *nbmpv2.Workflow) error
	Delete(ctx context.Context, wf *nbmpv2.Workflow) error
	Retrieve(ctx context.Context, wf *nbmpv2.Workflow) error
}

type workflowService struct {
	cfg *enginev1.GatewayNBMPConfiguration
	k8s client.Client
}

var _ WorkflowService = &workflowService{}

func NewWorkflowService(cfg *enginev1.GatewayNBMPConfiguration, k8sClient client.Client) *workflowService {
	return &workflowService{
		cfg: cfg,
		k8s: k8sClient,
	}
}

func (s *workflowService) Create(ctx context.Context, wf *nbmpv2.Workflow) error {
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

func (s *workflowService) Update(ctx context.Context, wf *nbmpv2.Workflow) error {
	l := log.FromContext(ctx, "workflowID", wf.General.ID)
	l.V(2).Info("updating workflow")
	PopulateWorkflowDefaults(wf)

	return nil
}

func (s *workflowService) Delete(ctx context.Context, wf *nbmpv2.Workflow) error {
	l := log.FromContext(ctx, "workflowID", wf.General.ID)
	l.V(2).Info("deleting workflow")
	PopulateWorkflowDefaults(wf)

	return nil
}

func (s *workflowService) Retrieve(ctx context.Context, wf *nbmpv2.Workflow) error {
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
