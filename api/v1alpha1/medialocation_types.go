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

// MediaLocationSpec defines the desired state of MediaLocation
type MediaLocationSpec struct {
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:categories={nagare-all,nme-all}

// MediaLocation is the Schema for the medialocations API
type MediaLocation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec MediaLocationSpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// MediaLocationList contains a list of MediaLocation
type MediaLocationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MediaLocation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MediaLocation{}, &MediaLocationList{})
}
