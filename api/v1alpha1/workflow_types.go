/*
Copyright 2022 The nagare media authors

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
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	meta "github.com/nagare-media/engine/pkg/apis/meta"
)

const (
	WorkflowLabel = "engine.nagare.media/workflow"
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
	Config *apiextensionsv1.JSON `json:"config,omitempty"`
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
	// The latest available observations of an object's current state. When a Workflow fails, one of the conditions will
	// have type "Failed" and status true. When a Workflow is completed, one of the conditions will have type "Complete"
	// and status true.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=atomic
	Conditions []WorkflowCondition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// Represents time when the Workflow controller started processing a Workflow. It is represented in RFC3339 form and
	// is in UTC.
	// +optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// Represents time when the Workflow has ended processing (either failed or completed). It is not guaranteed to be set
	// in happens-before order across separate operations. It is represented in RFC3339 form and is in UTC.
	// +optional
	EndTime *metav1.Time `json:"endTime,omitempty"`
}

type WorkflowCondition struct {
	// Type of Workflow condition.
	Type WorkflowConditionType `json:"type"`

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

// +kubebuilder:validation:Enum=Ready;Complete;Failed
type WorkflowConditionType string

const (
	// WorkflowReady means the Workflow has been processed by the Workflow controller.
	WorkflowReady = "Ready"

	// WorkflowComplete means the Workflow has completed its execution.
	WorkflowComplete = "Complete"

	// WorkflowFailed means the Workflow has failed its execution.
	WorkflowFailed = "Failed"
)

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:categories={nagare-all,nme-all,nagare,nme}

// Workflow is the Schema for the workflows API
type Workflow struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkflowSpec   `json:"spec,omitempty"`
	Status WorkflowStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WorkflowList contains a list of Workflow
type WorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workflow `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Workflow{}, &WorkflowList{})
}
