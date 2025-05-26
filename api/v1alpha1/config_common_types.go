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
	"fmt"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	configv1alpha1 "k8s.io/component-base/config/v1alpha1"
	"k8s.io/utils/ptr"
	ctrlcfg "sigs.k8s.io/controller-runtime/pkg/config"

	"github.com/nagare-media/models.go/base"
)

var (
	DefaultNATSConfig = NATSConfig{
		// TODO: set useful default
		URL: "nats://nats.nats.svc.cluster.local:4222",
	}
)

type NATSConfig struct {
	// TODO: add auth methods
	URL base.URI `json:"url"`
}

func (c *NATSConfig) Default() {
	c.doDefaultWithValuesFrom(DefaultNATSConfig)
}

func (c *NATSConfig) DefaultWithValuesFrom(d NATSConfig) {
	c.doDefaultWithValuesFrom(d)
	c.doDefaultWithValuesFrom(DefaultNATSConfig)
}

func (c *NATSConfig) doDefaultWithValuesFrom(d NATSConfig) {
	if c.URL == "" {
		c.URL = d.URL
	}
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
	DefaultCacheConfig = CacheConfig{
		SyncPeriod: &metav1.Duration{Duration: 10 * time.Hour},
	}
)

type CacheConfig struct {
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

func (c *CacheConfig) Default() {
	c.doDefaultWithValuesFrom(DefaultCacheConfig)
}

func (c *CacheConfig) DefaultWithValuesFrom(d CacheConfig) {
	c.doDefaultWithValuesFrom(d)
	c.doDefaultWithValuesFrom(DefaultCacheConfig)
}

func (c *CacheConfig) doDefaultWithValuesFrom(d CacheConfig) {
	if c.SyncPeriod == nil {
		c.SyncPeriod = d.SyncPeriod
	}
}

var (
	DefaultControllerConfig = ControllerConfig{
		SkipNameValidation:      ptr.To(false),
		GroupKindConcurrency:    nil,
		MaxConcurrentReconciles: 1,
		CacheSyncTimeout:        2 * time.Minute,
		RecoverPanic:            nil,
		NeedLeaderElection:      nil,
	}
)

// ControllerConfig defines the global configuration for controllers registered with the manager.
// +kubebuilder:object:generate=false
type ControllerConfig ctrlcfg.Controller

// TODO: Fix DeepCopyInto for ctrlcfg.Controller (Logger has no DeepCopyInto)
func (in *ControllerConfig) DeepCopyInto(out *ControllerConfig) {
	*out = *in
	if in.SkipNameValidation != nil {
		in, out := &in.SkipNameValidation, &out.SkipNameValidation
		*out = new(bool)
		**out = **in
	}
	if in.GroupKindConcurrency != nil {
		in, out := &in.GroupKindConcurrency, &out.GroupKindConcurrency
		*out = make(map[string]int, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.RecoverPanic != nil {
		in, out := &in.RecoverPanic, &out.RecoverPanic
		*out = new(bool)
		**out = **in
	}
	if in.NeedLeaderElection != nil {
		in, out := &in.NeedLeaderElection, &out.NeedLeaderElection
		*out = new(bool)
		**out = **in
	}
	if in.UsePriorityQueue != nil {
		in, out := &in.UsePriorityQueue, &out.UsePriorityQueue
		*out = new(bool)
		**out = **in
	}
	// in.Logger.DeepCopyInto(&out.Logger)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ControllerConfig.
func (in *ControllerConfig) DeepCopy() *ControllerConfig {
	if in == nil {
		return nil
	}
	out := new(ControllerConfig)
	in.DeepCopyInto(out)
	return out
}

func (c *ControllerConfig) Default() {
	c.doDefaultWithValuesFrom(DefaultControllerConfig)
}

func (c *ControllerConfig) DefaultWithValuesFrom(d ControllerConfig) {
	c.doDefaultWithValuesFrom(d)
	c.doDefaultWithValuesFrom(DefaultControllerConfig)
}

func (c *ControllerConfig) doDefaultWithValuesFrom(d ControllerConfig) {
	if c.SkipNameValidation == nil {
		c.SkipNameValidation = d.SkipNameValidation
	}
	if c.GroupKindConcurrency == nil {
		c.GroupKindConcurrency = d.GroupKindConcurrency
	}
	if c.MaxConcurrentReconciles == 0 {
		c.MaxConcurrentReconciles = d.MaxConcurrentReconciles
	}
	if c.CacheSyncTimeout == 0 {
		c.CacheSyncTimeout = d.CacheSyncTimeout
	}
	if c.RecoverPanic == nil {
		c.RecoverPanic = d.RecoverPanic
	}
	if c.NeedLeaderElection == nil {
		c.NeedLeaderElection = d.NeedLeaderElection
	}
}

var (
	DefaultMetricsConfig = MetricsConfig{
		BindAddress: ":8080",
	}
)

// MetricsConfig defines the metrics configs.
type MetricsConfig struct {
	// BindAddress is the TCP address that the controller should bind to
	// for serving prometheus metrics.
	// It can be set to "0" to disable the metrics serving.
	// +optional
	BindAddress string `json:"bindAddress,omitempty"`
}

func (c *MetricsConfig) Default() {
	c.doDefaultWithValuesFrom(DefaultMetricsConfig)
}

func (c *MetricsConfig) DefaultWithValuesFrom(d MetricsConfig) {
	c.doDefaultWithValuesFrom(d)
	c.doDefaultWithValuesFrom(DefaultMetricsConfig)
}

func (c *MetricsConfig) doDefaultWithValuesFrom(d MetricsConfig) {
	if c.BindAddress == "" {
		c.BindAddress = d.BindAddress
	}
}

var (
	DefaultHealthConfig = HealthConfig{
		BindAddress:           ":8081",
		ReadinessEndpointName: "/readyz",
		LivenessEndpointName:  "/healthz",
	}
)

// HealthConfig defines the health configs.
type HealthConfig struct {
	// BindAddress is the TCP address that the controller should bind to
	// for serving health probes
	// It can be set to "0" or "" to disable serving the health probe.
	// +optional
	BindAddress string `json:"bindAddress,omitempty"`

	// ReadinessEndpointName, defaults to "readyz"
	// +optional
	ReadinessEndpointName string `json:"readinessEndpointName,omitempty"`

	// LivenessEndpointName, defaults to "healthz"
	// +optional
	LivenessEndpointName string `json:"livenessEndpointName,omitempty"`
}

func (c *HealthConfig) Default() {
	c.doDefaultWithValuesFrom(DefaultHealthConfig)
}

func (c *HealthConfig) DefaultWithValuesFrom(d HealthConfig) {
	c.doDefaultWithValuesFrom(d)
	c.doDefaultWithValuesFrom(DefaultHealthConfig)
}

func (c *HealthConfig) doDefaultWithValuesFrom(d HealthConfig) {
	if c.BindAddress == "" {
		c.BindAddress = d.BindAddress
	}
	if c.ReadinessEndpointName == "" {
		c.ReadinessEndpointName = d.ReadinessEndpointName
	}
	if c.LivenessEndpointName == "" {
		c.LivenessEndpointName = d.LivenessEndpointName
	}
}

var (
	DefaultWebhookConfig = WebhookConfig{
		Port:    ptr.To(9443),
		Host:    "",
		CertDir: "",
	}
)

// WebhookConfig defines the webhook server for the controller.
type WebhookConfig struct {
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

func (c *WebhookConfig) Default() {
	c.doDefaultWithValuesFrom(DefaultWebhookConfig)
}

func (c *WebhookConfig) DefaultWithValuesFrom(d WebhookConfig) {
	c.doDefaultWithValuesFrom(d)
	c.doDefaultWithValuesFrom(DefaultWebhookConfig)
}

func (c *WebhookConfig) doDefaultWithValuesFrom(d WebhookConfig) {
	if c.Port == nil {
		c.Port = d.Port
	}
	if c.Host == "" {
		c.Host = d.Host
	}
	if c.CertDir == "" {
		c.CertDir = d.CertDir
	}
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
