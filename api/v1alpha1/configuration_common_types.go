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
	"fmt"
	"strings"
	"time"

	"github.com/nagare-media/models.go/base"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	configv1alpha1 "k8s.io/component-base/config/v1alpha1"
	"k8s.io/utils/ptr"
)

type NATSConfig struct {
	// TODO: add auth methods
	URL base.URI `json:"url"`
}

var (
	DefaultWebserverConfig = WebserverConfig{
		BindAddress:   ptr.To(":8080"),
		ReadTimeout:   &metav1.Duration{Duration: time.Minute},
		WriteTimeout:  &metav1.Duration{Duration: time.Minute},
		IdleTimeout:   &metav1.Duration{Duration: time.Minute},
		Network:       ptr.To("tcp"),
		PublicBaseURL: ptr.To("http://127.0.0.1:8080"),
	}
)

type WebserverConfig struct {
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

func (c *WebserverConfig) Default() {
	c.doDefaultWithValuesFrom(DefaultWebserverConfig)
}

func (c *WebserverConfig) DefaultWithValuesFrom(d WebserverConfig) {
	c.doDefaultWithValuesFrom(d)
	c.doDefaultWithValuesFrom(DefaultWebserverConfig)
}

func (c *WebserverConfig) doDefaultWithValuesFrom(d WebserverConfig) {
	if c.BindAddress == nil {
		c.BindAddress = d.BindAddress
	}
	if c.ReadTimeout == nil {
		c.ReadTimeout = d.ReadTimeout
	}
	if c.WriteTimeout == nil {
		c.WriteTimeout = d.WriteTimeout
	}
	if c.IdleTimeout == nil {
		c.IdleTimeout = d.IdleTimeout
	}
	if c.Network == nil {
		c.Network = d.Network
	}
	if c.PublicBaseURL == nil {
		c.PublicBaseURL = d.PublicBaseURL
	}
}

func (c *WebserverConfig) Validate(cfgPrefix string) error {
	if c.BindAddress == nil {
		return fmt.Errorf("missing %sbindAddress", cfgPrefix)
	}
	if c.ReadTimeout == nil {
		return fmt.Errorf("missing %sreadTimeout", cfgPrefix)
	}
	if c.WriteTimeout == nil {
		return fmt.Errorf("missing %swriteTimeout", cfgPrefix)
	}
	if c.IdleTimeout == nil {
		return fmt.Errorf("missing %sidleTimeout", cfgPrefix)
	}
	if c.Network == nil {
		return fmt.Errorf("missing %snetwork", cfgPrefix)
	}
	if c.PublicBaseURL != nil && strings.HasSuffix(*c.PublicBaseURL, "/") {
		return fmt.Errorf("trailing slash in %spublicBaseURL", cfgPrefix)
	}
	return nil
}

var (
	DefaultKubernetesCacheConfig = KubernetesCacheConfig{
		SyncPeriod: &metav1.Duration{Duration: 10 * time.Hour},
	}
)

type KubernetesCacheConfig struct {
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

func (c *KubernetesCacheConfig) Default() {
	c.doDefaultWithValuesFrom(DefaultKubernetesCacheConfig)
}

func (c *KubernetesCacheConfig) DefaultWithValuesFrom(d KubernetesCacheConfig) {
	c.doDefaultWithValuesFrom(d)
	c.doDefaultWithValuesFrom(DefaultKubernetesCacheConfig)
}

func (c *KubernetesCacheConfig) doDefaultWithValuesFrom(d KubernetesCacheConfig) {
	if c.SyncPeriod == nil {
		c.SyncPeriod = d.SyncPeriod
	}
}

// ControllerConfigSpec defines the global configuration for
// controllers registered with the manager.
type ControllerConfigSpec struct {
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

var (
	DefaultLeaderElectionConfig = LeaderElectionConfig{
		LeaderElect:   ptr.To(false),
		LeaseDuration: metav1.Duration{Duration: 15 * time.Second},
		RenewDeadline: metav1.Duration{Duration: 10 * time.Second},
		RetryPeriod:   metav1.Duration{Duration: 2 * time.Second},
		ResourceLock:  resourcelock.LeasesResourceLock,
		ResourceName:  "a3df9e9e.engine.nagare.media",
	}
)

type LeaderElectionConfig configv1alpha1.LeaderElectionConfiguration

func (c *LeaderElectionConfig) Default() {
	c.doDefaultWithValuesFrom(DefaultLeaderElectionConfig)
}

func (c *LeaderElectionConfig) DefaultWithValuesFrom(d LeaderElectionConfig) {
	c.doDefaultWithValuesFrom(d)
	c.doDefaultWithValuesFrom(DefaultLeaderElectionConfig)
}

func (c *LeaderElectionConfig) doDefaultWithValuesFrom(d LeaderElectionConfig) {
	if c.LeaderElect == nil {
		c.LeaderElect = d.LeaderElect
	}
	if c.LeaseDuration.Duration == 0 {
		c.LeaseDuration = d.LeaseDuration
	}
	if c.RenewDeadline.Duration == 0 {
		c.RenewDeadline = d.RenewDeadline
	}
	if c.RetryPeriod.Duration == 0 {
		c.RetryPeriod = d.RetryPeriod
	}
	if c.ResourceLock == "" {
		c.ResourceLock = d.ResourceLock
	}
	if c.ResourceName == "" {
		c.ResourceName = d.ResourceName
	}
}
