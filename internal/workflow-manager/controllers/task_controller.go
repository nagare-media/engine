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
	"encoding/json"
	"errors"
	"fmt"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/utils/ptr"
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
	apiclient "github.com/nagare-media/engine/internal/workflow-manager/client"
	"github.com/nagare-media/engine/pkg/apis/functions"
	"github.com/nagare-media/engine/pkg/apis/meta"
)

const (
	TaskControllerName = "nagare-media-engine-task-controller"
)

// TaskReconciler reconciles a Task object
type TaskReconciler struct {
	client.Client
	APIReader         client.Reader
	readOnlyAPIClient client.Client

	Config                          enginev1.WorkflowManagerConfiguration
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
// +kubebuilder:rbac:groups=engine.nagare.media,resources=medialocations,verbs=get;list;watch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=clustermedialocations,verbs=get;list;watch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=tasktemplates,verbs=get;list;watch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=clustertasktemplates,verbs=get;list;watch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=workflows,verbs=get;list;watch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=clusterworkflows,verbs=get;list;watch
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get

func (r *TaskReconciler) Reconcile(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := logf.FromContext(ctx)

	// fetch Task
	task := &enginev1.Task{}
	// TODO: There is a race between channel watch events from the Job controller and the Task informer. Reading from
	// Client could return stale data. APIReader will read directly from the API server but does not use the cache. This
	// should be synced in process.
	if err := r.APIReader.Get(ctx, req.NamespacedName, task); err != nil {
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
			if err := c.Delete(ctx, job, client.Preconditions{UID: &task.Status.JobRef.UID}, client.PropagationPolicy(metav1.DeletePropagationBackground)); err != nil {
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
		Controller:         ptr.To[bool](true),
		BlockOwnerDeletion: ptr.To[bool](true),
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
		enginev1.BetaIsDefaultMediaProcessingEntityAnnotation, "true", client.InNamespace(task.Namespace))
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
		enginev1.BetaIsDefaultMediaProcessingEntityAnnotation, "true")
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
		// TODO: sort results by function version number
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
	log := logf.FromContext(ctx)

	if task.Status.MediaProcessingEntityRef == nil || task.Status.FunctionRef == nil {
		// wrong phase: go back
		task.Status.Phase = enginev1.TaskPhaseInitializing
		return ctrl.Result{}, nil
	}

	// TODO: it would probably be better to use labels + list instead of relying on JobRef.

	if task.Status.JobRef != nil {
		// wrong phase: go forward
		task.Status.Phase = enginev1.TaskPhaseRunning
		return ctrl.Result{}, nil
	}

	// fetch Workflow
	wf := &enginev1.Workflow{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: task.Namespace, Name: task.Spec.WorkflowRef.Name}, wf); err != nil {
		return ctrl.Result{}, err
	}

	// resolve Function
	var funcSpec enginev1.FunctionSpec
	funcObj, err := utils.ResolveRef(ctx, r.Client, task.Status.FunctionRef)
	if err != nil {
		return ctrl.Result{}, err
	}
	switch f := funcObj.(type) {
	case *enginev1.ClusterFunction:
		funcSpec = f.Spec
	case *enginev1.Function:
		funcSpec = f.Spec
	default:
		return ctrl.Result{}, errors.New("functionRef does not reference a Function or ClusterFunction")
	}

	// resolve TaskTemplate
	var ttSpec *enginev1.TaskTemplateSpec
	if task.Spec.TaskTemplateRef != nil {
		ttObj, err := utils.ResolveLocalRef(ctx, r.Client, task.Namespace, task.Spec.TaskTemplateRef)
		if err != nil {
			return ctrl.Result{}, err
		}
		switch tt := ttObj.(type) {
		case *enginev1.ClusterTaskTemplate:
			ttSpec = &tt.Spec
		case *enginev1.TaskTemplate:
			ttSpec = &tt.Spec
		default:
			return ctrl.Result{}, errors.New("taskTemplateRef does not reference a TaskTemplate or ClusterTaskTemplate")
		}
	}

	// get correct API client
	jobClient, err := r.getJobClientForTask(task)
	if err != nil {
		return ctrl.Result{}, err
	}

	// check if function only runs on local MPEs
	if funcSpec.LocalMediaProcessingEntitiesOnly && jobClient.IsRemote() {
		return ctrl.Result{}, errors.New("localMediaProcessingEntitiesOnly is true, but selected MediaProcessingEntity is remote")
	}

	// construct Secret
	data := &functions.SecretData{
		Workflow: functions.Workflow{
			Info: functions.WorkflowInfo{
				Name: wf.Name,
			},
		},
		Task: functions.Task{
			Info: functions.TaskInfo{
				Name: task.Name,
			},
		},
		System: functions.System{
			NATS: r.Config.NATS,
		},
	}

	// add Workflow configuration
	wfCfgMap := make(map[string]any)
	if wf.Spec.Config != nil {
		if err = json.Unmarshal(wf.Spec.Config.Raw, &wfCfgMap); err != nil {
			return ctrl.Result{}, err
		}
	}
	data.Workflow.Config = wfCfgMap

	// add Task configuration (merge of Function, TaskTemplate and Task configuration)
	taskCfgMap := make(map[string]any)
	taskCfgPatches := make([]*apiextensionsv1.JSON, 0, 3)
	taskCfgPatches = append(taskCfgPatches, funcSpec.DefaultConfig)
	if ttSpec != nil {
		taskCfgPatches = append(taskCfgPatches, ttSpec.Config)
	}
	taskCfgPatches = append(taskCfgPatches, task.Spec.Config)
	for _, patch := range taskCfgPatches {
		if patch != nil {
			newTaskCfgMap := make(map[string]any)
			if err = utils.StrategicMerge(taskCfgMap, patch, &newTaskCfgMap); err != nil {
				return ctrl.Result{}, err
			}
			taskCfgMap = newTaskCfgMap
		}
	}
	data.Task.Config = taskCfgMap

	// add MediaLocations
	// TODO: add default MLs
	data.MediaLocations = make(map[string]enginev1.MediaLocationConfig, len(wf.Spec.MediaLocations))
	for _, mlRef := range wf.Spec.MediaLocations {
		mlObj, err := utils.ResolveLocalRef(ctx, r.Client, task.Namespace, &mlRef.Ref)
		if err != nil {
			return ctrl.Result{}, err
		}

		var mlCfg *enginev1.MediaLocationConfig
		switch ml := mlObj.(type) {
		case *enginev1.ClusterMediaLocation:
			mlCfg = &ml.Spec.MediaLocationConfig
		case *enginev1.MediaLocation:
			mlCfg = &ml.Spec.MediaLocationConfig
		default:
			return ctrl.Result{}, errors.New("mediaProcessingEntityRef does not reference a MediaProcessingEntity or ClusterMediaProcessingEntity")
		}

		// resolve MediaLocation secrets
		switch {
		case mlCfg.HTTP != nil:
			if mlCfg.HTTP.Auth != nil {
				if mlCfg.HTTP.Auth.Basic != nil {
					if err = r.prepareSecretRefForSecretData(ctx, &mlCfg.HTTP.Auth.Basic.SecretRef, task); err != nil {
						return ctrl.Result{}, err
					}
				}
				if mlCfg.HTTP.Auth.Digest != nil {
					if err = r.prepareSecretRefForSecretData(ctx, &mlCfg.HTTP.Auth.Digest.SecretRef, task); err != nil {
						return ctrl.Result{}, err
					}
				}
				if mlCfg.HTTP.Auth.Token != nil {
					if err = r.prepareSecretRefForSecretData(ctx, &mlCfg.HTTP.Auth.Token.SecretRef, task); err != nil {
						return ctrl.Result{}, err
					}
				}
			}
		case mlCfg.S3 != nil:
			if mlCfg.S3.Auth.AWS != nil {
				if err = r.prepareSecretRefForSecretData(ctx, &mlCfg.S3.Auth.AWS.SecretRef, task); err != nil {
					return ctrl.Result{}, err
				}
			}
		case mlCfg.Opencast != nil:
			if mlCfg.Opencast.Auth.Basic != nil {
				if err = r.prepareSecretRefForSecretData(ctx, &mlCfg.Opencast.Auth.Basic.SecretRef, task); err != nil {
					return ctrl.Result{}, err
				}
			}
		case mlCfg.RTMP != nil:
			if mlCfg.RTMP.Auth != nil {
				if mlCfg.RTMP.Auth.Basic != nil {
					if err = r.prepareSecretRefForSecretData(ctx, &mlCfg.RTMP.Auth.Basic.SecretRef, task); err != nil {
						return ctrl.Result{}, err
					}
				}
				if mlCfg.RTMP.Auth.StreamingKey != nil {
					if err = r.prepareSecretRefForSecretData(ctx, &mlCfg.RTMP.Auth.StreamingKey.SecretRef, task); err != nil {
						return ctrl.Result{}, err
					}
				}
			}
		case mlCfg.RTSP != nil:
			if mlCfg.RTSP.Auth != nil {
				if mlCfg.RTSP.Auth.Basic != nil {
					if err = r.prepareSecretRefForSecretData(ctx, &mlCfg.RTSP.Auth.Basic.SecretRef, task); err != nil {
						return ctrl.Result{}, err
					}
				}
			}
		case mlCfg.RIST != nil:
			if mlCfg.RIST.Encryption != nil {
				if err = r.prepareSecretRefForSecretData(ctx, &mlCfg.RIST.Encryption.SecretRef, task); err != nil {
					return ctrl.Result{}, err
				}
			}
		}

		data.MediaLocations[mlRef.Name] = *mlCfg
	}

	// create Secret
	jsonData, err := json.Marshal(data)
	if err != nil {
		return ctrl.Result{}, err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			// TODO: should we use fix names with collision detection?
			GenerateName: task.Name + "-",
			Namespace:    jobClient.Namespace(),
			Labels: map[string]string{
				enginev1.WorkflowNamespaceLabel: wf.Namespace,
				enginev1.WorkflowNameLabel:      wf.Name,
				enginev1.TaskNamespaceLabel:     task.Namespace,
				enginev1.TaskNameLabel:          task.Name,
			},
		},
		Data: map[string][]byte{
			"data.json": jsonData,
		},
	}
	if err = jobClient.Create(ctx, secret); err != nil {
		return ctrl.Result{}, err
	}

	// construct Job
	jobTemplate := funcSpec.Template.DeepCopy()

	// patch Job with change from TaskTemplate and Task
	jobTemplatePatches := make([]*batchv1.JobTemplateSpec, 0, 2)
	if ttSpec != nil {
		jobTemplatePatches = append(jobTemplatePatches, ttSpec.TemplatePatches)
	}
	jobTemplatePatches = append(jobTemplatePatches, task.Spec.TemplatePatches)
	for _, patch := range jobTemplatePatches {
		if patch != nil {
			newJobTemplate := &batchv1.JobTemplateSpec{}
			if err = utils.StrategicMerge(jobTemplate, patch, newJobTemplate); err != nil {
				return ctrl.Result{}, err
			}
			jobTemplate = newJobTemplate
		}
	}

	// add Job labels
	if jobTemplate.ObjectMeta.Labels == nil {
		jobTemplate.ObjectMeta.Labels = make(map[string]string)
	}
	jobTemplate.ObjectMeta.Labels[enginev1.WorkflowNamespaceLabel] = wf.Namespace
	jobTemplate.ObjectMeta.Labels[enginev1.WorkflowNameLabel] = wf.Name
	jobTemplate.ObjectMeta.Labels[enginev1.TaskNamespaceLabel] = task.Namespace
	jobTemplate.ObjectMeta.Labels[enginev1.TaskNameLabel] = task.Name

	// add data volume
	jobTemplate.Spec.Template.Spec.Volumes = append(jobTemplate.Spec.Template.Spec.Volumes, corev1.Volume{
		Name: "nagare-media-engine-task-data",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: secret.Name,
			},
		},
	})
	for _, containers := range [][]corev1.Container{jobTemplate.Spec.Template.Spec.InitContainers, jobTemplate.Spec.Template.Spec.Containers} {
		for i := range containers {
			c := &containers[i]
			c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
				Name:      "nagare-media-engine-task-data",
				MountPath: "/run/secrets/engine.nagare.media/task",
				ReadOnly:  true,
			})
		}
	}

	// set misc Job properties
	jobTemplate.Spec.TTLSecondsAfterFinished = ptr.To[int32](24 * 60 * 60) // = 1 day // TODO: adopt as config

	// create Job
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			// TODO: should we use fix names with collision detection?
			GenerateName: task.Name + "-",
			Namespace:    jobClient.Namespace(),
			Labels:       jobTemplate.ObjectMeta.Labels,
			Annotations:  jobTemplate.ObjectMeta.Annotations,
			Finalizers: []string{
				enginev1.JobProtectionFinalizer,
			},
		},
		Spec: jobTemplate.Spec,
	}
	if err = jobClient.Create(ctx, job); err != nil {
		errSec := jobClient.Delete(ctx, secret, client.Preconditions{UID: &secret.UID})
		if errSec != nil && !apierrors.IsNotFound(errSec) && !apierrors.IsConflict(errSec) {
			err = kerrors.NewAggregate([]error{err, errSec})
		}
		return ctrl.Result{}, err
	}

	// set Job reference
	task.Status.JobRef = utils.ToExactRef(job)

	// set Secret owner to Job
	secret.OwnerReferences = utils.EnsureOwnerRef(secret.OwnerReferences, metav1.OwnerReference{
		APIVersion: batchv1.SchemeGroupVersion.String(),
		Kind:       "Job",
		Name:       job.Name,
		UID:        job.UID,
		// Don't set controller as secret is not managed by job controller. Kubernetes will take care of garbage collection
		// automatically.
	})
	if err = jobClient.Update(ctx, secret); err != nil {
		// we have to ignore this as this could lead to the creation of duplicate Jobs
		log.Error(err, "failed to set owner reference on secret: ignoring leaked secret")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
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
			task.Status.JobRef = nil
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

func (r *TaskReconciler) prepareSecretRefForSecretData(ctx context.Context, ref *meta.ConfigMapOrSecretReference, task *enginev1.Task) error {
	ref.Namespace = task.Namespace
	if err := utils.ResolveSecretRefInline(ctx, r.readOnlyAPIClient, ref); err != nil {
		return err
	}
	ref.SetMarshalOnlyData(true)
	return nil
}

// TODO: emit Kubernetes events
// SetupWithManager sets up the controller with the Manager.
func (r *TaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// create read-only client
	r.readOnlyAPIClient = apiclient.NewReadOnlyClient(r.APIReader, r.Scheme, r.Client.RESTMapper())

	return ctrl.NewControllerManagedBy(mgr).
		Named(TaskControllerName).
		For(&enginev1.Task{}).
		Watches(&enginev1.Workflow{}, handler.EnqueueRequestsFromMapFunc(r.mapWorkflowToTaskRequests)).
		WatchesRawSource(&source.Channel{Source: r.JobEventChannel}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}

func (r *TaskReconciler) mapWorkflowToTaskRequests(ctx context.Context, wf client.Object) []reconcile.Request {
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
