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

const (
	WorkflowManagerHelperDataSecretType = "engine.nagare.media/workflow-manager-helper-data"
)

type WorkflowManagerHelperDataSpec struct {
	Function WorkflowManagerHelperDataFunction `json:"function"`
	Workflow WorkflowManagerHelperDataWorkflow `json:"workflow"`
	Task     WorkflowManagerHelperDataTask     `json:"task"`
	System   WorkflowManagerHelperDataSystem   `json:"system"`
}

type WorkflowManagerHelperDataFunction struct {
	Name string `json:"name"`

	Version string `json:"version"`
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
	InputPorts []InputPortBinding `json:"inputPorts,omitempty"`

	// +optional
	OutputPorts []OutputPortBinding `json:"outputPorts,omitempty"`

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
