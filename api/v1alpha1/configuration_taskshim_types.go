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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

type TaskShimConfigurationSpec struct {
	Webserver   WebserverConfiguration   `json:"webserver"`
	TaskService TaskServiceConfiguration `json:"tasks"`
}

type TaskServiceConfiguration struct {
}

// +kubebuilder:object:root=true

// TaskShimConfiguration defines the configuration for nagare media engine task-shim.
type TaskShimConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	TaskShimConfigurationSpec `json:",inline"`
}

func init() {
	SchemeBuilder.Register(&TaskShimConfiguration{})
}

func (c *TaskShimConfiguration) Default() {
	if c.Webserver.BindAddress == nil {
		c.Webserver.BindAddress = ptr.To[string]("127.0.0.1:8888")
	}

	if c.Webserver.ReadTimeout == nil {
		c.Webserver.ReadTimeout = &metav1.Duration{Duration: time.Minute}
	}

	if c.Webserver.WriteTimeout == nil {
		c.Webserver.WriteTimeout = &metav1.Duration{} // = unlimited
	}

	if c.Webserver.IdleTimeout == nil {
		c.Webserver.IdleTimeout = &metav1.Duration{} // = unlimited
	}

	if c.Webserver.Network == nil {
		c.Webserver.Network = ptr.To[string]("tcp")
	}
}

func (c *TaskShimConfiguration) Validate() error {
	if c.Webserver.BindAddress == nil {
		return errors.New("missing webserver.bindAddress")
	}

	if c.Webserver.ReadTimeout == nil {
		return errors.New("missing webserver.readTimeout")
	}

	if c.Webserver.Network == nil {
		return errors.New("missing webserver.network")
	}

	if c.Webserver.PublicBaseURL != nil && strings.HasSuffix(*c.Webserver.PublicBaseURL, "/") {
		return errors.New("trailing slash in webserver.publicBaseURL")
	}
	return nil
}
