/*
Copyright 2022-2025 The nagare media authors

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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// AMD GPU resource name.
	AMD_GPU = corev1.ResourceName("amd.com/gpu")

	// Intel GPU prefix resource name.
	IntelGPUPrefix = "gpu.intel.com/"

	// NVIDIA GPU resource name.
	NVIDIA_GPU = corev1.ResourceName("nvidia.com/gpu")

	// Shared NVIDIA GPU resource name.
	NVIDIA_GPUShared = corev1.ResourceName("nvidia.com/gpu.shared")
)

var (
	DefaultGatewayNBMPConfig = GatewayNBMPConfig{
		GatewayNBMPConfigSpec: GatewayNBMPConfigSpec{
			WorkflowService: GatewayNBMPWorkflowServiceConfig{
				Kubernetes: GatewayNBMPWorkflowServiceKubernetesConfig{
					GPUResourceName: NVIDIA_GPU,
				},
			},
			Webserver: DefaultWebserverConfig,
		},
	}
)

type GatewayNBMPConfigSpec struct {
	WorkflowService GatewayNBMPWorkflowServiceConfig `json:"workflow"`
	Webserver       WebserverConfig                  `json:"webserver"`
}

type GatewayNBMPWorkflowServiceConfig struct {
	// Kubernetes related configuration.
	Kubernetes GatewayNBMPWorkflowServiceKubernetesConfig `json:"kubernetes"`
}

type GatewayNBMPWorkflowServiceKubernetesConfig struct {
	// Namespace to use for nagare media engine Kubernetes resources.
	Namespace string `json:"namespace"`

	// Name of the GPU resource used in the Kubernetes cluster. Defaults to "nvidia.com/gpu".
	// +kubebuilder:default="nvidia.com/gpu"
	// +optional
	GPUResourceName corev1.ResourceName `json:"gpuResourceName,omitempty"`
}

// +kubebuilder:object:root=true

// GatewayNBMPConfig defines the configuration for nagare media engine gateway-nbmp.
type GatewayNBMPConfig struct {
	metav1.TypeMeta `json:",inline"`

	GatewayNBMPConfigSpec `json:",inline"`
}

func (c *GatewayNBMPConfig) Default() {
	c.doDefaultWithValuesFrom(DefaultGatewayNBMPConfig)
}

func (c *GatewayNBMPConfig) DefaultWithValuesFrom(d GatewayNBMPConfig) {
	c.doDefaultWithValuesFrom(d)
	c.doDefaultWithValuesFrom(DefaultGatewayNBMPConfig)
}

func (c *GatewayNBMPConfig) doDefaultWithValuesFrom(d GatewayNBMPConfig) {
	if c.WorkflowService.Kubernetes.GPUResourceName == "" {
		c.WorkflowService.Kubernetes.GPUResourceName = d.WorkflowService.Kubernetes.GPUResourceName
	}
	c.Webserver.DefaultWithValuesFrom(d.Webserver)
}

func (c *GatewayNBMPConfig) Validate() error {
	if c.WorkflowService.Kubernetes.Namespace == "" {
		return errors.New("missing workflow.kubernetes.namespace")
	}
	if c.WorkflowService.Kubernetes.GPUResourceName == "" {
		return errors.New("missing workflow.kubernetes.gpuResourceName")
	}
	if err := c.Webserver.Validate("webserver."); err != nil {
		return err
	}
	return nil
}

func init() {
	SchemeBuilder.Register(&GatewayNBMPConfig{})
}
