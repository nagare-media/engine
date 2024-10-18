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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type WorkflowManagerHelperDataSpec struct {
	Workflow WorkflowManagerHelperDataWorkflow `json:"workflow"`
	Task     WorkflowManagerHelperDataTask     `json:"task"`
	System   WorkflowManagerHelperDataSystem   `json:"system"`
}

type WorkflowManagerHelperDataWorkflow struct {
	ID string `json:"id"`

	// +optional
	HumanReadable *HumanReadableWorkflowDescription `json:"humanReadable,omitempty"`
}

type WorkflowManagerHelperDataTask struct {
	ID string `json:"id"`

	// +optional
	HumanReadable *HumanReadableTaskDescription `json:"humanReadable,omitempty"`

	// +optional
	Inputs []Media `json:"inputs"`

	// +optional
	Outputs []Media `json:"outputs"`

	// +optional
	Config map[string]string `json:"config,omitempty"`
}

type WorkflowManagerHelperDataSystem struct {
	NATS NATSConfig `json:"nats"`
}

// +kubebuilder:object:root=true

// WorkflowManagerHelperData defines the data input for nagare media engine workflow-manager-helper.
type WorkflowManagerHelperData struct {
	metav1.TypeMeta `json:",inline"`

	WorkflowManagerHelperDataSpec `json:",inline"`
}

func init() {
	SchemeBuilder.Register(&WorkflowManagerHelperData{})
}
