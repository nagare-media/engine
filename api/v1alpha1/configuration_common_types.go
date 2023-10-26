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

type WebserverConfiguration struct {
	// +optional
	BindAddress *string `json:"bindAddress,omitempty"`

	// +optional
	ReadTimeout *metav1.Duration `json:"readTimeout,omitempty"`

	// +optional
	WriteTimeout *metav1.Duration `json:"writeTimeout,omitempty"`

	// +optional
	IdleTimeout *metav1.Duration `json:"idleTimeout,omitempty"`

	// +kubebuilder:validation:Enum=tcp;tcp4;tcp6
	// +optional
	Network *string `json:"network,omitempty"`

	// +optional
	PublicBaseURL *string `json:"publicBaseURL,omitempty"`
}
