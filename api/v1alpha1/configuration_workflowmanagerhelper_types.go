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
	"errors"
	"net/url"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

type WorkflowManagerHelperConfigurationSpec struct {
	// Task controller configuration.
	TaskController WorkflowManagerHelperTaskControllerConfiguration `json:"task"`

	// Reports controller configuration.
	ReportsController WorkflowManagerHelperReportsControllerConfiguration `json:"reports"`
}

type WorkflowManagerHelperTaskControllerConfiguration struct {
	// TaskAPI HTTP URL.
	TaskAPI string `json:"taskAPI"`

	// CreateRequestTimeout is the timeout used for Create requests. Defaults to "3m".
	// +optional
	CreateRequestTimeout *metav1.Duration `json:"createRequestTimeout,omitempty"`

	// RetrieveRequestTimeout is the timeout used for Retrieve requests (i.e. task probes). Defaults to "10s".
	// +optional
	RetrieveRequestTimeout *metav1.Duration `json:"retrieveRequestTimeout,omitempty"`

	// DeleteRequestTimeout is the timeout used for Delete requests. Defaults to "10m".
	// +optional
	DeleteRequestTimeout *metav1.Duration `json:"deleteRequestTimeout,omitempty"`

	// ObservePeriode is the period the task is probed to retrieve the current state. Defaults to "2s".
	// +optional
	ObservePeriode *metav1.Duration `json:"observePeriode,omitempty"`

	// MaxFailedProbes indicates the maximum number of consecutive failed probes after which workflow-manager-helper will
	// terminate with an error. Defaults to "10".
	// +optional
	MaxFailedProbes *int `json:"maxFailedProbes,omitempty"`
}

type WorkflowManagerHelperReportsControllerConfiguration struct {
	// Webserver configuration.
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
	if c.TaskController.CreateRequestTimeout == nil {
		c.TaskController.CreateRequestTimeout = &metav1.Duration{Duration: 3 * time.Minute}
	}

	if c.TaskController.RetrieveRequestTimeout == nil {
		c.TaskController.RetrieveRequestTimeout = &metav1.Duration{Duration: 10 * time.Second}
	}

	if c.TaskController.DeleteRequestTimeout == nil {
		c.TaskController.DeleteRequestTimeout = &metav1.Duration{Duration: 10 * time.Minute}
	}

	if c.TaskController.ObservePeriode == nil {
		c.TaskController.ObservePeriode = &metav1.Duration{Duration: 2 * time.Second}
	}

	if c.TaskController.MaxFailedProbes == nil {
		c.TaskController.MaxFailedProbes = ptr.To(10)
	}

	if c.ReportsController.Webserver.BindAddress == nil {
		c.ReportsController.Webserver.BindAddress = ptr.To(":8181")
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
		c.ReportsController.Webserver.Network = ptr.To("tcp")
	}

	if c.ReportsController.Webserver.PublicBaseURL == nil {
		c.ReportsController.Webserver.PublicBaseURL = ptr.To("http://127.0.0.1:8181")
	}
}

func (c *WorkflowManagerHelperConfiguration) Validate() error {
	if c.TaskController.TaskAPI == "" {
		return errors.New("missing task.taskAPI")
	}

	TaskAPIURL, err := url.Parse(c.TaskController.TaskAPI)
	if err != nil {
		return errors.New("task.taskAPI is not a URL")
	}

	if TaskAPIURL.Scheme != "http" && TaskAPIURL.Scheme != "https" {
		return errors.New("task.taskAPI is not an HTTP URL")
	}

	if c.TaskController.CreateRequestTimeout == nil {
		return errors.New("missing task.createRequestTimeout")
	}

	if c.TaskController.RetrieveRequestTimeout == nil {
		return errors.New("missing task.retrieveRequestTimeout")
	}

	if c.TaskController.DeleteRequestTimeout == nil {
		return errors.New("missing task.deleteRequestTimeout")
	}

	if c.TaskController.ObservePeriode == nil {
		return errors.New("missing task.observePeriode")
	}

	if c.TaskController.MaxFailedProbes == nil {
		return errors.New("missing task.maxFailedProbes")
	}

	if c.ReportsController.Webserver.BindAddress == nil {
		return errors.New("missing reports.webserver.bindAddress")
	}

	if c.ReportsController.Webserver.ReadTimeout == nil {
		return errors.New("missing reports.webserver.readTimeout")
	}

	if c.ReportsController.Webserver.WriteTimeout == nil {
		return errors.New("missing reports.webserver.writeTimeout")
	}

	if c.ReportsController.Webserver.IdleTimeout == nil {
		return errors.New("missing reports.webserver.idleTimeout")
	}

	if c.ReportsController.Webserver.Network == nil {
		return errors.New("missing reports.webserver.network")
	}

	if c.ReportsController.Webserver.PublicBaseURL != nil && strings.HasSuffix(*c.ReportsController.Webserver.PublicBaseURL, "/") {
		return errors.New("trailing slash in webserver.publicBaseURL")
	}

	return nil
}
