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

package meta

import "k8s.io/apimachinery/pkg/types"

// Reference to a local object.
type LocalObjectReference struct {
	// API version of the referred object.
	// +optional
	APIVersion string `json:"apiVersion,omitempty"`

	// Kind of the referred object.
	// +optional
	Kind string `json:"kind,omitempty"`

	// Name of the referred object.
	Name string `json:"name"`
}

func (ref LocalObjectReference) ObjectReference(namespace string) ObjectReference {
	return ObjectReference{
		APIVersion: ref.APIVersion,
		Kind:       ref.Kind,
		Name:       ref.Name,
		Namespace:  namespace,
	}
}

// Reference to an object.
type ObjectReference struct {
	// API version of the referred object.
	// +optional
	APIVersion string `json:"apiVersion,omitempty"`

	// Kind of the referred object.
	// +optional
	Kind string `json:"kind,omitempty"`

	// Name of the referred object.
	Name string `json:"name"`

	// Namespace of the referred object.
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

func (ref ObjectReference) LocalObjectReference() LocalObjectReference {
	return LocalObjectReference{
		APIVersion: ref.APIVersion,
		Kind:       ref.Kind,
		Name:       ref.Name,
	}
}

// Reference to an exact object including the UID.
type ExactObjectReference struct {
	ObjectReference `json:",inline"`

	// UID of the object.
	UID types.UID `json:"uid"`
}

// Specifies a reference to a ConfigMap or Secret.
type ConfigMapOrSecretReference struct {
	ObjectReference `json:",inline"`

	// Key within the ConfigMap or Secret.
	// +optional
	Key string `json:"key,omitempty"`
}
