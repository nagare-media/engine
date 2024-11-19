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
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

var (
	DefaultWorkflowManagerHelperConfig = WorkflowManagerHelperConfig{
		WorkflowManagerHelperConfigSpec: WorkflowManagerHelperConfigSpec{
			TaskController: WorkflowManagerHelperTaskControllerConfig{
				CreateRequestTimeout:   &metav1.Duration{Duration: 3 * time.Minute},
				RetrieveRequestTimeout: &metav1.Duration{Duration: 10 * time.Second},
				UpdateRequestTimeout:   &metav1.Duration{Duration: 10 * time.Minute},
				DeleteRequestTimeout:   &metav1.Duration{Duration: 10 * time.Minute},
				ObservePeriode:         &metav1.Duration{Duration: 2 * time.Second},
				MaxFailedProbes:        ptr.To(10),
			},
			ReportsController: WorkflowManagerHelperReportsControllerConfig{
				Webserver: WebserverConfig{
					BindAddress:   ptr.To(":8181"),
					PublicBaseURL: ptr.To("http://127.0.0.1:8181"),
				},
			},
		},
	}
)

type WorkflowManagerHelperConfigSpec struct {
	// Task controller configuration.
	TaskController WorkflowManagerHelperTaskControllerConfig `json:"task"`

	// Reports controller configuration.
	ReportsController WorkflowManagerHelperReportsControllerConfig `json:"reports"`
}

type WorkflowManagerHelperTaskControllerConfig struct {
	// TaskAPI HTTP URL.
	TaskAPI string `json:"taskAPI"`

	// CreateRequestTimeout is the timeout used for Create requests. Defaults to "3m".
	// +optional
	CreateRequestTimeout *metav1.Duration `json:"createRequestTimeout,omitempty"`

	// RetrieveRequestTimeout is the timeout used for Retrieve requests (i.e. task probes). Defaults to "10s".
	// +optional
	RetrieveRequestTimeout *metav1.Duration `json:"retrieveRequestTimeout,omitempty"`

	// UpdateRequestTimeout is the timeout used for Update requests. Defaults to "10m".
	// +optional
	UpdateRequestTimeout *metav1.Duration `json:"updateRequestTimeout,omitempty"`

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

type WorkflowManagerHelperReportsControllerConfig struct {
	// Webserver configuration.
	Webserver WebserverConfig `json:"webserver"`
}

// +kubebuilder:object:root=true

// WorkflowManagerHelperConfig defines the configuration for nagare media engine workflow-manager-helper.
type WorkflowManagerHelperConfig struct {
	metav1.TypeMeta `json:",inline"`

	WorkflowManagerHelperConfigSpec `json:",inline"`
}

func (c *WorkflowManagerHelperConfig) Default() {
	c.doDefaultWithValuesFrom(DefaultWorkflowManagerHelperConfig)
}

func (c *WorkflowManagerHelperConfig) DefaultWithValuesFrom(d WorkflowManagerHelperConfig) {
	c.doDefaultWithValuesFrom(d)
	c.doDefaultWithValuesFrom(DefaultWorkflowManagerHelperConfig)
}

func (c *WorkflowManagerHelperConfig) doDefaultWithValuesFrom(d WorkflowManagerHelperConfig) {
	if c.TaskController.CreateRequestTimeout == nil {
		c.TaskController.CreateRequestTimeout = d.TaskController.CreateRequestTimeout
	}
	if c.TaskController.RetrieveRequestTimeout == nil {
		c.TaskController.RetrieveRequestTimeout = d.TaskController.RetrieveRequestTimeout
	}
	if c.TaskController.UpdateRequestTimeout == nil {
		c.TaskController.UpdateRequestTimeout = d.TaskController.UpdateRequestTimeout
	}
	if c.TaskController.DeleteRequestTimeout == nil {
		c.TaskController.DeleteRequestTimeout = d.TaskController.DeleteRequestTimeout
	}
	if c.TaskController.ObservePeriode == nil {
		c.TaskController.ObservePeriode = d.TaskController.ObservePeriode
	}
	if c.TaskController.MaxFailedProbes == nil {
		c.TaskController.MaxFailedProbes = d.TaskController.MaxFailedProbes
	}
	c.ReportsController.Webserver.DefaultWithValuesFrom(d.ReportsController.Webserver)
}

func (c *WorkflowManagerHelperConfig) Validate() error {
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
	if c.TaskController.UpdateRequestTimeout == nil {
		return errors.New("missing task.updateRequestTimeout")
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
	if err := c.ReportsController.Webserver.Validate("reports.webserver."); err != nil {
		return err
	}
	return nil
}

func init() {
	SchemeBuilder.Register(&WorkflowManagerHelperConfig{})
}
