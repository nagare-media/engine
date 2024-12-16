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

package svc

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	nbmpconvv2 "github.com/nagare-media/engine/internal/pkg/nbmpconv/v2"
	nbmpsvcv2 "github.com/nagare-media/engine/pkg/nbmp/svc/v2"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

// +kubebuilder:rbac:groups=engine.nagare.media,resources=tasks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=engine.nagare.media,resources=workflows,verbs=get;list;watch;create;update;patch;delete

type workflowService struct {
	cfg *enginev1.GatewayNBMPWorkflowServiceConfig
	k8s client.Client
}

var _ nbmpsvcv2.WorkflowService = &workflowService{}

func NewWorkflowService(cfg *enginev1.GatewayNBMPWorkflowServiceConfig, k8sClient client.Client) *workflowService {
	return &workflowService{
		cfg: cfg,
		k8s: k8sClient,
	}
}

func (s *workflowService) Create(ctx context.Context, wf *nbmpv2.Workflow) error {
	l := log.FromContext(ctx, "workflowID", wf.General.ID)
	l.V(1).Info("create workflow")

	// convert to Kubernetes resources
	w := enginev1.Workflow{}
	convw := nbmpconvv2.NewWDDToWorkflowConverter(wf)
	if err := convw.Convert(&w); err != nil {
		return err
	}

	tsks := []enginev1.Task{}
	convtsks := nbmpconvv2.NewWDDToTasksConverter(
		s.k8s.Scheme(),
		s.cfg.Kubernetes.GPUResourceName,
		wf,
	)
	if err := convtsks.Convert(&tsks); err != nil {
		return err
	}

	// create Kubernetes resources
	if err := s.k8s.Create(ctx, &w); err != nil {
		return err
	}

	for _, t := range tsks {
		s.setCommonLabels(t.Labels, wf)
		if err := s.k8s.Create(ctx, &t); err != nil {
			return err
		}
	}

	// return latest version
	return s.Retrieve(ctx, wf)
}

func (s *workflowService) Update(ctx context.Context, wf *nbmpv2.Workflow) error {
	l := log.FromContext(ctx, "workflowID", wf.General.ID)
	l.V(1).Info("update workflow")

	// retrieve current Kubernetes resources
	wOld := enginev1.Workflow{}
	err := s.k8s.Get(ctx, client.ObjectKey{Name: wf.General.ID}, &wOld)
	if err != nil {
		return err
	}

	tsksOld := &enginev1.TaskList{}
	sel := client.MatchingLabels{}
	s.setCommonLabels(sel, wf)
	err = s.k8s.List(ctx, tsksOld, sel)
	if err != nil {
		return err
	}

	// convert new to Kubernetes resources
	wNew := enginev1.Workflow{}
	convw := nbmpconvv2.NewWDDToWorkflowConverter(wf)
	if err := convw.Convert(&wNew); err != nil {
		return err
	}

	tsksNew := []enginev1.Task{}
	convtsks := nbmpconvv2.NewWDDToTasksConverter(
		s.k8s.Scheme(),
		s.cfg.Kubernetes.GPUResourceName,
		wf,
	)
	if err := convtsks.Convert(&tsksNew); err != nil {
		return err
	}

	// compare old and new
	tIDs := make(map[string]struct{}, len(tsksNew))
	inOldTsks := make(map[string]*enginev1.Task, len(tsksOld.Items))
	inNewTsks := make(map[string]*enginev1.Task, len(tsksNew))

	for _, t := range tsksOld.Items {
		tIDs[t.Name] = struct{}{}
		inOldTsks[t.Name] = &t
	}

	for _, t := range tsksNew {
		tIDs[t.Name] = struct{}{}
		inNewTsks[t.Name] = &t
	}

	// update Kubernetes resources
	// TODO: should we DeepCopy and try to merge with existing resources instead of overwrite?
	wNew.SetResourceVersion(wOld.GetResourceVersion())
	if err = s.k8s.Update(ctx, &wNew); err != nil {
		return err
	}

	for id := range tIDs {
		tOld, isOld := inOldTsks[id]
		tNew, isNew := inNewTsks[id]

		switch {
		case !isOld && isNew:
			// create task
			s.setCommonLabels(tNew.Labels, wf)
			if err = s.k8s.Create(ctx, tNew); err != nil {
				return err
			}

		case isOld && !isNew:
			// delete task
			if err = s.k8s.Delete(ctx, tOld); err != nil {
				return err
			}

		case isOld && isNew:
			// update task
			tNew.SetResourceVersion(tOld.GetResourceVersion())
			s.setCommonLabels(tNew.Labels, wf)
			if err = s.k8s.Update(ctx, tNew); err != nil {
				return err
			}
		}
	}

	// return latest version
	return s.Retrieve(ctx, wf)
}

func (s *workflowService) Delete(ctx context.Context, wf *nbmpv2.Workflow) error {
	l := log.FromContext(ctx, "workflowID", wf.General.ID)
	l.V(1).Info("delete workflow")

	// convert to Kubernetes resources
	w := &enginev1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: wf.General.ID,
		},
	}

	// delete Kubernetes resources
	// Tasks have a owner reference and will be deleted automatically
	if err := s.k8s.Delete(ctx, w); err != nil {
		return err
	}

	return nil
}

func (s *workflowService) Retrieve(ctx context.Context, wf *nbmpv2.Workflow) error {
	l := log.FromContext(ctx, "workflowID", wf.General.ID)
	l.V(1).Info("retrieve workflow")

	// retrieve Kubernetes resources
	w := enginev1.Workflow{}
	err := s.k8s.Get(ctx, client.ObjectKey{Name: wf.General.ID}, &w)
	if err != nil {
		return err
	}

	tsks := &enginev1.TaskList{}
	sel := client.MatchingLabels{}
	s.setCommonLabels(sel, wf)
	err = s.k8s.List(ctx, tsks, sel)
	if err != nil {
		return err
	}

	// convert to NBMP workflow description document
	wfNew := nbmpv2.Workflow{}
	convw := nbmpconvv2.NewWorkflowToWDDConverter(&w)
	if err := convw.Convert(&wfNew); err != nil {
		return err
	}

	convtsks := nbmpconvv2.NewTasksToWDDConverter(
		s.cfg.Kubernetes.GPUResourceName,
		tsks.Items,
	)
	if err := convtsks.Convert(&wfNew); err != nil {
		return err
	}

	*wf = wfNew

	return nil
}

func (s *workflowService) setCommonLabels(labels map[string]string, wf *nbmpv2.Workflow) {
	labels[enginev1.WorkflowNamespaceLabel] = s.cfg.Kubernetes.Namespace
	labels[enginev1.WorkflowNameLabel] = wf.General.ID
}
