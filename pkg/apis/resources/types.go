/*
Copyright 2023 The nagare media authors

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

package resources

import corev1 "k8s.io/api/core/v1"

const (
	// AMD GPU resource name.
	AMD_GPU corev1.ResourceName = "amd.com/gpu"

	// Intel GPU prefix resource name.
	IntelGPUPrefix string = "gpu.intel.com/"

	// NVIDIA GPU resource name.
	NVIDIA_GPU corev1.ResourceName = "nvidia.com/gpu"

	// Shared NVIDIA GPU resource name.
	NVIDIA_GPUShared corev1.ResourceName = "nvidia.com/gpu.shared"
)
