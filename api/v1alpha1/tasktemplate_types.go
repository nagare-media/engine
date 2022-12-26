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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TaskTemplateSpec defines the desired state of TaskTemplate
type TaskTemplateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of TaskTemplate. Edit tasktemplate_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// TaskTemplateStatus defines the observed state of TaskTemplate
type TaskTemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TaskTemplate is the Schema for the tasktemplates API
type TaskTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TaskTemplateSpec   `json:"spec,omitempty"`
	Status TaskTemplateStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TaskTemplateList contains a list of TaskTemplate
type TaskTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TaskTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TaskTemplate{}, &TaskTemplateList{})
}
