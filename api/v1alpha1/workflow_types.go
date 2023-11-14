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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	meta "github.com/nagare-media/engine/pkg/apis/meta"
)

const (
	// Label automatically assigned by nagare media engine indicating the workflow namespace. This allows quick filtering
	// for a specific workflow.
	WorkflowNamespaceLabel = "engine.nagare.media/workflow-namespace"

	// Label automatically assigned by nagare media engine indicating the workflow name. This allows quick filtering for a
	// specific workflow.
	WorkflowNameLabel = "engine.nagare.media/workflow-name"
)

const (
	WorkflowkProtectionFinalizer = "engine.nagare.media/workflow-protection"
)

// Specification of a Workflow.
type WorkflowSpec struct {
	// Human readable description of this Workflow.
	// +optional
	HumanReadable *HumanReadableWorkflowDescription `json:"humanReadable,omitempty"`

	// Named references to MediaLocations.
	// +listMapKey=name
	// +listType=map
	// +patchMergeKey=name
	// +patchStrategy=merge,retainKeys
	// +optional
	MediaLocations []NamedMediaLocationReference `json:"mediaLocations,omitempty" patchStrategy:"merge,retainKeys" patchMergeKey:"name"`

	// Workflow configuration values.
	// +optional
	Config map[string]string `json:"config,omitempty"`
}

type HumanReadableWorkflowDescription struct {
	// Human readable name of this Workflow.
	// +optional
	Name *string `json:"name,omitempty"`

	// Human readable description of this Workflow.
	// +optional
	Description *string `json:"description,omitempty"`
}

type NamedMediaLocationReference struct {
	// Name of the MediaLocation as used in the Workflow.
	// +kubebuilder:validation:Pattern="[a-zA-Z0-9-]+"
	Name string `json:"name"`

	// Reference to a MediaLocation of ClusterMediaLocation. Only references to these two kinds are allowed. A Workflow
	// can only reference MediaLocations from its own Namespace.
	Ref meta.LocalObjectReference `json:"ref"`
}

// Status of a Workflow.
type WorkflowStatus struct {
	// The status of this Workflow.
	// +optional
	Phase WorkflowPhase `json:"phase,omitempty"`

	// The latest available observations of an object's current state. When a Workflow fails, one of the conditions will
	// have type "Failed" and status true. When a Workflow is completed, one of the conditions will have type "Complete"
	// and status true.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=atomic
	Conditions []Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// A human readable message indicating why the Workflow is in this condition.
	Message string `json:"message,omitempty"`

	// Represents time when the Workflow controller first started processing a Workflow. It is represented in RFC3339 form
	// and is in UTC.
	// +optional
	QueuedTime *metav1.Time `json:"queuedTime,omitempty"`

	// Represents time when the Workflow controller transitioned to the "running" phase. It is represented in RFC3339 form
	// and is in UTC.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// Represents time when the Workflow has ended processing (either failed or completed). It is not guaranteed to be set
	// in happens-before order across separate operations. It is represented in RFC3339 form and is in UTC.
	// +optional
	EndTime *metav1.Time `json:"endTime,omitempty"`

	// The number of total Tasks.
	// +optional
	Total *int32 `json:"total,omitempty"`

	// The number of Tasks which reached phase "Initializing", "JobPending" or "Running".
	// +optional
	Active *int32 `json:"active,omitempty"`

	// The number of Tasks which reached phase "Succeeded".
	// +optional
	Succeeded *int32 `json:"succeeded,omitempty"`

	// The number of Tasks which reached phase "Failed".
	// +optional
	Failed *int32 `json:"failed,omitempty"`
}

// +kubebuilder:validation:Enum=Initializing;Running;AwaitingCompletion;Succeeded;Failed
type WorkflowPhase string

const (
	WorkflowPhaseInitializing       WorkflowPhase = "Initializing"
	WorkflowPhaseRunning            WorkflowPhase = "Running"
	WorkflowPhaseAwaitingCompletion WorkflowPhase = "AwaitingCompletion"
	WorkflowPhaseSucceeded          WorkflowPhase = "Succeeded"
	WorkflowPhaseFailed             WorkflowPhase = "Failed"
)

const (
	// WorkflowReadyConditionType means the Workflow has been processed by the Workflow controller.
	WorkflowReadyConditionType ConditionType = "Ready"

	// WorkflowCompleteConditionType means the Workflow has completed its execution.
	WorkflowCompleteConditionType ConditionType = "Complete"

	// WorkflowFailedConditionType means the Workflow has failed its execution.
	WorkflowFailedConditionType ConditionType = "Failed"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories={nagare-all,nme-all,nagare,nme}
// +kubebuilder:printcolumn:name="Human Name",type="string",JSONPath=`.spec.humanReadable.name`
// +kubebuilder:printcolumn:name="Status",type="string",JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=`.metadata.creationTimestamp`,description="CreationTimestamp is a timestamp representing the server time when this object was created. It is not guaranteed to be set in happens-before order across separate operations. Clients may not set this value. It is represented in RFC3339 form and is in UTC."
// +kubebuilder:printcolumn:name="Start",type="date",JSONPath=`.status.startTime`
// +kubebuilder:printcolumn:name="End",type="date",JSONPath=`.status.endTime`
// +kubebuilder:printcolumn:name="Total",type="integer",JSONPath=`.status.total`
// +kubebuilder:printcolumn:name="Active",type="integer",JSONPath=`.status.active`
// +kubebuilder:printcolumn:name="Succeeded",type="integer",JSONPath=`.status.succeeded`
// +kubebuilder:printcolumn:name="Failed",type="integer",JSONPath=`.status.failed`

// Workflow is the Schema for the workflows API
type Workflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkflowSpec   `json:"spec,omitempty"`
	Status WorkflowStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WorkflowList contains a list of Workflow
type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workflow `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Workflow{}, &WorkflowList{})
}
