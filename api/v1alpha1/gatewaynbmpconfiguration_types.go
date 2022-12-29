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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilpointer "k8s.io/utils/pointer"
)

type GatewayNBMPConfigurationSpec struct {
	// +optional
	Webserver WebserverConfiguration `json:"webserver,omitempty"`
}

type WebserverConfiguration struct {
	// +optional
	BindAddress *string `json:"bindAddress,omitempty"`

	// +optional
	ReadTimeout *time.Duration `json:"readTimeout"`

	// +optional
	WriteTimeout *time.Duration `json:"writeTimeout"`

	// +optional
	IdleTimeout *time.Duration `json:"idleTimeout"`

	// +kubebuilder:validation:Enum=tcp;tcp4;tcp6
	// +optional
	Network *string `json:"network"`
}

func (wc *WebserverConfiguration) Default() {
	if wc.BindAddress == nil {
		wc.BindAddress = utilpointer.String(":8080")
	}

	if wc.ReadTimeout == nil {
		wc.ReadTimeout = utilpointer.Duration(time.Minute)
	}

	if wc.WriteTimeout == nil {
		wc.WriteTimeout = utilpointer.Duration(time.Minute)
	}

	if wc.IdleTimeout == nil {
		wc.IdleTimeout = utilpointer.Duration(time.Minute)
	}

	if wc.Network == nil {
		wc.Network = utilpointer.String("tcp")
	}
}

//+kubebuilder:object:root=true

// GatewayNBMPConfiguration defines the configuration for nagare media engine gateway-nbmp.
type GatewayNBMPConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	GatewayNBMPConfigurationSpec `json:",inline"`
}

func init() {
	SchemeBuilder.Register(&GatewayNBMPConfiguration{})
}
