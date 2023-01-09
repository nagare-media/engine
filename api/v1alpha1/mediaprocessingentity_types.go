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

	meta "github.com/nagare-media/engine/pkg/apis/meta"
)

const (
	// MediaProcessingEntities with this annotation set to "true" will be used as default MPE for running tasks. This
	// annotation can only be used once per MediaProcessingEntity in a singe namespace or once per
	// ClusterMediaProcessingEntity in the whole Kubernetes cluster. A MediaProcessingEntity with this annotation has
	// precedence over a ClusterMediaProcessingEntity.
	BetaIsDefaultMediaProcessingEntityAnnotation = "beta.engine.nagare.media/is-default-media-processing-entity"
)

// Specification of a Media Processing Entity (MPE).
type MediaProcessingEntitySpec struct {
	MediaProcessingEntityConfig `json:",inline"`
}

// Configuration of the Media Processing Entity (MPE).
// Exactly one of these must be set.
// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type MediaProcessingEntityConfig struct {
	// Configures the Media Processing Entity (MPE) to talk to the local Kubernetes cluster.
	// +optional
	Local *LocalMediaProcessingEntity `json:"local,omitempty"`

	// Configures the Media Processing Entity (MPE) to talk to a remote Kubernetes cluster.
	// +optional
	Remote *RemoteMediaProcessingEntity `json:"remote,omitempty"`
}

// Configuration of a local Media Processing Entity (MPE).
type LocalMediaProcessingEntity struct {
}

// Configuration of a remote Media Processing Entity (MPE).
type RemoteMediaProcessingEntity struct {
	// Kubeconfig that defines connection configuration.
	Kubeconfig Kubeconfig `json:"kubeconfig"`
}

// Configuration for connecting to a Kubernetes cluster.
type Kubeconfig struct {
	// Reference to a Secret that contains the kubeconfig in specified key. If no key is specified, "kubeconfig" is used
	// by default. Only references to Secrets are allowed. A MediaProcessingEntity can only reference Secrets from its
	// own Namespace.
	SecretRef meta.ConfigMapOrSecretReference `json:"secretRef"`
}

//+kubebuilder:object:root=true
//+kubebuilder:resource:categories={nagare-all,nme-all}

// MediaProcessingEntity is the Schema for the mediaprocessingentities API
type MediaProcessingEntity struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec MediaProcessingEntitySpec `json:"spec,omitempty"`
}

//+kubebuilder:object:root=true

// MediaProcessingEntityList contains a list of MediaProcessingEntity
type MediaProcessingEntityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MediaProcessingEntity `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MediaProcessingEntity{}, &MediaProcessingEntityList{})
}
