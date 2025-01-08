/*
Copyright 2022-2025 The nagare media authors

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
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/gofiber/fiber/v2/log"
	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/internal/pkg/apis/utils"
	apiclient "github.com/nagare-media/engine/internal/workflow-manager/client"
	"github.com/nagare-media/engine/pkg/apis/meta"
	"github.com/nagare-media/engine/pkg/engineurl"
	"github.com/nagare-media/engine/pkg/maps"
	"github.com/nagare-media/models.go/base"
)

const (
	TaskControllerName = "nagare-media-engine-task-controller"
)

const (
	internalInitTaskPhase   = enginev1.TaskPhase("InternalInit")
	internalDeleteTaskPhase = enginev1.TaskPhase("InternalDelete")
	internalNormalTaskPhase = enginev1.TaskPhase("InternalAlways")
)

const (
	WorkflowManagerHelperDataVolume          = "nagare-media-engine-workflow-manager-helper-data"
	WorkflowManagerHelperDataVolumeMountPath = "/run/secrets/engine.nagare.media/task"
	WorkflowManagerHelperDataFileName        = "data.json"
)

const (
	ResourcePrefix           = "task-"
	HTTPStreamTargetPortName = "stream-http"
)

// TaskReconciler reconciles a Task object
type TaskReconciler struct {
	client.Client
	APIReader         client.Reader
	readOnlyAPIClient client.Client
	serializer        *json.Serializer

	Config                          *enginev1.WorkflowManagerConfig
	Scheme                          *runtime.Scheme
	JobEventChannel                 <-chan event.GenericEvent
	MediaProcessingEntityReconciler *MediaProcessingEntityReconciler

	ensureFuncs map[enginev1.TaskPhase][]EnsureFunc[*enginev1.Task]
}

// +kubebuilder:rbac:groups="",resources=configmaps,verbs=get
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=engine.nagare.media,resources=clusterfunctions,verbs=get;list;watch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=clustermedialocations,verbs=get;list;watch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=clustermediaprocessingentities,verbs=get;list;watch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=clustertasktemplates,verbs=get;list;watch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=clusterworkflows,verbs=get;list;watch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=functions,verbs=get;list;watch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=medialocations,verbs=get;list;watch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=mediaprocessingentities,verbs=get;list;watch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=tasks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=engine.nagare.media,resources=tasks/finalizers,verbs=update
// +kubebuilder:rbac:groups=engine.nagare.media,resources=tasks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=tasktemplates,verbs=get;list;watch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=workflows,verbs=get;list;watch

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
		_, err := utils.FullPatch(ctx, r.Client, task, oldTask)
		apiErrs := kerrors.FilterOut(err, apierrors.IsNotFound)
		if apiErrs != nil {
			log.Error(apiErrs, "error patching Task")
		}
		reterr = kerrors.Flatten(kerrors.NewAggregate([]error{apiErrs, reterr}))
	}()

	// always apply internal init phase functions
	if res, err := ApplyEnsureFuncs(ctx, task, r.ensureFuncs[internalInitTaskPhase]); err != nil || res.Stop {
		return res.ReconcileResult(), err
	}

	// handle delete
	if utils.IsInDeletion(task) {
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

	case enginev1.InitializingTaskPhase:
		task.Status.Conditions = utils.MarkConditionFalse(task.Status.Conditions, enginev1.TaskInitializedConditionType)

	case enginev1.JobPendingTaskPhase:
		task.Status.Conditions = utils.MarkCondition(task.Status.Conditions, enginev1.TaskInitializedConditionType, phaseConditionStatus)

	case enginev1.RunningTaskPhase:
		task.Status.Conditions = utils.MarkConditionTrue(task.Status.Conditions, enginev1.TaskInitializedConditionType)
		task.Status.Conditions = utils.MarkCondition(task.Status.Conditions, enginev1.TaskReadyConditionType, phaseConditionStatus)

	case enginev1.SucceededTaskPhase:
		task.Status.Conditions = utils.MarkConditionTrue(task.Status.Conditions, enginev1.TaskInitializedConditionType)
		task.Status.Conditions = utils.MarkConditionFalse(task.Status.Conditions, enginev1.TaskReadyConditionType)
		task.Status.Conditions = utils.MarkConditionTrue(task.Status.Conditions, enginev1.TaskCompleteConditionType)

	case enginev1.FailedTaskPhase:
		task.Status.Conditions = utils.MarkConditionTrue(task.Status.Conditions, enginev1.TaskInitializedConditionType)
		task.Status.Conditions = utils.MarkConditionFalse(task.Status.Conditions, enginev1.TaskReadyConditionType)
		task.Status.Conditions = utils.MarkConditionTrue(task.Status.Conditions, enginev1.TaskFailedConditionType)
	}
}

func (r *TaskReconciler) reconcileDelete(ctx context.Context, task *enginev1.Task) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info("reconcile deleted Task")

	// apply internal delete phase functions
	res, err := ApplyEnsureFuncs(ctx, task, r.ensureFuncs[internalDeleteTaskPhase])
	return res.ReconcileResult(), err
}

func (r *TaskReconciler) reconcile(ctx context.Context, task *enginev1.Task) (ctrl.Result, error) {
	log := logf.FromContext(ctx, "phase", task.Status.Phase)
	log.Info("reconcile Task")

	// always apply normal task phase functions
	if res, err := ApplyEnsureFuncs(ctx, task, r.ensureFuncs[internalNormalTaskPhase]); err != nil || res.Stop {
		return res.ReconcileResult(), err
	}

	// apply task phase functions
	res, err := ApplyEnsureFuncs(ctx, task, r.ensureFuncs[task.Status.Phase])
	return res.ReconcileResult(), err
}

func (r *TaskReconciler) ensureFinalizerIsSet(_ context.Context, task *enginev1.Task) (Result, error) {
	changed := controllerutil.AddFinalizer(task, enginev1.TaskProtectionFinalizer)
	return Result{Stop: changed}, nil
}

func (r *TaskReconciler) ensureFinalizerIsUnset(_ context.Context, task *enginev1.Task) (Result, error) {
	changed := controllerutil.RemoveFinalizer(task, enginev1.TaskProtectionFinalizer)
	return Result{Stop: changed}, nil
}

func (r *TaskReconciler) ensureStatusReferencesAreNormalized(_ context.Context, task *enginev1.Task) (Result, error) {
	if task.Status.MediaProcessingEntityRef != nil {
		if err := utils.NormalizeMediaProcessingEntityRef(r.Scheme, task.Status.MediaProcessingEntityRef); err != nil {
			return Result{}, err
		}
	}

	if task.Status.FunctionRef != nil {
		if err := utils.NormalizeFunctionRef(r.Scheme, task.Status.FunctionRef); err != nil {
			return Result{}, err
		}
	}

	if task.Status.JobRef != nil {
		if err := utils.NormalizeExactRef(r.Scheme, task.Status.JobRef, &batchv1.Job{}); err != nil {
			return Result{}, err
		}
	}

	return Result{}, nil
}

func (r *TaskReconciler) ensureQueuedTimeIsSet(_ context.Context, task *enginev1.Task) (Result, error) {
	if task.Status.QueuedTime == nil {
		task.Status.QueuedTime = &metav1.Time{Time: time.Now()}
	}
	return Result{}, nil
}

func (r *TaskReconciler) ensureStartTimeIsSet(_ context.Context, task *enginev1.Task) (Result, error) {
	if task.Status.StartTime == nil {
		task.Status.StartTime = &metav1.Time{Time: time.Now()}
	}
	return Result{}, nil
}

func (r *TaskReconciler) ensureEndTimeIsSet(_ context.Context, task *enginev1.Task) (Result, error) {
	if task.Status.EndTime == nil {
		task.Status.EndTime = &metav1.Time{Time: time.Now()}
	}
	return Result{}, nil
}

func (r *TaskReconciler) ensureTaskLifecycleIsSyncedToWorkflow(ctx context.Context, task *enginev1.Task) (Result, error) {
	res := Result{}

	// resolve Workflow
	wf, err := r.resolveWorkflowRef(ctx, task)
	if err != nil {
		return res, err
	}

	// set owner reference
	task.OwnerReferences = utils.EnsureOwnerRef(task.OwnerReferences, metav1.OwnerReference{
		APIVersion:         wf.GroupVersionKind().GroupVersion().String(),
		Kind:               wf.GroupVersionKind().Kind,
		Name:               wf.Name,
		UID:                wf.UID,
		Controller:         ptr.To(true),
		BlockOwnerDeletion: ptr.To(true),
	})

	// is Workflow marked for deletion
	if utils.IsInDeletion(wf) {
		res.Stop = true
		// we have to delete this Task ourselves as we use blockOwnerDeletion
		if err := r.Client.Delete(ctx, task, client.Preconditions{UID: &task.UID}); err != nil {
			return res, fmt.Errorf("failed to delete Task: %w", err)
		}
	}

	// set Workflow and Task label for filtering
	if task.Labels == nil {
		res.Stop = true
		task.Labels = make(map[string]string)
	}
	r.setCommonLabels(task.Labels, task)

	// check termination status of Workflow
	if utils.WorkflowHasTerminated(wf) && utils.TaskIsActive(task) {
		res.Stop = true
		// fail this Task since Workflow has already terminated
		// TODO: set reason
		task.Status.Phase = enginev1.FailedTaskPhase
	}

	return res, nil
}

func (r *TaskReconciler) ensurePhaseIsSetCorrectly(ctx context.Context, task *enginev1.Task) (Result, error) {
	res := Result{}

	switch task.Status.Phase {
	case enginev1.SucceededTaskPhase, enginev1.FailedTaskPhase:
		// Task has terminated -> no further checks
		return Result{}, nil

	case enginev1.RunningTaskPhase:
		if task.Status.JobRef == nil {
			task.Status.Phase = enginev1.JobPendingTaskPhase
			res.Stop = true
		}
		fallthrough

	case enginev1.JobPendingTaskPhase:
		if task.Status.MediaProcessingEntityRef == nil || task.Status.FunctionRef == nil {
			task.Status.Phase = enginev1.InitializingTaskPhase
			res.Stop = true
		}
		fallthrough

	case enginev1.InitializingTaskPhase:
		// no checks in initial phase

	default:
		// empty or unknown phase -> move to initializing
		task.Status.Phase = enginev1.InitializingTaskPhase
		res.Stop = true
	}

	return res, nil
}

func (r *TaskReconciler) ensurePhaseIsSetTo(p enginev1.TaskPhase) EnsureFunc[*enginev1.Task] {
	return func(_ context.Context, task *enginev1.Task) (Result, error) {
		changed := task.Status.Phase != p
		task.Status.Phase = p
		return Result{Stop: changed}, nil
	}
}

func (r *TaskReconciler) ensureMediaProcessingEntityIsResolved(ctx context.Context, task *enginev1.Task) (Result, error) {
	// We resolve the media processing entity only once. The relevant spec fields are read-only for existing objects and
	// we don't support changes that result from a new resolve. If changes need to be made, a new task should be created.
	if task.Status.MediaProcessingEntityRef != nil {
		return Result{}, nil
	}

	// A MediaProcessingEntity can be selected through various ways with different precedence:

	// 1. MediaProcessingEntityRef
	if task.Spec.MediaProcessingEntityRef != nil {
		ref, err := utils.LocalMediaProcessingEntityToObjectRef(task.Spec.MediaProcessingEntityRef, task.Namespace)
		if err != nil {
			return Result{}, err
		}
		if err = utils.NormalizeMediaProcessingEntityRef(r.Scheme, ref); err != nil {
			return Result{}, err
		}
		task.Status.MediaProcessingEntityRef = ref
		return Result{}, nil
	}

	// 2. MediaProcessingEntitySelector
	if task.Spec.MediaProcessingEntitySelector != nil {
		sel, err := metav1.LabelSelectorAsSelector(task.Spec.MediaProcessingEntitySelector)
		if err != nil {
			return Result{}, err
		}

		ref, err := utils.SelectMediaProcessingEntityRef(ctx, r.Client, task.Namespace, sel)
		if err != nil {
			return Result{}, err
		}
		if ref != nil {
			if err = utils.NormalizeMediaProcessingEntityRef(r.Scheme, ref); err != nil {
				return Result{}, err
			}
			task.Status.MediaProcessingEntityRef = ref
			return Result{}, nil
		}
	}

	// 3. TaskTemplateRef
	if task.Spec.TaskTemplateRef != nil {
		// shallow copy
		ttRef := task.Spec.TaskTemplateRef
		if err := utils.NormalizeLocalTaskTemplateRef(r.Scheme, ttRef); err != nil {
			return Result{}, err
		}

		// resolve TaskTemplate
		ttObj, err := utils.ResolveLocalRef(ctx, r.Client, task.Namespace, task.Spec.TaskTemplateRef)
		if !apierrors.IsNotFound(err) {
			if err != nil {
				return Result{}, err
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
				return Result{}, errors.New("taskTemplateRef does not reference a TaskTemplate or ClusterTaskTemplate")
			}

			// 3.1. MediaProcessingEntityRef
			if ttMPERef != nil {
				ref, err := utils.LocalMediaProcessingEntityToObjectRef(ttMPERef, task.Namespace)
				if err != nil {
					return Result{}, err
				}
				if err = utils.NormalizeMediaProcessingEntityRef(r.Scheme, ref); err != nil {
					return Result{}, err
				}
				task.Status.MediaProcessingEntityRef = ref
				return Result{}, nil
			}

			// 3.2. MediaProcessingEntitySelector
			if ttMPESel != nil {
				sel, err := metav1.LabelSelectorAsSelector(ttMPESel)
				if err != nil {
					return Result{}, err
				}

				ref, err := utils.SelectMediaProcessingEntityRef(ctx, r.Client, task.Namespace, sel)
				if err != nil {
					return Result{}, err
				}
				if ref != nil {
					if err = utils.NormalizeMediaProcessingEntityRef(r.Scheme, ref); err != nil {
						return Result{}, err
					}
					task.Status.MediaProcessingEntityRef = ref
					return Result{}, nil
				}
			}
		}
	}

	// 4. default MediaProcessingEntity
	mpeObj, err := utils.GetAnnotatedObject(ctx, r.Client, &enginev1.MediaProcessingEntity{},
		enginev1.IsDefaultMediaProcessingEntityAnnotation, "true", client.InNamespace(task.Namespace))
	if !apierrors.IsNotFound(err) {
		if err != nil {
			return Result{}, err
		}
		mpe, ok := mpeObj.(*enginev1.MediaProcessingEntity)
		if !ok {
			return Result{}, errors.New("unexpected object")
		}
		task.Status.MediaProcessingEntityRef = utils.ToRef(mpe)
		return Result{}, nil
	}

	// 5. default ClusterMediaProcessingEntity
	cmpeObj, err := utils.GetAnnotatedObject(ctx, r.Client, &enginev1.ClusterMediaProcessingEntity{},
		enginev1.IsDefaultMediaProcessingEntityAnnotation, "true")
	if !apierrors.IsNotFound(err) {
		if err != nil {
			return Result{}, err
		}
		mpe, ok := cmpeObj.(*enginev1.ClusterMediaProcessingEntity)
		if !ok {
			return Result{}, errors.New("unexpected object")
		}
		task.Status.MediaProcessingEntityRef = utils.ToRef(mpe)
		return Result{}, nil
	}

	// could not find a MediaProcessingEntity or ClusterMediaProcessingEntity
	return Result{}, errors.New("failed to reconcile MediaProcessingEntity")
}

func (r *TaskReconciler) ensureFunctionIsResolved(ctx context.Context, task *enginev1.Task) (Result, error) {
	// We resolve the function only once. The relevant spec fields are read-only for existing objects and we don't support
	// changes that result from a new resolve. If changes need to be made, a new task should be created.
	if task.Status.FunctionRef != nil {
		return Result{}, nil
	}

	// A Function can be selected through various ways with different precedence:

	// 1. FunctionRef
	if task.Spec.FunctionRef != nil {
		ref, err := utils.LocalFunctionToObjectRef(task.Spec.FunctionRef, task.Namespace)
		if err != nil {
			return Result{}, err
		}
		if err = utils.NormalizeFunctionRef(r.Scheme, ref); err != nil {
			return Result{}, err
		}
		task.Status.FunctionRef = ref
		return Result{}, nil
	}

	// 2. FunctionSelector
	if task.Spec.FunctionSelector != nil {
		// TODO: sort results by function version number
		sel, err := metav1.LabelSelectorAsSelector(task.Spec.FunctionSelector)
		if err != nil {
			return Result{}, err
		}

		ref, err := utils.SelectFunctionRef(ctx, r.Client, task.Namespace, sel)
		if err != nil {
			return Result{}, err
		}
		if ref != nil {
			if err = utils.NormalizeFunctionRef(r.Scheme, ref); err != nil {
				return Result{}, err
			}
			task.Status.FunctionRef = ref
			return Result{}, nil
		}
	}

	// 3. TaskTemplateRef
	if task.Spec.TaskTemplateRef != nil {
		// shallow copy
		ttRef := task.Spec.TaskTemplateRef
		if err := utils.NormalizeLocalTaskTemplateRef(r.Scheme, ttRef); err != nil {
			return Result{}, err
		}

		ttObj, err := utils.ResolveLocalRef(ctx, r.Client, task.Namespace, task.Spec.TaskTemplateRef)
		if !apierrors.IsNotFound(err) {
			if err != nil {
				return Result{}, err
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
				return Result{}, errors.New("taskTemplateRef does not reference a TaskTemplate or ClusterTaskTemplate")
			}

			// 3.1. FunctionRef
			if ttFuncRef != nil {
				ref, err := utils.LocalFunctionToObjectRef(ttFuncRef, task.Namespace)
				if err != nil {
					return Result{}, err
				}
				if err = utils.NormalizeFunctionRef(r.Scheme, ref); err != nil {
					return Result{}, err
				}
				task.Status.FunctionRef = ref
				return Result{}, nil
			}

			// 3.2. FunctionSelector
			if ttFuncSel != nil {
				sel, err := metav1.LabelSelectorAsSelector(ttFuncSel)
				if err != nil {
					return Result{}, err
				}

				ref, err := utils.SelectFunctionRef(ctx, r.Client, task.Namespace, sel)
				if err != nil {
					return Result{}, err
				}
				if ref != nil {
					if err = utils.NormalizeFunctionRef(r.Scheme, ref); err != nil {
						return Result{}, err
					}
					task.Status.FunctionRef = ref
					return Result{}, nil
				}
			}
		}
	}

	// could not find a Function or ClusterFunction
	return Result{}, errors.New("failed to reconcile Function")
}

func (r *TaskReconciler) ensureJobExists(ctx context.Context, task *enginev1.Task) (Result, error) {
	// check if Job already exists
	job, err := r.resolveJobRef(ctx, task)
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			// Job no longer exists: force recreation
			task.Status.JobRef = nil
			return Result{Stop: true}, nil
		case apierrors.IsConflict(err):
			// conflicting Job reference: log and adopt this Job
			log.Error(err, "conflicting job reference: adopting Job as new reference")
			task.Status.JobRef = utils.ToExactRef(job)
			return Result{Stop: true}, nil
		}
		return Result{}, err
	}
	oldJob := job.DeepCopy()
	exists := (job != nil)

	// get MPE client for this task
	c, err := r.getMPEClientForTask(task)
	if err != nil {
		return Result{}, err
	}

	// resolve Function
	funcSpec, err := r.resolveFunctionRef(ctx, task)
	if err != nil {
		return Result{}, err
	}

	// resolve TaskTemplate
	ttSpec, err := r.resolveTaskTemplateFunctionRef(ctx, task)
	if err != nil {
		return Result{}, err
	}

	// construct JobTemplate
	jobTmpl := batchv1.JobTemplateSpec{}
	patches := make([]*batchv1.JobTemplateSpec, 0, 3)
	patches = append(patches, &funcSpec.Template)
	if ttSpec != nil {
		patches = append(patches, ttSpec.TemplatePatches)
	}
	patches = append(patches, task.Spec.TemplatePatches)
	err = utils.StrategicMergeList(&jobTmpl, patches...)
	if err != nil {
		return Result{}, err
	}

	// create new Job if necessary
	if !exists {
		jobName := utils.GenerateName(ResourcePrefix + task.Name)

		// check if function only runs on local MPEs
		// We only check for new Jobs as existing Jobs are already scheduled and cannot be moved.
		if funcSpec.LocalMediaProcessingEntitiesOnly && c.IsRemote() {
			return Result{}, errors.New("localMediaProcessingEntitiesOnly is true, but selected MediaProcessingEntity is remote")
		}

		// add volume for workflow-manager-helper data Secret
		jobTmpl.Spec.Template.Spec.Volumes = append(jobTmpl.Spec.Template.Spec.Volumes, corev1.Volume{
			Name: WorkflowManagerHelperDataVolume,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: jobName,
				},
			},
		})
		for _, containers := range [][]corev1.Container{jobTmpl.Spec.Template.Spec.InitContainers, jobTmpl.Spec.Template.Spec.Containers} {
			for i := range containers {
				c := &containers[i]
				c.VolumeMounts = append(c.VolumeMounts, corev1.VolumeMount{
					Name:      WorkflowManagerHelperDataVolume,
					MountPath: WorkflowManagerHelperDataVolumeMountPath,
					ReadOnly:  true,
				})
			}
		}

		// TODO: should we make sure, ports are defined on the containers?

		// set additional spec fields
		// TODO: make configurable
		jobTmpl.Spec.TTLSecondsAfterFinished = ptr.To[int32](24 * 60 * 60) // = 1 day

		// define Job
		job = &batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      jobName,
				Namespace: c.Namespace(),
			},
			Spec: jobTmpl.Spec,
		}
	}

	// Job labels
	job.Labels = maps.Merge(job.Labels, jobTmpl.Labels)
	r.setCommonLabels(job.Labels, task)

	// Pod labels
	if job.Spec.Template.Labels == nil {
		job.Spec.Template.Labels = make(map[string]string)
	}
	r.setCommonLabels(job.Spec.Template.Labels, task)

	// Job annotations
	job.Annotations = maps.Merge(job.Annotations, jobTmpl.Annotations)

	// Job finalizers
	controllerutil.AddFinalizer(job, enginev1.JobProtectionFinalizer)

	// create or patch Job
	if exists {
		_, err = utils.Patch(ctx, c, job, oldJob)
	} else {
		err = c.Create(ctx, job)
	}
	if err != nil {
		return Result{}, err
	}

	// set Job reference
	task.Status.JobRef = utils.ToExactRef(job)
	if err := utils.NormalizeExactRef(r.Scheme, task.Status.JobRef, &batchv1.Job{}); err != nil {
		return Result{}, err
	}

	return Result{}, nil
}

func (r *TaskReconciler) ensureJobServiceExists(ctx context.Context, task *enginev1.Task) (Result, error) {
	// get MPE client for this task
	c, err := r.getMPEClientForTask(task)
	if err != nil {
		return Result{}, err
	}

	// check if Service already exists
	svc := &corev1.Service{}
	svcName := ResourcePrefix + task.Name
	exists := true
	err = c.Get(ctx, client.ObjectKey{Namespace: c.Namespace(), Name: svcName}, svc)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return Result{}, err
		}
		exists = false
		svc = &corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      svcName,
				Namespace: c.Namespace(),
			},
		}
	}
	svcOld := svc.DeepCopy()

	// Service label
	if svc.Labels == nil {
		svc.Labels = make(map[string]string)
	}
	r.setCommonLabels(svc.Labels, task)

	// Service spec
	// TODO: the Service spec should be created form a ServiceTemplate similar to the JobTemplate
	svc.Spec = corev1.ServiceSpec{
		Type:     corev1.ServiceTypeClusterIP,
		Selector: make(map[string]string), // set below
		Ports: []corev1.ServicePort{{
			Name:        "http",
			Protocol:    corev1.ProtocolTCP,
			AppProtocol: ptr.To("http"),
			Port:        80,
			TargetPort:  intstr.FromString(HTTPStreamTargetPortName),
		}},
	}
	r.setCommonLabels(svc.Spec.Selector, task)

	// Service owner
	svc.OwnerReferences = utils.EnsureOwnerRef(svc.OwnerReferences, metav1.OwnerReference{
		APIVersion: task.Status.JobRef.APIVersion,
		Kind:       task.Status.JobRef.Kind,
		Name:       task.Status.JobRef.Name,
		UID:        task.Status.JobRef.UID,
		// Don't set controller=true as this resource is not managed by the Job controller.
	})

	if exists {
		_, err = utils.Patch(ctx, c, svc, svcOld)
	} else {
		err = c.Create(ctx, svc)
	}

	return Result{}, err
}

func (r *TaskReconciler) ensureWorkflowManagerHelperDataSecretExists(ctx context.Context, task *enginev1.Task) (Result, error) {
	log := logf.FromContext(ctx)

	// get MPE client for this task
	c, err := r.getMPEClientForTask(task)
	if err != nil {
		return Result{}, err
	}

	// check if Secret already exists
	secret := &corev1.Secret{}
	secretName := task.Status.JobRef.Name
	exists := true
	err = c.Get(ctx, client.ObjectKey{Namespace: c.Namespace(), Name: secretName}, secret)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return Result{}, err
		}
		exists = false
		secret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Name:      secretName,
				Namespace: c.Namespace(),
			},
		}
	}
	secretOld := secret.DeepCopy()

	// resolve Workflow
	wf, err := r.resolveWorkflowRef(ctx, task)
	if err != nil {
		return Result{}, err
	}

	// resolve Function
	funcSpec, err := r.resolveFunctionRef(ctx, task)
	if err != nil {
		return Result{}, err
	}

	// resolve TaskTemplate
	ttSpec, err := r.resolveTaskTemplateFunctionRef(ctx, task)
	if err != nil {
		return Result{}, err
	}

	// define WorkflowManagerHelperData
	data := enginev1.WorkflowManagerHelperData{
		WorkflowManagerHelperDataSpec: enginev1.WorkflowManagerHelperDataSpec{
			Function: enginev1.WorkflowManagerHelperDataFunction{
				Name:    task.Status.FunctionRef.Name,
				Version: funcSpec.Version,
			},
			Workflow: enginev1.WorkflowManagerHelperDataWorkflow{
				ID:            wf.Name,
				HumanReadable: wf.Spec.HumanReadable,
			},
			Task: enginev1.WorkflowManagerHelperDataTask{
				ID:            task.Name,
				HumanReadable: task.Spec.HumanReadable,
			},
			System: enginev1.WorkflowManagerHelperDataSystem{
				NATS: r.Config.NATS,
			},
		},
	}

	// resolve Input and Output streams
	data.Task.InputPorts, err = r.resolveInputPorts(ctx, task)
	if err != nil {
		return Result{}, err
	}
	data.Task.OutputPorts, err = r.resolveOutputPorts(ctx, task)
	if err != nil {
		return Result{}, err
	}

	// merge config
	configs := make([]map[string]string, 0, 3)
	configs = append(configs, funcSpec.DefaultConfig)
	if ttSpec != nil {
		configs = append(configs, ttSpec.Config)
	}
	configs = append(configs, task.Spec.Config)
	data.Task.Config = maps.Merge(configs...)

	// serialize data
	buf := bytes.Buffer{}
	if err = r.serializer.Encode(&data, &buf); err != nil {
		return Result{}, err
	}

	// define Secret
	if secret.Labels == nil {
		secret.Labels = make(map[string]string)
	}
	r.setCommonLabels(secret.Labels, task)

	secret.OwnerReferences = utils.EnsureOwnerRef(secret.OwnerReferences, metav1.OwnerReference{
		APIVersion: task.Status.JobRef.APIVersion,
		Kind:       task.Status.JobRef.Kind,
		Name:       task.Status.JobRef.Name,
		UID:        task.Status.JobRef.UID,
		// Don't set controller=true as this resource is not managed by the Job controller.
	})

	secret.Type = enginev1.WorkflowManagerHelperDataSecretType

	secret.StringData = nil
	secret.Data = map[string][]byte{
		WorkflowManagerHelperDataFileName: buf.Bytes(),
	}

	if exists {
		changed, err := utils.Patch(ctx, c, secret, secretOld)
		if changed {
			// It takes a sync loop interval to propagate Secret changes to Pods. This can be faster if we update the Pod
			// container and force a reconsiliation.
			if err2 := r.tryForcedJobPodsSync(ctx, task); err2 != nil {
				log.Error(err, "failed to force a Pod sync to speedup Secret propagation")
			}
		}
	} else {
		err = c.Create(ctx, secret)
	}

	return Result{}, err
}

func (r *TaskReconciler) ensureTaskLifecycleIsSyncedToJob(ctx context.Context, task *enginev1.Task) (Result, error) {
	log := logf.FromContext(ctx)

	// get Job
	job, err := r.resolveJobRef(ctx, task)
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			// Job no longer exists: force recreation
			task.Status.JobRef = nil
			return Result{Stop: true}, nil
		case apierrors.IsConflict(err):
			// conflicting Job reference: log and adopt this Job
			log.Error(err, "conflicting job reference: adopting Job as new reference")
			task.Status.JobRef = utils.ToExactRef(job)
			return Result{Stop: true}, nil
		}
		return Result{}, err
	}

	if job == nil {
		// no Job was ever created: force recreation
		task.Status.JobRef = nil
		return Result{Stop: true}, nil
	}

	// check Job status
	if utils.JobIsActive(job) {
		// still active: nothing to do
		return Result{}, nil
	}

	// Job has terminated
	if utils.JobWasSuccessful(job) {
		task.Status.Phase = enginev1.SucceededTaskPhase
	} else {
		task.Status.Phase = enginev1.FailedTaskPhase
	}
	return Result{}, nil
}

func (r *TaskReconciler) ensureJobHasTerminated(ctx context.Context, task *enginev1.Task) (Result, error) {
	log := logf.FromContext(ctx)

	// If task is being deleted, we need to delete the Job ourselves as we can't set owner references between clusters.
	taskIsInDeletion := utils.IsInDeletion(task)

	// resolve Job
	job, err := r.resolveJobRef(ctx, task)
	if err != nil {
		switch {
		case apierrors.IsNotFound(err):
			// already deleted
			return Result{}, nil
		case apierrors.IsConflict(err):
			// conflicting Job reference: log and adopt this Job
			log.Error(err, "conflicting job reference: adopting Job as new reference")
			task.Status.JobRef = utils.ToExactRef(job)
			return Result{Stop: true}, nil
		case taskIsInDeletion:
			// We could retry, but we assume the Job is well behaved and will stop executing eventually. Kubernetes should
			// cleanup after some time.
			log.Error(err, "ignoring dangling Job reference")
			return Result{}, nil
		}
		return Result{}, err
	}

	// no job was created
	if job == nil {
		return Result{}, err
	}

	// delete Job only if necessary
	// We want to keep the Job around if the Task still exist and it also already terminated. This allows to read logs
	// and see more details than on a Task.
	if taskIsInDeletion || utils.JobIsActive(job) {
		// get correct API client
		c, err := r.getMPEClientForTask(task)
		if err != nil {
			if taskIsInDeletion {
				// We could retry, but we assume the Job is well behaved and will stop executing eventually. Kubernetes should
				// cleanup after some time.
				log.Error(err, "ignoring dangling Job reference")
				return Result{}, nil
			}
			return Result{}, err
		}

		if err := c.Delete(ctx, job, client.Preconditions{UID: &task.Status.JobRef.UID}, client.PropagationPolicy(metav1.DeletePropagationForeground)); err != nil {
			switch {
			case apierrors.IsNotFound(err):
				// already deleted so nothing to do
				return Result{}, nil
			case taskIsInDeletion:
				// We could retry, but we assume the Job is well behaved and will stop executing eventually. Kubernetes should
				// cleanup after some time.
				log.Error(err, "ignoring dangling Job reference")
				return Result{}, nil
			}
			return Result{}, err
		}
	}

	return Result{}, nil
}

func (r *TaskReconciler) setCommonLabels(labels map[string]string, task *enginev1.Task) {
	labels[enginev1.WorkflowNamespaceLabel] = task.Namespace
	labels[enginev1.WorkflowNameLabel] = task.Spec.WorkflowRef.Name
	labels[enginev1.TaskNamespaceLabel] = task.Namespace
	labels[enginev1.TaskNameLabel] = task.Name
}

func (r *TaskReconciler) tryForcedJobPodsSync(ctx context.Context, task *enginev1.Task) error {
	job, err := r.resolveJobRef(ctx, task)
	if err != nil {
		return err
	}

	c, err := r.getMPEClientForTask(task)
	if err != nil {
		return err
	}

	sel, err := metav1.LabelSelectorAsSelector(job.Spec.Selector)
	if err != nil {
		return err
	}

	pods := &corev1.PodList{}
	err = c.List(ctx, pods, client.MatchingLabelsSelector{Selector: sel})
	if err != nil {
		return err
	}

	var errs []error

	for _, pod := range pods.Items {
		if pod.Status.Phase == corev1.PodRunning {
			oldPod := pod.DeepCopy()
			if pod.Annotations == nil {
				pod.Annotations = make(map[string]string)
			}
			pod.Annotations[enginev1.LastConfigChangePodAnnotation] = time.Now().UTC().Format(time.RFC3339)

			_, err = utils.Patch(ctx, c, &pod, oldPod)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}

	return kerrors.NewAggregate(errs)
}

func (r *TaskReconciler) resolveWorkflowRef(ctx context.Context, task *enginev1.Task) (*enginev1.Workflow, error) {
	wf := &enginev1.Workflow{}
	if err := r.Client.Get(ctx, client.ObjectKey{Namespace: task.Namespace, Name: task.Spec.WorkflowRef.Name}, wf); err != nil {
		return nil, fmt.Errorf("failed to fetch Workflow: %w", err)
	}
	return wf, nil
}

func (r *TaskReconciler) resolveTaskTemplateFunctionRef(ctx context.Context, task *enginev1.Task) (*enginev1.TaskTemplateSpec, error) {
	if task.Spec.TaskTemplateRef == nil {
		return nil, nil
	}

	ttObj, err := utils.ResolveLocalRef(ctx, r.Client, task.Namespace, task.Spec.TaskTemplateRef)
	if err != nil {
		return nil, err
	}

	var ttSpec enginev1.TaskTemplateSpec
	switch tt := ttObj.(type) {
	case *enginev1.ClusterTaskTemplate:
		ttSpec = tt.Spec
	case *enginev1.TaskTemplate:
		ttSpec = tt.Spec
	default:
		return nil, errors.New("taskTemplateRef does not reference a TaskTemplate or ClusterTaskTemplate")
	}

	return &ttSpec, nil
}

func (r *TaskReconciler) resolveFunctionRef(ctx context.Context, task *enginev1.Task) (*enginev1.FunctionSpec, error) {
	funcObj, err := utils.ResolveRef(ctx, r.Client, task.Status.FunctionRef)
	if err != nil {
		return nil, err
	}

	var funcSpec enginev1.FunctionSpec
	switch f := funcObj.(type) {
	case *enginev1.ClusterFunction:
		funcSpec = f.Spec
	case *enginev1.Function:
		funcSpec = f.Spec
	default:
		return nil, errors.New("functionRef does not reference a Function or ClusterFunction")
	}

	return &funcSpec, nil
}

func (r *TaskReconciler) resolveJobRef(ctx context.Context, task *enginev1.Task) (*batchv1.Job, error) {
	if task.Status.JobRef == nil {
		return nil, nil
	}

	// get correct API client
	c, err := r.getMPEClientForTask(task)
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

func (r *TaskReconciler) getMPEClientForTask(task *enginev1.Task) (Client, error) {
	c, ok := r.MediaProcessingEntityReconciler.GetClient(task.Status.MediaProcessingEntityRef)
	if !ok {
		return nil, errors.New("MediaProcessingEntity does not exist or is not ready")
	}
	return c, nil
}

func (r *TaskReconciler) resolveInputPorts(ctx context.Context, task *enginev1.Task) ([]enginev1.InputPortBinding, error) {
	rp := make([]enginev1.InputPortBinding, 0, len(task.Spec.InputPorts))
	for _, p := range task.Spec.InputPorts {
		m, err := r.resolveMediaStream(ctx, task, p.Input)
		if err != nil {
			return nil, err
		}
		rp = append(rp, enginev1.InputPortBinding{
			ID:    p.ID,
			Input: m,
		})
	}
	return rp, nil
}

func (r *TaskReconciler) resolveOutputPorts(ctx context.Context, task *enginev1.Task) ([]enginev1.OutputPortBinding, error) {
	rp := make([]enginev1.OutputPortBinding, 0, len(task.Spec.OutputPorts))
	for _, p := range task.Spec.OutputPorts {
		m, err := r.resolveMediaStream(ctx, task, p.Output)
		if err != nil {
			return nil, err
		}
		rp = append(rp, enginev1.OutputPortBinding{
			ID:     p.ID,
			Output: m,
		})
	}
	return rp, nil
}

func (r *TaskReconciler) resolveMediaStream(ctx context.Context, task *enginev1.Task, m *enginev1.Media) (*enginev1.Media, error) {
	if m == nil {
		return nil, nil
	}

	// URL must be set
	if m.URL == nil {
		return nil, fmt.Errorf("undefined URL for stream '%s'", m.ID)
	}

	// we only resolve nagare media engine URLs
	if !engineurl.IsEngineURL(string(*m.URL)) {
		return m.DeepCopy(), nil
	}

	parsedUrl, err := engineurl.Parse(string(*m.URL))
	if err != nil {
		return nil, err
	}

	var newUrl base.URI
	switch u := parsedUrl.(type) {
	case *engineurl.TaskURL:
		// Transform Task URL:
		//
		//    nagare-media-engine:///task/{WorkflowID}/{TaskID}/{PortName}/{StreamID}?{RawQuery}
		//    buffered://task-{TaskID}.{namespace}.svc.cluster.local/streams/{PortName}/{StreamID}?{RawQuery}

		// TODO: add support for non-default cluster-domains (i.e. don't assume .cluster.local)
		// TODO: add support for cross-namespace connections (i.e. don't assume task.Namespace is namespace of Task URL)
		// TODO: add support for cross-cluster connections (i.e. allow external connections through some gateway)
		// TODO: add support for arbitrary streaming protocols

		q, err := url.ParseQuery(u.RawQuery)
		if err != nil {
			return nil, err
		}
		q.Add(engineurl.BufferedProtocolQueryKey, "http")

		res := url.URL{
			Scheme:   "buffered",
			Host:     fmt.Sprintf("%s%s.%s.svc.cluster.local", ResourcePrefix, u.TaskID, task.Namespace),
			Path:     fmt.Sprintf("/streams/%s/%s", u.PortName, u.StreamID),
			RawQuery: q.Encode(),
		}
		newUrl = base.URI(res.String())

	case *engineurl.MediaLocationURL:
		// Transform MediaLocation URL
		//
		//    nagare-media-engine:///media/{Name}/{Path}?{RawQuery}

		// resolve MediaLocation
		ml, err := r.resolveMediaLocation(ctx, client.ObjectKey{Namespace: task.Namespace, Name: u.Name})
		if err != nil {
			return nil, err
		}

		// resolve MediaLocation secrets
		// TODO: implement
		// ref.Namespace = task.Namespace
		// if err := utils.ResolveSecretRefInline(ctx, r.readOnlyAPIClient, ref); err != nil {
		// 	return err
		// }
		// ref.SetMarshalOnlyData(true)
		// return nil

		// construct MediaLocation base URL
		mlURL, err := ml.URL()
		if err != nil {
			return nil, err
		}

		// add parsed path and query to base URL
		res := mlURL.JoinPath(u.Path)
		q := res.Query()
		q2, err := url.ParseQuery(u.RawQuery)
		if err != nil {
			return nil, err
		}
		for k, vl := range q2 {
			for _, v := range vl {
				q.Add(k, v)
			}
		}

		res.RawQuery = q.Encode()

		newUrl = base.URI(res.String())

	default:
		return nil, fmt.Errorf("unexpected nagare-media-engine URL type: %T", parsedUrl)
	}

	m2 := m.DeepCopy()
	m2.URL = &newUrl
	return m2, nil
}

func (r *TaskReconciler) resolveMediaLocation(ctx context.Context, key client.ObjectKey) (*enginev1.MediaLocationSpec, error) {
	// try MediaLocation
	ml := &enginev1.MediaLocation{}
	err := r.Client.Get(ctx, key, ml)
	if err == nil {
		return &ml.Spec, nil
	}
	if !apierrors.IsNotFound(err) {
		return nil, err
	}

	// try ClusterMediaLocation
	cml := &enginev1.ClusterMediaLocation{}
	err = r.Client.Get(ctx, key, cml)
	if err == nil {
		return &cml.Spec, nil
	}
	return nil, err
}

// TODO: set HumanReadable from task template if unset
// TODO: emit Kubernetes events
// SetupWithManager sets up the controller with the Manager.
func (r *TaskReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// init TaskReconciler
	r.readOnlyAPIClient = apiclient.NewReadOnlyClient(r.APIReader, r.Scheme, r.Client.RESTMapper())
	r.serializer = json.NewSerializerWithOptions(json.DefaultMetaFactory, r.Scheme, r.Scheme, json.SerializerOptions{})
	r.ensureFuncs = map[enginev1.TaskPhase][]EnsureFunc[*enginev1.Task]{
		internalInitTaskPhase: {
			r.ensureFinalizerIsSet,
			r.ensureStatusReferencesAreNormalized,
		},
		internalNormalTaskPhase: {
			r.ensureQueuedTimeIsSet,
			r.ensureTaskLifecycleIsSyncedToWorkflow,
			r.ensurePhaseIsSetCorrectly,
		},
		enginev1.InitializingTaskPhase: {
			r.ensureMediaProcessingEntityIsResolved,
			r.ensureFunctionIsResolved,
			r.ensurePhaseIsSetTo(enginev1.JobPendingTaskPhase),
		},
		enginev1.JobPendingTaskPhase: {
			r.ensureJobExists,
			r.ensureJobServiceExists,
			r.ensureWorkflowManagerHelperDataSecretExists,
			r.ensureStartTimeIsSet,
			r.ensurePhaseIsSetTo(enginev1.RunningTaskPhase),
		},
		enginev1.RunningTaskPhase: {
			r.ensureJobExists,
			r.ensureJobServiceExists,
			r.ensureWorkflowManagerHelperDataSecretExists,
			r.ensureTaskLifecycleIsSyncedToJob,
			// Task phase change is handled in ensureTaskLifecycleIsSyncedToJob
		},
		enginev1.SucceededTaskPhase: {
			r.ensureJobHasTerminated,
			r.ensureEndTimeIsSet,
		},
		enginev1.FailedTaskPhase: {
			r.ensureJobHasTerminated,
			r.ensureEndTimeIsSet,
		},
		internalDeleteTaskPhase: {
			r.ensureJobHasTerminated,
			r.ensureFinalizerIsUnset,
		},
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(TaskControllerName).
		For(&enginev1.Task{}).
		Watches(&enginev1.Workflow{}, handler.EnqueueRequestsFromMapFunc(r.mapWorkflowToTaskRequests)).
		WatchesRawSource(source.Channel(r.JobEventChannel, &handler.EnqueueRequestForObject{})).
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
