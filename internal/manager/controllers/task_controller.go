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
	"errors"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/internal/pkg/apis/utils"
	"github.com/nagare-media/engine/pkg/apis/meta"
)

const (
	TaskControllerName = "nagare-media-engine-task-controller"
)

// TaskReconciler reconciles a Task object
type TaskReconciler struct {
	client.Client
	APIReader client.Reader

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
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;create;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;create;update;patch

func (r *TaskReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := logf.FromContext(ctx)

	// fetch Task
	task := &enginev1.Task{}
	if err := r.Client.Get(ctx, req.NamespacedName, task); err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return ctrl.Result{}, nil
		}
		log.Error(err, "error fetching Task")
		return ctrl.Result{}, err
	}

	// apply patches on return
	oldTask := task.DeepCopy()
	defer func() {
		r.reconcileCondition(ctx, reterr, task)
		apiErrs := kerrors.FilterOut(utils.FullPatch(ctx, r.Client, task, oldTask), apierrors.IsNotFound)
		if apiErrs != nil {
			log.Error(apiErrs, "error patching Task")
		}
		reterr = kerrors.Flatten(kerrors.NewAggregate([]error{apiErrs, reterr}))
	}()

	// add finalizers
	if !controllerutil.ContainsFinalizer(task, enginev1.TaskProtectionFinalizer) {
		controllerutil.AddFinalizer(task, enginev1.TaskProtectionFinalizer)
		return ctrl.Result{}, nil
	}

	// always normalize status references
	if err := r.normalizeStatusReferences(task); err != nil {
		return ctrl.Result{}, err
	}

	// handle delete
	if !task.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, task)
	}

	// handle normal reconciliation
	return r.reconcile(ctx, task)
}

func (r *TaskReconciler) reconcileCondition(ctx context.Context, err error, task *enginev1.Task) {
	log := logf.FromContext(ctx)

	var phaseConditionStatus corev1.ConditionStatus
	if err != nil {
		log.Error(err, "error reconciling Task")
		task.Status.Message = err.Error()
		phaseConditionStatus = corev1.ConditionFalse
	} else {
		task.Status.Message = ""
		phaseConditionStatus = corev1.ConditionTrue
	}

	switch task.Status.Phase {
	default:
		task.Status.Conditions = []enginev1.Condition{}

	case enginev1.TaskPhaseInitializing:
		task.Status.Conditions = utils.MarkConditionFalse(task.Status.Conditions, enginev1.TaskInitializedConditionType)

	case enginev1.TaskPhaseJobPending:
		task.Status.Conditions = utils.MarkCondition(task.Status.Conditions, enginev1.TaskInitializedConditionType, phaseConditionStatus)

	case enginev1.TaskPhaseRunning:
		task.Status.Conditions = utils.MarkConditionTrue(task.Status.Conditions, enginev1.TaskInitializedConditionType)
		task.Status.Conditions = utils.MarkCondition(task.Status.Conditions, enginev1.TaskReadyConditionType, phaseConditionStatus)

	case enginev1.TaskPhaseSucceeded:
		task.Status.Conditions = utils.MarkConditionTrue(task.Status.Conditions, enginev1.TaskInitializedConditionType)
		task.Status.Conditions = utils.MarkConditionFalse(task.Status.Conditions, enginev1.TaskReadyConditionType)
		task.Status.Conditions = utils.MarkConditionTrue(task.Status.Conditions, enginev1.TaskCompleteConditionType)

	case enginev1.TaskPhaseFailed:
		task.Status.Conditions = utils.MarkConditionTrue(task.Status.Conditions, enginev1.TaskInitializedConditionType)
		task.Status.Conditions = utils.MarkConditionFalse(task.Status.Conditions, enginev1.TaskReadyConditionType)
		task.Status.Conditions = utils.MarkConditionTrue(task.Status.Conditions, enginev1.TaskFailedConditionType)
	}
}

func (r *TaskReconciler) reconcileDelete(ctx context.Context, task *enginev1.Task) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("reconcile deleted Task")

	// terminate if necessary
	if res, err := r.reconcileTerminatedTask(ctx, task, true); err != nil {
		return res, err
	}

	// remove finalizer
	controllerutil.RemoveFinalizer(task, enginev1.TaskProtectionFinalizer)

	return ctrl.Result{}, nil
}

func (r *TaskReconciler) reconcile(ctx context.Context, task *enginev1.Task) (ctrl.Result, error) {
	log := logf.FromContext(ctx, "phase", task.Status.Phase)
	log.Info("reconcile Task")

	if task.Status.QueuedTime == nil {
		task.Status.QueuedTime = &metav1.Time{Time: time.Now()}
	}

	// always reconcile Workflow as it might have terminated
	if res, err := r.reconcileWorkflow(ctx, task); err != nil {
		return res, err
	}

	switch task.Status.Phase {
	default:
		// empty or unknown phase -> move to initializing
		task.Status.Phase = enginev1.TaskPhaseInitializing

	case enginev1.TaskPhaseInitializing:
		if res, err := r.reconcileMediaProcessingEntity(ctx, task); err != nil {
			return res, err
		}
		if res, err := r.reconcileFunction(ctx, task); err != nil {
			return res, err
		}
		task.Status.Phase = enginev1.TaskPhaseJobPending

	case enginev1.TaskPhaseJobPending:
		if res, err := r.reconcilePendingJob(ctx, task); err != nil {
			return res, err
		}
		task.Status.StartTime = &metav1.Time{Time: time.Now()}
		task.Status.Phase = enginev1.TaskPhaseRunning

	case enginev1.TaskPhaseRunning:
		if res, err := r.reconcileRunningJob(ctx, task); err != nil {
			return res, err
		}
		// transition to Succeeded or Failed or just make sure everything is still running
		task.Status.EndTime = &metav1.Time{Time: time.Now()}

	case enginev1.TaskPhaseSucceeded, enginev1.TaskPhaseFailed:
		if res, err := r.reconcileTerminatedTask(ctx, task, false); err != nil {
			return res, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *TaskReconciler) reconcileTerminatedTask(ctx context.Context, task *enginev1.Task, forceDeleteJob bool) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	if task.Status.JobRef != nil {
		// get correct API client
		c, err := r.getJobClientForTask(task)
		if err != nil {
			if forceDeleteJob {
				// We could retry, but we assume the job is well behaved and will stop executing eventually. Kubernetes should
				// cleanup after some time.
				log.Error(err, "ignoring dangling job reference")
				task.Status.JobRef = nil
			}
			return ctrl.Result{}, err
		}

		// resolve job
		job, err := r.resolveJobRef(ctx, task)
		if err != nil {
			switch {
			case apierrors.IsNotFound(err):
				// already deleted
				task.Status.JobRef = nil
				return ctrl.Result{}, nil
			case apierrors.IsConflict(err):
				// job UID does not match the job reference: log error and assume that job already terminated
				log.Error(err, "conflicting jobRef: assuming original job terminated")
				task.Status.JobRef = nil
				return ctrl.Result{}, nil
			case forceDeleteJob:
				// We could retry, but we assume the job is well behaved and will stop executing eventually. Kubernetes should
				// cleanup after some time.
				log.Error(err, "ignoring dangling job reference")
				task.Status.JobRef = nil
			}
			return ctrl.Result{}, err
		}

		// delete job if necessary
		// We want to keep the job around if the Task still exist and it also already terminated. This allows to read logs
		// and see more details than on a Task.
		if forceDeleteJob || utils.JobIsActive(job) {
			if err := c.Delete(ctx, job, client.Preconditions{UID: &task.Status.JobRef.UID}); err != nil {
				switch {
				case apierrors.IsNotFound(err):
					// already deleted so nothing to do
					task.Status.JobRef = nil
					return ctrl.Result{}, nil
				case apierrors.IsConflict(err):
					// job UID does not match the job reference: log error and assume that job already terminated
					log.Error(err, "conflicting jobRef: assuming original job terminated")
					task.Status.JobRef = nil
					return ctrl.Result{}, nil
				case forceDeleteJob:
					// We could retry, but we assume the job is well behaved and will stop executing eventually. Kubernetes
					// should cleanup after some time.
					log.Error(err, "ignoring dangling job reference")
					task.Status.JobRef = nil
				}
				return ctrl.Result{}, err
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *TaskReconciler) reconcileWorkflow(ctx context.Context, task *enginev1.Task) (ctrl.Result, error) {
	// fetch Workflow
	wf := &enginev1.Workflow{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: task.Namespace, Name: task.Spec.WorkflowRef.Name}, wf); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to fetch Workflow: %w", err)
	}

	// set owner reference
	task.OwnerReferences = utils.EnsureOwnerRef(task.OwnerReferences, metav1.OwnerReference{
		APIVersion:         enginev1.GroupVersion.String(),
		Kind:               wf.GroupVersionKind().Kind,
		Name:               wf.Name,
		UID:                wf.UID,
		Controller:         pointer.Bool(true),
		BlockOwnerDeletion: pointer.Bool(true),
	})

	// is Workflow marked for deletion
	if !wf.DeletionTimestamp.IsZero() {
		// we have to delete this Task ourselves as we use blockOwnerDeletion
		if err := r.Client.Delete(ctx, task, client.Preconditions{UID: &task.UID}); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to delete Task: %w", err)
		}
	}

	// set Workflow and Task label for filtering
	if task.Labels == nil {
		task.Labels = make(map[string]string)
	}
	task.Labels[enginev1.WorkflowNamespaceLabel] = wf.Namespace
	task.Labels[enginev1.WorkflowNameLabel] = wf.Name
	task.Labels[enginev1.TaskNamespaceLabel] = task.Namespace
	task.Labels[enginev1.TaskNameLabel] = task.Name

	// check termination status of Workflow
	if utils.WorkflowHasTerminated(wf) && utils.TaskIsActive(task) {
		// fail this Task since Workflow has already terminated
		// TODO: set reason
		task.Status.Phase = enginev1.TaskPhaseFailed
	}

	return ctrl.Result{}, nil
}

func (r *TaskReconciler) reconcileMediaProcessingEntity(ctx context.Context, task *enginev1.Task) (ctrl.Result, error) {
	// a MediaProcessingEntity can be selected through various ways with different precedence:

	// 1. MediaProcessingEntityRef
	if task.Spec.MediaProcessingEntityRef != nil {
		ref, err := utils.LocalMediaProcessingEntityToObjectRef(task.Spec.MediaProcessingEntityRef, task.Namespace)
		if err != nil {
			return ctrl.Result{}, err
		}
		if err = utils.NormalizeMediaProcessingEntityRef(r.Scheme, ref); err != nil {
			return ctrl.Result{}, err
		}
		task.Status.MediaProcessingEntityRef = ref
		return ctrl.Result{}, nil
	}

	// 2. MediaProcessingEntitySelector
	if task.Spec.MediaProcessingEntitySelector != nil {
		sel, err := metav1.LabelSelectorAsSelector(task.Spec.MediaProcessingEntitySelector)
		if err != nil {
			return ctrl.Result{}, err
		}

		ref, err := utils.SelectMediaProcessingEntityRef(ctx, r.Client, task.Namespace, sel)
		if err != nil {
			return ctrl.Result{}, err
		}
		if ref != nil {
			if err = utils.NormalizeMediaProcessingEntityRef(r.Scheme, ref); err != nil {
				return ctrl.Result{}, err
			}
			task.Status.MediaProcessingEntityRef = ref
			return ctrl.Result{}, nil
		}
	}

	// 3. TaskTemplateRef
	if task.Spec.TaskTemplateRef != nil {
		// shallow copy
		ttRef := task.Spec.TaskTemplateRef
		if err := utils.NormalizeLocalTaskTemplateRef(r.Scheme, ttRef); err != nil {
			return ctrl.Result{}, err
		}

		// resolve TaskTemplate
		ttObj, err := utils.ResolveLocalRef(ctx, r.Client, task.Namespace, task.Spec.TaskTemplateRef)
		if !apierrors.IsNotFound(err) {
			if err != nil {
				return ctrl.Result{}, err
			}

			var ttMPERef *meta.LocalObjectReference
			var ttMPESel *metav1.LabelSelector

			switch tt := ttObj.(type) {
			case *enginev1.TaskTemplate:
				ttMPERef = tt.Spec.MediaProcessingEntityRef
				ttMPESel = tt.Spec.MediaProcessingEntitySelector
			case *enginev1.ClusterTaskTemplate:
				ttMPERef = tt.Spec.MediaProcessingEntityRef
				ttMPESel = tt.Spec.MediaProcessingEntitySelector
			default:
				return ctrl.Result{}, errors.New("taskTemplateRef does not reference a TaskTemplate or ClusterTaskTemplate")
			}

			// 3.1. MediaProcessingEntityRef
			if ttMPERef != nil {
				ref, err := utils.LocalMediaProcessingEntityToObjectRef(ttMPERef, task.Namespace)
				if err != nil {
					return ctrl.Result{}, err
				}
				if err = utils.NormalizeMediaProcessingEntityRef(r.Scheme, ref); err != nil {
					return ctrl.Result{}, err
				}
				task.Status.MediaProcessingEntityRef = ref
				return ctrl.Result{}, nil
			}

			// 3.2. MediaProcessingEntitySelector
			if ttMPESel != nil {
				sel, err := metav1.LabelSelectorAsSelector(ttMPESel)
				if err != nil {
					return ctrl.Result{}, err
				}

				ref, err := utils.SelectMediaProcessingEntityRef(ctx, r.Client, task.Namespace, sel)
				if err != nil {
					return ctrl.Result{}, err
				}
				if ref != nil {
					if err = utils.NormalizeMediaProcessingEntityRef(r.Scheme, ref); err != nil {
						return ctrl.Result{}, err
					}
					task.Status.MediaProcessingEntityRef = ref
					return ctrl.Result{}, nil
				}
			}
		}
	}

	// 4. default MediaProcessingEntity
	mpeObj, err := utils.GetAnnotatedObject(ctx, r.Client, &enginev1.MediaProcessingEntity{},
		enginev1.BetaIsDefaultStepMediaLocationAnnotation, "true", client.InNamespace(task.Namespace))
	if !apierrors.IsNotFound(err) {
		if err != nil {
			return ctrl.Result{}, err
		}
		mpe, ok := mpeObj.(*enginev1.MediaProcessingEntity)
		if !ok {
			return ctrl.Result{}, errors.New("unexpected object")
		}
		task.Status.MediaProcessingEntityRef = utils.ToRef(mpe)
	}

	// 5. default ClusterMediaProcessingEntity
	cmpeObj, err := utils.GetAnnotatedObject(ctx, r.Client, &enginev1.ClusterMediaProcessingEntity{},
		enginev1.BetaIsDefaultStepMediaLocationAnnotation, "true")
	if !apierrors.IsNotFound(err) {
		if err != nil {
			return ctrl.Result{}, err
		}
		mpe, ok := cmpeObj.(*enginev1.ClusterMediaProcessingEntity)
		if !ok {
			return ctrl.Result{}, errors.New("unexpected object")
		}
		task.Status.MediaProcessingEntityRef = utils.ToRef(mpe)
	}

	// could not find a MediaProcessingEntity or ClusterMediaProcessingEntity
	return ctrl.Result{}, errors.New("failed to reconcile MediaProcessingEntity")
}

func (r *TaskReconciler) reconcileFunction(ctx context.Context, task *enginev1.Task) (ctrl.Result, error) {
	// a Function can be selected through various ways with different precedence:

	// 1. FunctionRef
	if task.Spec.FunctionRef != nil {
		ref, err := utils.LocalFunctionToObjectRef(task.Spec.FunctionRef, task.Namespace)
		if err != nil {
			return ctrl.Result{}, err
		}
		if err = utils.NormalizeFunctionRef(r.Scheme, ref); err != nil {
			return ctrl.Result{}, err
		}
		task.Status.FunctionRef = ref
		return ctrl.Result{}, nil
	}

	// 2. FunctionSelector
	if task.Spec.FunctionSelector != nil {
		sel, err := metav1.LabelSelectorAsSelector(task.Spec.FunctionSelector)
		if err != nil {
			return ctrl.Result{}, err
		}

		ref, err := utils.SelectFunctionRef(ctx, r.Client, task.Namespace, sel)
		if err != nil {
			return ctrl.Result{}, err
		}
		if ref != nil {
			if err = utils.NormalizeFunctionRef(r.Scheme, ref); err != nil {
				return ctrl.Result{}, err
			}
			task.Status.FunctionRef = ref
			return ctrl.Result{}, nil
		}
	}

	// 3. TaskTemplateRef
	if task.Spec.TaskTemplateRef != nil {
		// shallow copy
		ttRef := task.Spec.TaskTemplateRef
		if err := utils.NormalizeLocalTaskTemplateRef(r.Scheme, ttRef); err != nil {
			return ctrl.Result{}, err
		}

		ttObj, err := utils.ResolveLocalRef(ctx, r.Client, task.Namespace, task.Spec.TaskTemplateRef)
		if !apierrors.IsNotFound(err) {
			if err != nil {
				return ctrl.Result{}, err
			}

			var ttFuncRef *meta.LocalObjectReference
			var ttFuncSel *metav1.LabelSelector

			switch tt := ttObj.(type) {
			case *enginev1.TaskTemplate:
				ttFuncRef = tt.Spec.FunctionRef
				ttFuncSel = tt.Spec.FunctionSelector
			case *enginev1.ClusterTaskTemplate:
				ttFuncRef = tt.Spec.FunctionRef
				ttFuncSel = tt.Spec.FunctionSelector
			default:
				return ctrl.Result{}, errors.New("taskTemplateRef does not reference a TaskTemplate or ClusterTaskTemplate")
			}

			// 3.1. FunctionRef
			if ttFuncRef != nil {
				ref, err := utils.LocalFunctionToObjectRef(ttFuncRef, task.Namespace)
				if err != nil {
					return ctrl.Result{}, err
				}
				if err = utils.NormalizeFunctionRef(r.Scheme, ref); err != nil {
					return ctrl.Result{}, err
				}
				task.Status.FunctionRef = ref
				return ctrl.Result{}, nil
			}

			// 3.2. FunctionSelector
			if ttFuncSel != nil {
				sel, err := metav1.LabelSelectorAsSelector(ttFuncSel)
				if err != nil {
					return ctrl.Result{}, err
				}

				ref, err := utils.SelectFunctionRef(ctx, r.Client, task.Namespace, sel)
				if err != nil {
					return ctrl.Result{}, err
				}
				if ref != nil {
					if err = utils.NormalizeFunctionRef(r.Scheme, ref); err != nil {
						return ctrl.Result{}, err
					}
					task.Status.FunctionRef = ref
					return ctrl.Result{}, nil
				}
			}
		}
	}

	// could not find a Function or ClusterFunction
	return ctrl.Result{}, errors.New("failed to reconcile Function")
}

func (r *TaskReconciler) reconcilePendingJob(ctx context.Context, task *enginev1.Task) (ctrl.Result, error) {
	// TODO: implement
	return ctrl.Result{}, errors.New("not implemented")
}

func (r *TaskReconciler) reconcileRunningJob(ctx context.Context, task *enginev1.Task) (ctrl.Result, error) {
	log := logf.FromContext(ctx)

	if task.Status.JobRef == nil {
		// wrong phase: go back
		task.Status.Phase = enginev1.TaskPhaseJobPending
		return ctrl.Result{}, nil
	}

	// resolve job
	job, err := r.resolveJobRef(ctx, task)
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			// job no longer exists: go back
			task.Status.Phase = enginev1.TaskPhaseJobPending
			return ctrl.Result{}, nil
		case apierrors.IsConflict(err):
			// conflicting job reference: log and adopt this job
			log.Error(err, "conflicting job reference: adopting Job as new reference")
			task.Status.JobRef = utils.ToExactRef(job)
		default:
			return ctrl.Result{}, err
		}
	}

	// check job status
	if utils.JobIsActive(job) {
		// still active: nothing to do
		return ctrl.Result{}, nil
	}

	// job has terminated
	if utils.JobWasSuccessful(job) {
		task.Status.Phase = enginev1.TaskPhaseSucceeded
	} else {
		task.Status.Phase = enginev1.TaskPhaseFailed
	}

	if task.Status.EndTime == nil {
		task.Status.EndTime = &metav1.Time{Time: time.Now()}
	}

	return ctrl.Result{}, nil
}

func (r *TaskReconciler) resolveJobRef(ctx context.Context, task *enginev1.Task) (*batchv1.Job, error) {
	// get correct API client
	c, err := r.getJobClientForTask(task)
	if err != nil {
		return nil, err
	}

	// resolve job
	jobObj, err := utils.ResolveExactRef(ctx, c, task.Status.JobRef)
	if err != nil {
		return nil, err
	}

	job, ok := jobObj.(*batchv1.Job)
	if !ok {
		return nil, errors.New("jobRef does not reference a Job")
	}

	return job, nil
}

func (r *TaskReconciler) getJobClientForTask(task *enginev1.Task) (Client, error) {
	c, ok := r.MediaProcessingEntityReconciler.GetClient(task.Status.MediaProcessingEntityRef)
	if !ok {
		return nil, errors.New("MediaProcessingEntity does not exist or is not ready")
	}
	return c, nil
}

func (r *TaskReconciler) normalizeStatusReferences(task *enginev1.Task) error {
	if task.Status.MediaProcessingEntityRef != nil {
		if err := utils.NormalizeMediaProcessingEntityRef(r.Scheme, task.Status.MediaProcessingEntityRef); err != nil {
			return err
		}
	}

	if task.Status.FunctionRef != nil {
		if err := utils.NormalizeFunctionRef(r.Scheme, task.Status.FunctionRef); err != nil {
			return err
		}
	}

	if task.Status.JobRef != nil {
		if err := utils.NormalizeExactRef(r.Scheme, task.Status.JobRef, &batchv1.Job{}); err != nil {
			return err
		}
	}

	return nil
}

// TODO: emit Kubernetes events
// SetupWithManager sets up the controller with the Manager.
func (r *TaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		Named(TaskControllerName).
		For(&enginev1.Task{}).
		Watches(&source.Kind{Type: &enginev1.Workflow{}}, handler.EnqueueRequestsFromMapFunc(r.mapWorkflowToTaskRequests)).
		Watches(&source.Channel{Source: r.JobEventChannel}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}

func (r *TaskReconciler) mapWorkflowToTaskRequests(wf client.Object) []reconcile.Request {
	ctx := context.Background()

	taskList := &enginev1.TaskList{}
	err := r.List(ctx, taskList, client.MatchingLabels{
		enginev1.WorkflowNamespaceLabel: wf.GetNamespace(),
		enginev1.WorkflowNameLabel:      wf.GetName(),
	})
	if err != nil {
		return nil
	}

	req := make([]reconcile.Request, len(taskList.Items))
	for i, task := range taskList.Items {
		req[i] = reconcile.Request{NamespacedName: client.ObjectKeyFromObject(&task)}
	}
	return req
}
