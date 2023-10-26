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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	"github.com/nagare-media/engine/pkg/apis/resources"
)

type GatewayNBMPConfigurationSpec struct {
	Webserver       WebserverConfiguration       `json:"webserver"`
	WorkflowService WorkflowServiceConfiguration `json:"workflows"`
}

type WorkflowServiceConfiguration struct {
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

func (c *GatewayNBMPConfiguration) Default() {
	if c.Webserver.BindAddress == nil {
		c.Webserver.BindAddress = ptr.To[string](":8080")
	}

	if c.Webserver.ReadTimeout == nil {
		c.Webserver.ReadTimeout = &metav1.Duration{Duration: time.Minute}
	}

	if c.Webserver.WriteTimeout == nil {
		c.Webserver.WriteTimeout = &metav1.Duration{Duration: time.Minute}
	}

	if c.Webserver.IdleTimeout == nil {
		c.Webserver.IdleTimeout = &metav1.Duration{Duration: time.Minute}
	}

	if c.Webserver.Network == nil {
		c.Webserver.Network = ptr.To[string]("tcp")
	}

	if c.WorkflowService.DefaultKubernetesGPUResource == "" {
		// TODO: what should the default be?
		c.WorkflowService.DefaultKubernetesGPUResource = resources.NVIDIA_GPU
	}
}

func (c *GatewayNBMPConfiguration) Validate() error {
	if c.Webserver.BindAddress == nil {
		return errors.New("missing webserver.bindAddress")
	}

	if c.Webserver.ReadTimeout == nil {
		return errors.New("missing webserver.readTimeout")
	}

	if c.Webserver.WriteTimeout == nil {
		return errors.New("missing webserver.writeTimeout")
	}

	if c.Webserver.IdleTimeout == nil {
		return errors.New("missing webserver.idleTimeout")
	}

	if c.Webserver.Network == nil {
		return errors.New("missing webserver.network")
	}

	if c.Webserver.PublicBaseURL != nil && strings.HasSuffix(*c.Webserver.PublicBaseURL, "/") {
		return errors.New("trailing slash in webserver.publicBaseURL")
	}

	if c.WorkflowService.KubernetesNamespace == "" {
		return errors.New("missing services.kubernetesNamespace")
	}

	if c.WorkflowService.DefaultKubernetesGPUResource == "" {
		return errors.New("missing services.defaultKubernetesGPUResource")
	}
	return nil
}
