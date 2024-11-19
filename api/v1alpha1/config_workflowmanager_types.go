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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

var (
	DefaultWorkflowManagerConfig = WorkflowManagerConfig{
		WorkflowManagerConfigSpec: WorkflowManagerConfigSpec{
			CacheConfig: DefaultKubernetesCacheConfig,
			LeaderElection: LeaderElectionConfig{
				LeaderElect:  ptr.To(false),
				ResourceName: "a3df9e9e.engine.nagare.media",
			},
			WorkflowTerminationWaitingDuration:             &metav1.Duration{Duration: 20 * time.Second},
			RemoteMediaProcessingEntityStabilizingDuration: &metav1.Duration{Duration: 5 * time.Second},
		},
	}
)

type WorkflowManagerConfigSpec struct {
	// Kubernetes cache configuration.
	CacheConfig KubernetesCacheConfig `json:"cache"`

	// LeaderElection is the LeaderElection config to be used when configuring
	// the manager.Manager leader election
	LeaderElection LeaderElectionConfig `json:"leaderElection,omitempty"`

	// GracefulShutdownTimeout is the duration given to runnable to stop before the manager actually returns on stop.
	// To disable graceful shutdown, set to time.Duration(0)
	// To use graceful shutdown without timeout, set to a negative duration, e.G. time.Duration(-1)
	// The graceful shutdown is skipped for safety reasons in case the leader election lease is lost.
	// +optional
	GracefulShutdownTimeout *metav1.Duration `json:"gracefulShutDown,omitempty"`

	// Controller contains global configuration options for controllers
	// registered within this manager.
	// +optional
	Controller *ControllerConfigSpec `json:"controller,omitempty"`

	// Metrics contains the controller metrics configuration
	// +optional
	Metrics *ControllerMetrics `json:"metrics,omitempty"`

	// Health contains the controller health configuration
	// +optional
	Health *ControllerHealth `json:"health,omitempty"`

	// Webhook contains the controllers webhook configuration
	// +optional
	Webhook *ControllerWebhook `json:"webhook,omitempty"`

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

// +kubebuilder:object:root=true

// WorkflowManagerConfig defines the configuration for nagare media engine controller manager.
type WorkflowManagerConfig struct {
	metav1.TypeMeta `json:",inline"`

	WorkflowManagerConfigSpec `json:",inline"`
}

func (c *WorkflowManagerConfig) Default() {
	c.doDefaultWithValuesFrom(DefaultWorkflowManagerConfig)
}

func (c *WorkflowManagerConfig) DefaultWithValuesFrom(d WorkflowManagerConfig) {
	c.doDefaultWithValuesFrom(d)
	c.doDefaultWithValuesFrom(DefaultWorkflowManagerConfig)
}

func (c *WorkflowManagerConfig) doDefaultWithValuesFrom(d WorkflowManagerConfig) {
	if c.WorkflowTerminationWaitingDuration == nil {
		c.WorkflowTerminationWaitingDuration = d.WorkflowTerminationWaitingDuration
	}
	if c.RemoteMediaProcessingEntityStabilizingDuration == nil {
		c.RemoteMediaProcessingEntityStabilizingDuration = d.RemoteMediaProcessingEntityStabilizingDuration
	}
	c.CacheConfig.DefaultWithValuesFrom(d.CacheConfig)
	c.LeaderElection.DefaultWithValuesFrom(d.LeaderElection)
}

func (c *WorkflowManagerConfig) Validate() error {
	return nil
}

func init() {
	SchemeBuilder.Register(&WorkflowManagerConfig{})
}
