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
	// +kubebuilder:validation:Pattern="[a-Z0-9-]+"
	Name string `json:"name"`

	// Reference to a MediaLocation of ClusterMediaLocation. Only references to these two kinds are allowed. A Workflow
	// can only reference MediaLocations from its own Namespace.
	Ref meta.LocalObjectReference `json:"ref"`
}

// Status of a Workflow.
type WorkflowStatus struct {
}

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
