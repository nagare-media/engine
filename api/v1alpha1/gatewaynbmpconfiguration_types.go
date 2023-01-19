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
	"errors"
	"strings"
	"time"

	"github.com/nagare-media/engine/pkg/apis/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
)

type GatewayNBMPConfigurationSpec struct {
	Webserver WebserverConfiguration `json:"webserver"`
	Services  ServicesConfiguration  `json:"services"`
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

	// +optional
	PublicBaseURL *string `json:"publicBaseURL"`
}

type ServicesConfiguration struct {
	// Limit gateway-nbmp to a specific Kubernetes namespace.
	// +optional
	KubernetesNamespace string `json:"kubernetesNamespace,omitempty"`

	// Name of the GPU resource used in the Kubernetes cluster (e.g. "nvidia.com/gpu").
	// +optional
	DefaultKubernetesGPUResource corev1.ResourceName `json:"defaultKubernetesGPUResource,omitempty"`
}

// +kubebuilder:object:root=true

// GatewayNBMPConfiguration defines the configuration for nagare media engine gateway-nbmp.
type GatewayNBMPConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	GatewayNBMPConfigurationSpec `json:",inline"`
}

func init() {
	SchemeBuilder.Register(&GatewayNBMPConfiguration{})
}

func (wc *GatewayNBMPConfiguration) Default() {
	if wc.Webserver.BindAddress == nil {
		wc.Webserver.BindAddress = pointer.String(":8080")
	}

	if wc.Webserver.ReadTimeout == nil {
		wc.Webserver.ReadTimeout = pointer.Duration(time.Minute)
	}

	if wc.Webserver.WriteTimeout == nil {
		wc.Webserver.WriteTimeout = pointer.Duration(time.Minute)
	}

	if wc.Webserver.IdleTimeout == nil {
		wc.Webserver.IdleTimeout = pointer.Duration(time.Minute)
	}

	if wc.Webserver.Network == nil {
		wc.Webserver.Network = pointer.String("tcp")
	}

	if wc.Services.DefaultKubernetesGPUResource == "" {
		// TODO: what should the default be?
		wc.Services.DefaultKubernetesGPUResource = resources.NVIDIA_GPU
	}
}

func (wc *GatewayNBMPConfiguration) Validate() error {
	if wc.Webserver.BindAddress == nil {
		return errors.New("missing webserver.bindAddress")
	}
	if wc.Webserver.ReadTimeout == nil {
		return errors.New("missing webserver.readTimeout")
	}
	if wc.Webserver.WriteTimeout == nil {
		return errors.New("missing webserver.writeTimeout")
	}
	if wc.Webserver.IdleTimeout == nil {
		return errors.New("missing webserver.idleTimeout")
	}
	if wc.Webserver.Network == nil {
		return errors.New("missing webserver.network")
	}
	if wc.Webserver.PublicBaseURL != nil && strings.HasSuffix(*wc.Webserver.PublicBaseURL, "/") {
		return errors.New("trailing slash in webserver.publicBaseURL")
	}
	if wc.Services.KubernetesNamespace == "" {
		return errors.New("missing services.kubernetesNamespace")
	}
	if wc.Services.DefaultKubernetesGPUResource == "" {
		return errors.New("missing services.defaultKubernetesGPUResource")
	}
	return nil
}
