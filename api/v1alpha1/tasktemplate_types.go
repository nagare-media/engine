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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	meta "github.com/nagare-media/engine/pkg/apis/meta"
)

// TaskTemplateSpec defines the desired state of TaskTemplate
type TaskTemplateSpec struct {
	// Human readable description of this task.
	// +optional
	HumanReadable *HumanReadableTaskDescription `json:"humanReadable,omitempty"`

	// Reference to a MediaProcessingEntity or ClusterMediaProcessingEntity. Only references to these two kinds are
	// allowed. A TaskTemplate can only reference MediaProcessingEntities from its own Namespace.
	// +optional
	MediaProcessingEntityRef *meta.LocalObjectReference `json:"mediaProcessingEntityRef,omitempty"`

	// Label selector for a MediaProcessingEntity or ClusterMediaProcessingEntity. MediaProcessingEntity has precedence
	// over ClusterMediaProcessingEntity. If multiple Media Processing Entities are selected, the newest one is chosen.
	// +optional
	MediaProcessingEntitySelector *metav1.LabelSelector `json:"mediaProcessingEntitySelector,omitempty"`

	// Reference to a Function or ClusterFunction. Only references to these two kinds are allowed. A TaskTemplate can only
	// reference Functions from its own Namespace.
	// +optional
	FunctionRef *meta.LocalObjectReference `json:"functionRef,omitempty"`

	// Label selector for a Function or ClusterFunction. Function has precedence over ClusterFunction. If multiple
	// Functions are selected, the Function with the newest version is chosen.
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
	Config map[string]string `json:"config,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories={nagare-all,nme-all}

// TaskTemplate is the Schema for the tasktemplates API
type TaskTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec TaskTemplateSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// TaskTemplateList contains a list of TaskTemplate
type TaskTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TaskTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TaskTemplate{}, &TaskTemplateList{})
}
