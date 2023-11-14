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

type WorkflowManagerHelperConfigurationSpec struct {
	TaskController    WorkflowManagerHelperTaskControllerConfiguration    `json:"task"`
	ReportsController WorkflowManagerHelperReportsControllerConfiguration `json:"reports"`
}

type WorkflowManagerHelperTaskControllerConfiguration struct {
}

type WorkflowManagerHelperReportsControllerConfiguration struct {
	Webserver WebserverConfiguration `json:"webserver"`
}

// +kubebuilder:object:root=true

// WorkflowManagerHelperConfiguration defines the configuration for nagare media engine workflow-manager-helper.
type WorkflowManagerHelperConfiguration struct {
	metav1.TypeMeta `json:",inline"`

	WorkflowManagerHelperConfigurationSpec `json:",inline"`
}

func init() {
	SchemeBuilder.Register(&WorkflowManagerHelperConfiguration{})
}

func (c *WorkflowManagerHelperConfiguration) Default() {
	if c.ReportsController.Webserver.BindAddress == nil {
		c.ReportsController.Webserver.BindAddress = ptr.To[string]("127.0.0.1:8181")
	}

	if c.ReportsController.Webserver.ReadTimeout == nil {
		c.ReportsController.Webserver.ReadTimeout = &metav1.Duration{Duration: time.Minute}
	}

	if c.ReportsController.Webserver.WriteTimeout == nil {
		c.ReportsController.Webserver.WriteTimeout = &metav1.Duration{Duration: time.Minute}
	}

	if c.ReportsController.Webserver.IdleTimeout == nil {
		c.ReportsController.Webserver.IdleTimeout = &metav1.Duration{Duration: time.Minute}
	}

	if c.ReportsController.Webserver.Network == nil {
		c.ReportsController.Webserver.Network = ptr.To[string]("tcp")
	}
}

func (c *WorkflowManagerHelperConfiguration) Validate() error {
	if c.ReportsController.Webserver.BindAddress == nil {
		return errors.New("missing webserver.bindAddress")
	}

	if c.ReportsController.Webserver.ReadTimeout == nil {
		return errors.New("missing webserver.readTimeout")
	}

	if c.ReportsController.Webserver.WriteTimeout == nil {
		return errors.New("missing webserver.writeTimeout")
	}

	if c.ReportsController.Webserver.IdleTimeout == nil {
		return errors.New("missing webserver.idleTimeout")
	}

	if c.ReportsController.Webserver.Network == nil {
		return errors.New("missing webserver.network")
	}

	if c.ReportsController.Webserver.PublicBaseURL != nil && strings.HasSuffix(*c.ReportsController.Webserver.PublicBaseURL, "/") {
		return errors.New("trailing slash in webserver.publicBaseURL")
	}

	return nil
}
