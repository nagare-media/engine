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
	"strings"

	"github.com/nagare-media/engine/internal/manager/predicate"
	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
)

const (
	JobControllerName = "nagare-media-engine-job-controller"
)

// JobReconciler reconciles Job objects
type JobReconciler struct {
	client.Client

	Config       enginev1.NagareMediaEngineControllerManagerConfiguration
	Scheme       *runtime.Scheme
	EventChannel chan<- event.GenericEvent
}

// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch

func (r *JobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

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

	// Reconciliation logic is handled by task controller. However, this it is managed by another manager connected to the
	// management cluster. We therefore use an event channel to trigger a reconciliation of the accompanying task.

	taskKey := job.Labels[enginev1.TaskLabel]

	if n := strings.Count(taskKey, string(types.Separator)); n != 1 {
		err := fmt.Errorf("task label %s is illformed", enginev1.TaskLabel)
		log.Error(err, "error determining Task")
		return ctrl.Result{}, err
	}

	i := strings.Index(taskKey, string(types.Separator))
	r.EventChannel <- event.GenericEvent{
		Object: &metav1.PartialObjectMetadata{
			ObjectMeta: metav1.ObjectMeta{
				Name:      taskKey[:i],
				Namespace: taskKey[i+1:],
			},
		},
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named(JobControllerName).
		For(&batchv1.Job{}).
		WithEventFilter(predicate.HasLabels{enginev1.TaskLabel}).
		Complete(r)
}
