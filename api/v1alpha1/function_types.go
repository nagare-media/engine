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
)

const (
	// Selection label to indicate the type of function.
	BetaFunctionTypeLabel = "beta.engine.nagare.media/function-type"

	// Selection label to indicate the language used in script functions.
	BetaFunctionScriptLanguageTypeLabel = "beta.engine.nagare.media/function-script-language"
)

// Specification of a function.
type FunctionSpec struct {
	// Version number of this function following the SemVer 2.0.0 specification (see https://semver.org/spec/v2.0.0.html).
	// When label selections are used to specify functions, the version number determines the final function selection if
	// multiple functions have the same labels.
	// Cannot be updated.
	// +kubebuilder:validation:Pattern="^(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(-(0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(\\.(0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*)?(\\+[0-9a-zA-Z-]+(\\.[0-9a-zA-Z-]+)*)?$"
	Version string `json:"version"`

	// Describes the Job that will be created when executing this function.
	//
	// The following limitations are imposed:
	// - spec.template.spec.restartPolicy: must be "OnFailure" or "Never". The default is "Never"
	//
	// Cannot be updated.
	// TODO: make some exceptions
	// TODO: hinder changes to this field
	Template batchv1.JobTemplateSpec `json:"template"`

	// Indicates that this function can only run on a local Media Processing Entities (MPE), e.g. if this function changes
	// the Workflow on the management cluster. The default is "false"
	// +kubebuilder:default=false
	// +optional
	LocalMediaProcessingEntitiesOnly bool `json:"localMediaProcessingEntitiesOnly"`

	// Default configuration values.
	// +optional
	DefaultConfig map[string]string `json:"defaultConfig,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:categories={nagare-all,nme-all}

// Function is the Schema for the functions API
type Function struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec FunctionSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// FunctionList contains a list of Function
type FunctionList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Function `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Function{}, &FunctionList{})
}
