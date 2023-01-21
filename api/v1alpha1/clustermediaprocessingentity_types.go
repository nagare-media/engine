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

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={nagare-all,nme-all}
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`

// ClusterMediaProcessingEntity is the Schema for the clustermediaprocessingentities API
type ClusterMediaProcessingEntity struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MediaProcessingEntitySpec   `json:"spec,omitempty"`
	Status MediaProcessingEntityStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterMediaProcessingEntityList contains a list of ClusterMediaProcessingEntity
type ClusterMediaProcessingEntityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterMediaProcessingEntity `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterMediaProcessingEntity{}, &ClusterMediaProcessingEntityList{})
}
