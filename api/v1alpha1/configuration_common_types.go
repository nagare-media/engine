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
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	configv1alpha1 "k8s.io/component-base/config/v1alpha1"
	"k8s.io/utils/ptr"
)

const (
	DefaultLeaderElectionID = "a3df9e9e.engine.nagare.media"
)

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

type KubernetesCacheConfiguration struct {
	// SyncPeriod determines the minimum frequency at which watched resources are
	// reconciled. A lower period will correct entropy more quickly, but reduce
	// responsiveness to change if there are many watched resources. Change this
	// value only if you know what you are doing. Defaults to 10 hours if unset.
	// there will a 10 percent jitter between the SyncPeriod of all controllers
	// so that all controllers will not send list requests simultaneously.
	// +optional
	SyncPeriod *metav1.Duration `json:"syncPeriod,omitempty"`

	// Namespace if specified restricts the manager's cache to watch objects in
	// the desired namespace Defaults to all namespaces
	//
	// Note: If a namespace is specified, controllers can still Watch for a
	// cluster-scoped resource (e.g Node).  For namespaced resources the cache
	// will only hold objects from the desired namespace.
	// +optional
	Namespace *string `json:"Namespace,omitempty"`
}

func (c *KubernetesCacheConfiguration) Default() {
	if c.SyncPeriod == nil {
		c.SyncPeriod = &metav1.Duration{Duration: 10 * time.Hour}
	}
}

// ControllerConfigurationSpec defines the global configuration for
// controllers registered with the manager.
type ControllerConfigurationSpec struct {
	// GroupKindConcurrency is a map from a Kind to the number of concurrent reconciliation
	// allowed for that controller.
	//
	// When a controller is registered within this manager using the builder utilities,
	// users have to specify the type the controller reconciles in the For(...) call.
	// If the object's kind passed matches one of the keys in this map, the concurrency
	// for that controller is set to the number specified.
	//
	// The key is expected to be consistent in form with GroupKind.String(),
	// e.g. ReplicaSet in apps group (regardless of version) would be `ReplicaSet.apps`.
	// +optional
	GroupKindConcurrency map[string]int `json:"groupKindConcurrency,omitempty"`

	// MaxConcurrentReconciles is the maximum number of concurrent Reconciles which can be run. Defaults to 1.
	// +optional
	MaxConcurrentReconciles *int `json:"maxConcurrentReconciles,omitempty"`

	// CacheSyncTimeout refers to the time limit set to wait for syncing caches.
	// Defaults to 2 minutes if not set.
	// +optional
	CacheSyncTimeout time.Duration `json:"cacheSyncTimeout,omitempty"`

	// RecoverPanic indicates whether the panic caused by reconcile should be recovered.
	// Defaults to the Controller.RecoverPanic setting from the Manager if unset.
	// +optional
	RecoverPanic *bool `json:"recoverPanic,omitempty"`

	// NeedLeaderElection indicates whether the controller needs to use leader election.
	// Defaults to true, which means the controller will use leader election.
	// +optional
	NeedLeaderElection *bool `json:"needLeaderElection,omitempty"`
}

// ControllerMetrics defines the metrics configs.
type ControllerMetrics struct {
	// BindAddress is the TCP address that the controller should bind to
	// for serving prometheus metrics.
	// It can be set to "0" to disable the metrics serving.
	// +optional
	BindAddress string `json:"bindAddress,omitempty"`
}

// ControllerHealth defines the health configs.
type ControllerHealth struct {
	// HealthProbeBindAddress is the TCP address that the controller should bind to
	// for serving health probes
	// It can be set to "0" or "" to disable serving the health probe.
	// +optional
	HealthProbeBindAddress string `json:"healthProbeBindAddress,omitempty"`

	// ReadinessEndpointName, defaults to "readyz"
	// +optional
	ReadinessEndpointName string `json:"readinessEndpointName,omitempty"`

	// LivenessEndpointName, defaults to "healthz"
	// +optional
	LivenessEndpointName string `json:"livenessEndpointName,omitempty"`
}

// ControllerWebhook defines the webhook server for the controller.
type ControllerWebhook struct {
	// Port is the port that the webhook server serves at.
	// It is used to set webhook.Server.Port.
	// +optional
	Port *int `json:"port,omitempty"`

	// Host is the hostname that the webhook server binds to.
	// It is used to set webhook.Server.Host.
	// +optional
	Host string `json:"host,omitempty"`

	// CertDir is the directory that contains the server key and certificate.
	// if not set, webhook server would look up the server key and certificate in
	// {TempDir}/k8s-webhook-server/serving-certs. The server key and certificate
	// must be named tls.key and tls.crt, respectively.
	// +optional
	CertDir string `json:"certDir,omitempty"`
}

func DefaultLeaderElection(l *configv1alpha1.LeaderElectionConfiguration) {
	if l.LeaderElect == nil {
		l.LeaderElect = ptr.To(false)
	}
	if l.LeaseDuration.Duration == 0 {
		l.LeaseDuration.Duration = 15 * time.Second
	}
	if l.RenewDeadline.Duration == 0 {
		l.RenewDeadline.Duration = 10 * time.Second
	}
	if l.RetryPeriod.Duration == 0 {
		l.RetryPeriod.Duration = 2 * time.Second
	}
	if l.ResourceLock == "" {
		l.ResourceLock = resourcelock.LeasesResourceLock
	}
	if l.ResourceName == "" {
		l.ResourceName = DefaultLeaderElectionID
	}
}
