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

package controllers

import (
	"context"
	"errors"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/internal/pkg/apis/utils"
)

const (
	WorkflowControllerName = "nagare-media-engine-workflow-controller"
)

// WorkflowReconciler reconciles a Workflow object
type WorkflowReconciler struct {
	client.Client
	APIReader client.Reader

	Config *enginev1.WorkflowManagerConfig
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=engine.nagare.media,resources=workflows,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=engine.nagare.media,resources=workflows/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=workflows/finalizers,verbs=update
// +kubebuilder:rbac:groups=engine.nagare.media,resources=tasks,verbs=get;list;watch

func (r *WorkflowReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := logf.FromContext(ctx)

	// fetch Workflow
	wf := &enginev1.Workflow{}
	if err := r.Client.Get(ctx, req.NamespacedName, wf); err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return ctrl.Result{}, nil
		}
		log.Error(err, "error fetching Workflow")
		return ctrl.Result{}, err
	}

	// apply patches on return
	oldWf := wf.DeepCopy()
	defer func() {
		r.reconcileCondition(ctx, reterr, wf)
		apiErrs := kerrors.FilterOut(utils.FullPatch(ctx, r.Client, wf, oldWf), apierrors.IsNotFound)
		if apiErrs != nil {
			log.Error(apiErrs, "error patching Workflow")
		}
		reterr = kerrors.Flatten(kerrors.NewAggregate([]error{apiErrs, reterr}))
	}()

	// add finalizers
	if !controllerutil.ContainsFinalizer(wf, enginev1.WorkflowProtectionFinalizer) {
		controllerutil.AddFinalizer(wf, enginev1.WorkflowProtectionFinalizer)
		return ctrl.Result{}, nil
	}

	// handle delete
	if !wf.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, wf)
	}

	// handle normal reconciliation
	return r.reconcile(ctx, wf)
}

func (r *WorkflowReconciler) reconcileCondition(ctx context.Context, err error, wf *enginev1.Workflow) {
	log := logf.FromContext(ctx)

	var phaseConditionStatus corev1.ConditionStatus
	if err != nil {
		log.Error(err, "error reconciling Workflow")
		wf.Status.Message = err.Error()
		phaseConditionStatus = corev1.ConditionFalse
	} else {
		wf.Status.Message = ""
		phaseConditionStatus = corev1.ConditionTrue
	}

	switch wf.Status.Phase {
	default:
		wf.Status.Conditions = []enginev1.Condition{}

	case enginev1.InitializingWorkflowPhase:
		wf.Status.Conditions = utils.MarkConditionFalse(wf.Status.Conditions, enginev1.WorkflowReadyConditionType)

	case enginev1.RunningWorkflowPhase, enginev1.AwaitingCompletionWorkflowPhase:
		wf.Status.Conditions = utils.MarkCondition(wf.Status.Conditions, enginev1.WorkflowReadyConditionType, phaseConditionStatus)

	case enginev1.SucceededWorkflowPhase:
		wf.Status.Conditions = utils.MarkConditionFalse(wf.Status.Conditions, enginev1.WorkflowReadyConditionType)
		wf.Status.Conditions = utils.MarkConditionTrue(wf.Status.Conditions, enginev1.WorkflowCompleteConditionType)

	case enginev1.FailedWorkflowPhase:
		wf.Status.Conditions = utils.MarkConditionFalse(wf.Status.Conditions, enginev1.WorkflowReadyConditionType)
		wf.Status.Conditions = utils.MarkConditionTrue(wf.Status.Conditions, enginev1.WorkflowFailedConditionType)
	}
}

func (r *WorkflowReconciler) reconcileDelete(ctx context.Context, wf *enginev1.Workflow) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("reconcile deleted Workflow")

	// wait until all Task have terminated
	if res, err := r.reconcileTasks(ctx, wf); err != nil {
		return res, err
	}
	if wf.Status.Active != nil && *wf.Status.Active > 0 {
		return ctrl.Result{}, errors.New("not all Tasks have terminated")
	}

	// remove finalizer
	controllerutil.RemoveFinalizer(wf, enginev1.WorkflowProtectionFinalizer)

	return ctrl.Result{}, nil
}

func (r *WorkflowReconciler) reconcile(ctx context.Context, wf *enginev1.Workflow) (ctrl.Result, error) {
	log := logf.FromContext(ctx, "phase", wf.Status.Phase)
	log.Info("reconcile Workflow")

	if wf.Status.QueuedTime == nil {
		wf.Status.QueuedTime = &metav1.Time{Time: time.Now()}
	}

	// set Workflow label for filtering
	if wf.Labels == nil {
		wf.Labels = make(map[string]string)
	}
	wf.Labels[enginev1.WorkflowNamespaceLabel] = wf.Namespace
	wf.Labels[enginev1.WorkflowNameLabel] = wf.Name

	switch wf.Status.Phase {
	default:
		// empty or unknown phase -> move to initializing
		wf.Status.Phase = enginev1.InitializingWorkflowPhase
		wf.Status.Total = nil
		wf.Status.Active = nil
		wf.Status.Succeeded = nil
		wf.Status.Failed = nil

	case enginev1.InitializingWorkflowPhase:
		if res, err := r.reconcileTasks(ctx, wf); err != nil {
			return res, err
		}

		// as soon as we see Tasks belonging to this Workflow transition to running phase
		if wf.Status.Total != nil && *wf.Status.Total > 0 {
			wf.Status.Phase = enginev1.RunningWorkflowPhase
			wf.Status.StartTime = &metav1.Time{Time: time.Now()}
		}

	case enginev1.RunningWorkflowPhase:
		if res, err := r.reconcileTasks(ctx, wf); err != nil {
			return res, err
		}

		// TODO: ensure trace is started
		// TODO: ensure NATS stream exist

		// If we have no active Tasks, do not transition to succeeded directly as new Tasks could be created. This helps
		// mitigate races during initial setup and very quick Tasks:
		//   Workflow w is created
		//   Task t1 is created
		//   Task t1 terminates very quick
		//   Workflow w is marked as successful
		//   Task t2 is created and fails because w has already terminated
		if wf.Status.Active != nil && *wf.Status.Active == 0 {
			wf.Status.Phase = enginev1.AwaitingCompletionWorkflowPhase
			wf.Status.EndTime = &metav1.Time{Time: time.Now()}
		}

	case enginev1.AwaitingCompletionWorkflowPhase:
		if res, err := r.reconcileTasks(ctx, wf); err != nil {
			return res, err
		}

		// check if indeed new Tasks have been created and are active
		if wf.Status.Active != nil && *wf.Status.Active > 0 {
			wf.Status.Phase = enginev1.RunningWorkflowPhase
			return ctrl.Result{}, nil
		}

		// check if we waited enough
		if wf.Status.EndTime == nil {
			// this state should not happen: move back to running to set end time
			wf.Status.Phase = enginev1.RunningWorkflowPhase
			return ctrl.Result{}, nil
		}

		waitingDuration := time.Since(wf.Status.EndTime.Time)
		remainingDuration := r.Config.WorkflowTerminationWaitingDuration.Duration - waitingDuration
		if remainingDuration > 0 {
			return ctrl.Result{RequeueAfter: remainingDuration}, nil
		}

		wf.Status.Phase = enginev1.SucceededWorkflowPhase

	case enginev1.SucceededWorkflowPhase, enginev1.FailedWorkflowPhase:
		// Workflow has terminated: nothing to do
		// TODO: stop trace
		// TODO: delete NATS stream
	}

	return ctrl.Result{}, nil
}

func (r *WorkflowReconciler) reconcileTasks(ctx context.Context, wf *enginev1.Workflow) (ctrl.Result, error) {
	var total, active, succeeded, failed int32

	// fetch Tasks
	taskList := &enginev1.TaskList{}
	err := r.List(ctx, taskList, client.MatchingLabels{
		enginev1.WorkflowNamespaceLabel: wf.Namespace,
		enginev1.WorkflowNameLabel:      wf.Name,
	})
	if err != nil {
		return ctrl.Result{}, nil
	}

	// check Tasks
	total = int32(len(taskList.Items))
	for _, task := range taskList.Items {
		switch task.Status.Phase {
		default:
			active++
		case enginev1.SucceededTaskPhase:
			succeeded++
		case enginev1.FailedTaskPhase:
			failed++
			r.reconcileFailedTask(ctx, wf, &task)
		}
	}

	wf.Status.Total = &total
	wf.Status.Active = &active
	wf.Status.Succeeded = &succeeded
	wf.Status.Failed = &failed

	return ctrl.Result{}, nil
}

func (r *WorkflowReconciler) reconcileFailedTask(ctx context.Context, wf *enginev1.Workflow, task *enginev1.Task) {
	if utils.WorkflowHasTerminated(wf) {
		// ignore failed tasks if Workflow already terminated
		return
	}

	// TODO: this should probably be handled by the task controller?

	policy := enginev1.FailWorkflowJobFailurePolicyAction
	if task.Spec.JobFailurePolicy != nil && task.Spec.JobFailurePolicy.DefaultAction != nil {
		policy = *task.Spec.JobFailurePolicy.DefaultAction
	}

	switch policy {
	case enginev1.IgnoreJobFailurePolicyAction:
	case enginev1.FailWorkflowJobFailurePolicyAction:
		wf.Status.Phase = enginev1.FailedWorkflowPhase
	}
}

// TODO: emit Kubernetes events
// SetupWithManager sets up the controller with the Manager.
func (r *WorkflowReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named(WorkflowControllerName).
		For(&enginev1.Workflow{}).
		Owns(&enginev1.Task{}).
		Complete(r)
}
