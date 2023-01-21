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

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/event"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
)

const (
	TaskControllerName = "nagare-media-engine-task-controller"
)

// TaskReconciler reconciles a Task object
type TaskReconciler struct {
	client.Client

	Config                          enginev1.NagareMediaEngineControllerManagerConfiguration
	Scheme                          *runtime.Scheme
	JobEventChannel                 <-chan event.GenericEvent
	MediaProcessingEntityReconciler *MediaProcessingEntityReconciler
}

// +kubebuilder:rbac:groups=engine.nagare.media,resources=tasks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=engine.nagare.media,resources=tasks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=tasks/finalizers,verbs=update
// +kubebuilder:rbac:groups=engine.nagare.media,resources=mediaprocessingentities,verbs=get;list;watch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=clustermediaprocessingentities,verbs=get;list;watch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=functions,verbs=get;list;watch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=clusterfunctions,verbs=get;list;watch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=tasktemplates,verbs=get;list;watch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=clustertasktemplates,verbs=get;list;watch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=workflows,verbs=get;list;watch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=clusterworkflows,verbs=get;list;watch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Task object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.13.1/pkg/reconcile
func (r *TaskReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named(TaskControllerName).
		For(&enginev1.Task{}).
		Watches(&source.Channel{Source: r.JobEventChannel}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}
