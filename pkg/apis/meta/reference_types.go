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

import (
	"encoding/json"

	"k8s.io/apimachinery/pkg/types"
)

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
	// +optional
	Name string `json:"name,omitempty"`

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

	// Resolved data from ConfigMap or Secret.
	Data map[string][]byte `json:"-"`

	// Flag indicating whether to only include resolved ConfigMap or Secret data in JSON output.
	dataOnly bool `json:"-"`
}

type resolvedConfigMapOrSecretReference struct {
	ObjectReference `json:",inline"`
	Key             string            `json:"key,omitempty"`
	Data            map[string][]byte `json:"data,omitempty"`
}

type unresolvedConfigMapOrSecretReference struct {
	ObjectReference `json:",inline"`
	Key             string `json:"key,omitempty"`
}

var (
	_ json.Marshaler   = &ConfigMapOrSecretReference{}
	_ json.Unmarshaler = &ConfigMapOrSecretReference{}
)

func (r *ConfigMapOrSecretReference) SetMarshalOnlyData(dataOnly bool) {
	r.dataOnly = dataOnly
}

func (r *ConfigMapOrSecretReference) MarshalJSON() ([]byte, error) {
	if r == nil {
		return json.Marshal(nil)
	}

	var data any
	if r.dataOnly {
		data = resolvedConfigMapOrSecretReference{
			Data: r.Data,
		}
	} else {
		data = unresolvedConfigMapOrSecretReference{
			ObjectReference: r.ObjectReference,
			Key:             r.Key,
		}
	}

	return json.Marshal(data)
}

func (r *ConfigMapOrSecretReference) UnmarshalJSON(b []byte) error {
	tmp := resolvedConfigMapOrSecretReference{}
	err := json.Unmarshal(b, &tmp)
	if err != nil {
		return err
	}
	r.ObjectReference = tmp.ObjectReference
	r.Key = tmp.Key
	r.Data = tmp.Data
	return nil
}
