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

	meta "github.com/nagare-media/engine/pkg/apis/meta"
)

const (
	// MediaProcessingEntities with this annotation set to "true" will be used as default MPE for running tasks. This
	// annotation can only be used once per MediaProcessingEntity in a single namespace or once per
	// ClusterMediaProcessingEntity in the whole Kubernetes cluster. A MediaProcessingEntity with this annotation has
	// precedence over a ClusterMediaProcessingEntity.
	IsDefaultMediaProcessingEntityAnnotation = "engine.nagare.media/is-default-media-processing-entity"

	// Description of the location the MediaProcessingEntity is in (e.g. "DE-muenster").
	MediaProcessingEntityLocationLabel = "engine.nagare.media/media-processing-entity-location"
)

const (
	// Protects MediaProcessingEntities from deletion before cleanup.
	MediaProcessingEntityProtectionFinalizer = "engine.nagare.media/mpe-protection"
)

const (
	// Default key used in Secrets for the Kubeconfig.
	DefaultSecretKeyKubeconfig = "kubeconfig"
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
	// Configures the namespace Jobs should run in. For MediaProcessingEntities this field is optional in which case the
	// Jobs are created in the namespace the MediaProcessingEntity is in. This field is required for
	// ClusterMediaProcessingEntities.
	// +optional
	Namespace string `json:"namespace,omitempty"`
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

// Status of a MediaProcessingEntity
type MediaProcessingEntityStatus struct {
	// The latest available observations of an object's current state. When a connection to a MediaProcessingEntity is
	// established, one of the conditions will have type "Ready" and status true.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=atomic
	Conditions []Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type"`

	// A human readable message indicating why the MediaProcessingEntity is in this condition.
	Message string `json:"message,omitempty"`
}

const (
	// MediaProcessingEntityReadyConditionType means the a connection to the MediaProcessingEntity could be established.
	MediaProcessingEntityReadyConditionType = ConditionType("Ready")

	// MediaProcessingEntityFailedConditionType means the a connection to the MediaProcessingEntity could not be
	// established.
	MediaProcessingEntityFailedConditionType = ConditionType("Failed")
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories={nagare-all,nme-all}
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`

// MediaProcessingEntity is the Schema for the mediaprocessingentities API
type MediaProcessingEntity struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MediaProcessingEntitySpec   `json:"spec,omitempty"`
	Status MediaProcessingEntityStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MediaProcessingEntityList contains a list of MediaProcessingEntity
type MediaProcessingEntityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MediaProcessingEntity `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MediaProcessingEntity{}, &MediaProcessingEntityList{})
}
