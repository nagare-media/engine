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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlCfg "sigs.k8s.io/controller-runtime/pkg/config/v1alpha1"

	"github.com/nagare-media/models.go/base"
)

// +kubebuilder:object:root=true

// WorkflowManagerConfiguration defines the configuration for nagare media engine controller manager.
type WorkflowManagerConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	// WorkflowManagerConfigurationSpec returns the configurations for controllers
	// TODO: migrate to custom struct
	ctrlCfg.ControllerManagerConfigurationSpec `json:",inline"`

	// Duration to wait after all Tasks of a Workflow terminated to mark the Workflow as successful. This helps mitigate
	// race conditions and should not be too low. Defaults to "20s".
	// +optional
	WorkflowTerminationWaitingDuration *metav1.Duration `json:"workflowTerminationWaitingDuration,omitempty"`

	// Duration before marking a remove MediaProcessingEntity as ready. Local MediaProcessingEntities are marked as ready
	// immediately. Defaults to "5s".
	// +optional
	RemoteMediaProcessingEntityStabilizingDuration *metav1.Duration `json:"remoteMediaProcessingEntityStabilizingDuration,omitempty"`

	// NATS connection configuration.
	NATS NATSConfig `json:"nats"`
}

type NATSConfig struct {
	// TODO: add auth methods
	URL base.URI `json:"url"`
}

func (c *WorkflowManagerConfiguration) Default() {
	if c.WorkflowTerminationWaitingDuration == nil {
		c.WorkflowTerminationWaitingDuration = &metav1.Duration{Duration: 20 * time.Second}
	}
	if c.RemoteMediaProcessingEntityStabilizingDuration == nil {
		c.RemoteMediaProcessingEntityStabilizingDuration = &metav1.Duration{Duration: 5 * time.Second}
	}
}

func (c *WorkflowManagerConfiguration) Validate() error {
	return nil
}

func init() {
	SchemeBuilder.Register(&WorkflowManagerConfiguration{})
}
