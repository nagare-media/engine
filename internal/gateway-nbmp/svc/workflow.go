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

package svc

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/internal/pkg/uuid"
	nbmpsvcv2 "github.com/nagare-media/engine/pkg/nbmp/svc/v2"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

// +kubebuilder:rbac:groups=engine.nagare.media,resources=workflows,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=engine.nagare.media,resources=tasks,verbs=get;list;watch;create;update;patch;delete

type workflowService struct {
	cfg *enginev1.WorkflowServiceConfiguration
	k8s client.Client
}

var _ nbmpsvcv2.WorkflowService = &workflowService{}

func NewWorkflowService(cfg *enginev1.WorkflowServiceConfiguration, k8sClient client.Client) *workflowService {
	return &workflowService{
		cfg: cfg,
		k8s: k8sClient,
	}
}

func (s *workflowService) Create(ctx context.Context, wf *nbmpv2.Workflow) error {
	l := log.FromContext(ctx)
	PopulateWorkflowDefaults(wf)

	// validate
	err := ValidateWorkflowCreate(wf)
	if err != nil {
		return err
	}

	// initialize
	wf.General.ID = uuid.UUIDv4()
	l = l.WithValues("workflowID", wf.General.ID)
	l.V(2).Info("creating workflow")
	wf.General.State = &nbmpv2.InstantiatedState

	// convert to Kubernetes resources
	w, err := s.wddToWorkflow(wf)
	if err != nil {
		return err
	}

	tasks, err := s.wddToTasks(wf, w)
	if err != nil {
		return err
	}

	// create Kubernetes resources
	err = s.k8s.Create(ctx, w)
	if err != nil {
		return apiErrorHandler(err)
	}

	for _, t := range tasks {
		err = s.k8s.Create(ctx, t)
		if err != nil {
			return apiErrorHandler(err)
		}
	}

	return nil
}

func (s *workflowService) Update(ctx context.Context, wf *nbmpv2.Workflow) error {
	l := log.FromContext(ctx, "workflowID", wf.General.ID)
	l.V(2).Info("updating workflow")
	PopulateWorkflowDefaults(wf)

	// validate
	err := ValidateWorkflowCommon(wf)
	if err != nil {
		return err
	}

	// convert to Kubernetes resources
	w, err := s.wddToWorkflow(wf)
	if err != nil {
		return err
	}

	tasks, err := s.wddToTasks(wf, w)
	if err != nil {
		return err
	}

	// update Kubernetes resources
	err = s.k8s.Update(ctx, w)
	if err != nil {
		return apiErrorHandler(err)
	}

	for _, t := range tasks {
		err = s.k8s.Update(ctx, t)
		if err != nil {
			return apiErrorHandler(err)
		}
	}

	// return latest version
	return s.Retrieve(ctx, wf)
}

func (s *workflowService) Delete(ctx context.Context, wf *nbmpv2.Workflow) error {
	l := log.FromContext(ctx, "workflowID", wf.General.ID)
	l.V(2).Info("deleting workflow")
	PopulateWorkflowDefaults(wf)

	// validate
	err := ValidateWorkflowDelete(wf)
	if err != nil {
		return err
	}

	// convert to Kubernetes resources
	w := &enginev1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: wf.General.ID,
		},
	}

	// delete Kubernetes resources
	// Tasks have a owner reference and will be deleted automatically
	err = s.k8s.Delete(ctx, w)
	if err != nil {
		return apiErrorHandler(err)
	}

	return nil
}

func (s *workflowService) Retrieve(ctx context.Context, wf *nbmpv2.Workflow) error {
	l := log.FromContext(ctx, "workflowID", wf.General.ID)
	l.V(2).Info("retrieving workflow")
	PopulateWorkflowDefaults(wf)

	// validate
	err := ValidateWorkflowRetrieve(wf)
	if err != nil {
		return err
	}

	// retrieve Kubernetes resources
	w := &enginev1.Workflow{}
	err = s.k8s.Get(ctx, client.ObjectKey{Name: wf.General.ID}, w)
	if err != nil {
		return apiErrorHandler(err)
	}
	if w.Labels[IsNBMPLabel] != "true" {
		return nbmpsvcv2.ErrNotFound
	}

	tasks := &enginev1.TaskList{}
	err = s.k8s.List(ctx, tasks, client.MatchingLabels{
		IsNBMPLabel:                "true",
		enginev1.WorkflowNameLabel: wf.General.ID,
		// TODO: set namespace
	})
	if err != nil {
		return apiErrorHandler(err)
	}

	// convert to NBMP workflow description document
	// TODO: implement

	return nil
}
