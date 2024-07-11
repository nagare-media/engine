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

package controllers

import (
	"context"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/internal/workflow-manager/predicate"
	"github.com/nagare-media/engine/pkg/apis/meta"
)

const (
	JobControllerName = "nagare-media-engine-job-controller"
)

// JobReconciler reconciles Job objects
type JobReconciler struct {
	client.Client
	APIReader client.Reader

	Config                   *enginev1.WorkflowManagerConfiguration
	Scheme                   *runtime.Scheme
	EventChannel             chan<- event.GenericEvent
	MediaProcessingEntityRef *meta.ObjectReference
}

// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;update

func (r *JobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("reconcile Job")

	// fetch Job
	job := &batchv1.Job{}
	if err := r.Client.Get(ctx, req.NamespacedName, job); err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return ctrl.Result{}, nil
		}
		log.Error(err, "error fetching Job")
		return ctrl.Result{}, err
	}

	// get Task infos
	taskName := job.Labels[enginev1.TaskNameLabel]
	taskNamespace := job.Labels[enginev1.TaskNamespaceLabel]

	// handle delete
	if !job.DeletionTimestamp.IsZero() {
		// remove finalizer
		controllerutil.RemoveFinalizer(job, enginev1.JobProtectionFinalizer)
		if err := r.Client.Update(ctx, job); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Reconciliation logic is handled by task controller. However, this it is managed by another manager connected to the
	// management cluster. We therefore use an event channel to trigger a reconciliation of the accompanying task.
	r.EventChannel <- event.GenericEvent{
		Object: &metav1.PartialObjectMetadata{
			ObjectMeta: metav1.ObjectMeta{
				Name:      taskName,
				Namespace: taskNamespace,
			},
		},
	}

	return ctrl.Result{}, nil
}

func (r *JobReconciler) Name() string {
	return fmt.Sprintf("%s-%s-%s", r.MediaProcessingEntityRef.Namespace, r.MediaProcessingEntityRef.Name, JobControllerName)
}

// SetupWithManager sets up the controller with the Manager.
func (r *JobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named(r.Name()).
		For(&batchv1.Job{}).
		WithEventFilter(predicate.HasLabels{enginev1.TaskNamespaceLabel, enginev1.TaskNameLabel}).
		Complete(r)
}
