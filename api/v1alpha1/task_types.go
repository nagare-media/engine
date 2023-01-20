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

package v1alpha1

import (
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	meta "github.com/nagare-media/engine/pkg/apis/meta"
)

const (
	TaskLabel = "engine.nagare.media/task"
)

const (
	JobProtectionFinalizer = "engine.nagare.media/job-protection"
)

// Specification of a Task.
type TaskSpec struct {
	// Reference to a TaskTemplate or ClusterTaskTemplate. Only references to these two kinds are allowed. A Task can only
	// reference TaskTemplates from its own Namespace.
	// +optional
	TaskTemplateRef *meta.LocalObjectReference `json:"taskTemplateRef,omitempty"`

	// Human readable description of this Task.
	// +optional
	HumanReadable *HumanReadableTaskDescription `json:"humanReadable,omitempty"`

	// Reference to a MediaProcessingEntity or ClusterMediaProcessingEntity. Only references to these two kinds are
	// allowed. A Task can only reference MediaProcessingEntities from its own Namespace. This field is required if no
	// mediaProcessingEntitySelector is specified. If both are specified, mediaProcessingEntityRef has precedence. Both
	// fields may be omitted if a TaskTemplate is used that specifies a MediaProcessingEntity.
	// +optional
	MediaProcessingEntityRef *meta.LocalObjectReference `json:"mediaProcessingEntityRef,omitempty"`

	// Label selector for a MediaProcessingEntity or ClusterMediaProcessingEntity. MediaProcessingEntity has precedence
	// over ClusterMediaProcessingEntity. If multiple Media Processing Entities are selected, the newest one is chosen.
	// This field is required if no mediaProcessingEntityRef is specified. If both are specified, mediaProcessingEntityRef
	// has precedence. Both fields may be omitted if a TaskTemplate is used that specifies a MediaProcessingEntity.
	// +optional
	MediaProcessingEntitySelector *metav1.LabelSelector `json:"mediaProcessingEntitySelector,omitempty"`

	// Reference to a Workflow. Only references to Workflow are allowed. A Task can only reference Workflow from its own
	// Namespace.
	WorkflowRef meta.LocalObjectReference `json:"workflowRef"`

	// Reference to a Function or ClusterFunction. Only references to these two kinds are allowed. A Task can only
	// reference Functions from its own Namespace. This field is required if no FunctionSelector is specified. If both are
	// specified, FunctionRef has precedence. Both fields may be omitted if a TaskTemplate is used that specifies a
	// Function.
	// +optional
	FunctionRef *meta.LocalObjectReference `json:"functionRef,omitempty"`

	// Label selector for a Function or ClusterFunction. Function has precedence over ClusterFunction. If multiple
	// Functions are selected, the Function with the newest version is chosen. This field is required if no FunctionRef is
	// specified. If both are specified, FunctionRef has precedence. Both fields may be omitted if a TaskTemplate is used
	// that specifies a Function.
	// +optional
	FunctionSelector *metav1.LabelSelector `json:"functionSelector,omitempty"`

	// Patches applied to the Job template description of the Function.
	//
	// Only these fields may be patched:
	// TODO: update white list
	// TODO: enforce limits
	// TODO: which fields should be patchable?
	// +optional
	TemplatePatches *batchv1.JobTemplateSpec `json:"templatePatches,omitempty"`

	// Policy for dealing with a failed Job resulting from this Task. Conditions for when a Job is considered as failure
	// are defined in the templates `jobFailurePolicy` field.
	// +optional
	JobFailurePolicy *JobFailurePolicy `json:"jobFailurePolicy,omitempty"`

	// Configuration values.
	// +optional
	Config *apiextensionsv1.JSON `json:"config,omitempty"`
}

type HumanReadableTaskDescription struct {
	// Human readable name of this Task.
	// +optional
	Name *string `json:"name,omitempty"`

	// Human readable description of this Task.
	// +optional
	Description *string `json:"description,omitempty"`
}

type JobFailurePolicy struct {
	// The default action taken when a job fails and no rule applies.
	// +kubebuilder:default=FailWorkflow
	// +optional
	DefaultAction *JobFailurePolicyAction `json:"defaultAction,omitempty"`

	// TODO: allow for more detailed policy rules
	// At most 20 elements are allowed.
	// +kubebuilder:validation:MaxItems=20
	// +listType=atomic
	// Rules []JobFailurePolicyRule `json:"rules"`
}

// +kubebuilder:validation:Enum=FailWorkflow;Ignore
type JobFailurePolicyAction string

var (
	// This is an action which might be taken on a Job failure - mark the Task and Workflow as Failed and terminate all
	// running Tasks.
	JobFailurePolicyActionFailWorkflow JobFailurePolicyAction = "FailWorkflow"

	// This is an action which might be taken on a Job failure - mark the Task as Failed but do not terminate other Tasks.
	// Dependent Tasks have to deal with this fail-state.
	JobFailurePolicyActionIgnore JobFailurePolicyAction = "Ignore"
)

// Status of a Task.
type TaskStatus struct {
	// The status of this Task.
	Phase TaskPhase `json:"phase"`

	// The latest available observations of an object's current state. When a Task fails, one of the conditions will have
	// type "Failed" and status true. When a Task is completed, one of the conditions will have type "Complete" and status
	// true.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=atomic
	Conditions []TaskCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// A human readable message indicating why the Task is in this condition.
	Message string `json:"message,omitempty"`

	// Represents time when the Task controller started processing a Task. It is represented in RFC3339 form and is in
	// UTC.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// Represents time when the Task has ended processing (either failed or completed). It is not guaranteed to be set in
	// happens-before order across separate operations. It is represented in RFC3339 form and is in UTC.
	// +optional
	EndTime *metav1.Time `json:"endTime,omitempty"`

	// Reference to the selected MediaProcessingEntity.
	// +optional
	MediaProcessingEntityRef *meta.LocalObjectReference `json:"mediaProcessingEntityRef,omitempty"`

	// Reference to the selected Function.
	// +optional
	FunctionRef *meta.LocalObjectReference `json:"functionRef,omitempty"`

	// Reference to the Job.
	// +optional
	JobRef *meta.ExactObjectReference `json:"jobRef,omitempty"`
}

// +kubebuilder:validation:Enum=Initializing;JobPending;Running;Succeeded;Failed
type TaskPhase string

const (
	TaskPhaseInitializing TaskPhase = "Initializing"
	TaskPhaseJobPending   TaskPhase = "JobPending"
	TaskPhaseRunning      TaskPhase = "Running"
	TaskPhaseSucceeded    TaskPhase = "Succeeded"
	TaskPhaseFailed       TaskPhase = "Failed"
)

// Description of the current Task status.
type TaskCondition struct {
	// Type of Task condition.
	Type TaskConditionType `json:"type"`

	// Status of the condition, one of True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`

	// Last time the condition was checked.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`

	// Last time the condition transit from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`

	// (brief) reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`

	// Human readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

// +kubebuilder:validation:Enum=Initialized;Ready;Complete;Failed
type TaskConditionType string

var (
	// TaskConditionTypeInitialized means the Task has been processed by the Task controller and a Job was created.
	TaskConditionTypeInitialized TaskConditionType = "Initialized"

	// TaskConditionTypeReady means the Task has been processed by the Task controller and a Job was created.
	TaskConditionTypeReady TaskConditionType = "Ready"

	// TaskConditionTypeComplete means the Task has completed its execution.
	TaskConditionTypeComplete TaskConditionType = "Complete"

	// TaskConditionTypeFailed means the Task has failed its execution.
	TaskConditionTypeFailed TaskConditionType = "Failed"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories={nagare-all,nme-all,nagare,nme}
// +kubebuilder:printcolumn:name="MPE",type="string",JSONPath=`.status.mediaProcessingEntityRef.name`
// +kubebuilder:printcolumn:name="Workflow",type="string",JSONPath=`.spec.workflowRef.name`
// +kubebuilder:printcolumn:name="Function",type="string",JSONPath=`.status.functionRef.name`
// +kubebuilder:printcolumn:name="Human Name",type="string",JSONPath=`.spec.humanReadable.name`
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`,description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."
// +kubebuilder:printcolumn:name="Start",type="date",JSONPath=`.status.startTime`
// +kubebuilder:printcolumn:name="End",type="date",JSONPath=`.status.endTime`

// Task is the Schema for the tasks API
type Task struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TaskSpec   `json:"spec,omitempty"`
	Status TaskStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// TaskList contains a list of Task
type TaskList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Task `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Task{}, &TaskList{})
}
