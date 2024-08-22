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
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	metaaction "github.com/nagare-media/engine/internal/task-shim/actions/meta"
	"github.com/nagare-media/engine/pkg/strobj"
)

type TaskShimConfigurationSpec struct {
	// TaskService configuration.
	TaskService TaskShimTaskServiceConfiguration `json:"task"`

	// Webserver configuration.
	Webserver WebserverConfiguration `json:"webserver"`
}

type TaskShimTaskServiceConfiguration struct {
	// Actions that define the task execution.
	Actions []TaskServiceAction `json:"actions"`

	// OnCreateActions are executed when a Create request was received. Defaults to the "start-task" meta action.
	// +optional
	OnCreateActions []TaskServiceAction `json:"onCreate,omitempty"`

	// OnUpdateActions are executed when a Update request was received. Defaults to the "restart-task" meta action.
	// +optional
	OnUpdateActions []TaskServiceAction `json:"onUpdate,omitempty"`

	// OnDeleteActions are executed when a Delete request was received. Defaults to the "stop-task" meta action.
	// +optional
	OnDeleteActions []TaskServiceAction `json:"onDelete,omitempty"`

	// CreateTimeout is the duration after which the process will terminate if no Create request is ever made. Defaults
	// to "2m".
	// +optional
	CreateTimeout *metav1.Duration `json:"createTimeout,omitempty"`

	// DeleteTimeout is the duration after which the process will terminate if the task has stopped and no Delete request
	// is ever made. Defaults to "5m".
	// +optional
	DeleteTimeout *metav1.Duration `json:"deleteTimeout,omitempty"`
}

type TaskServiceAction struct {
	// Name that is user defined and human readable. Might be blank.
	Name string `json:"name"`

	// Action that should be executed.
	Action string `json:"action"`

	// Config options for this action.
	// +optional
	Config *strobj.StringOrObject `json:"config,omitempty"`
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
	if len(c.TaskService.OnCreateActions) == 0 {
		c.TaskService.OnCreateActions = append(c.TaskService.OnCreateActions, TaskServiceAction{
			Action: metaaction.Name,
			Config: &strobj.StringOrObject{
				Type:   strobj.String,
				StrVal: string(metaaction.ConfigTypeStartTask),
			},
		})
	}

	if len(c.TaskService.OnUpdateActions) == 0 {
		c.TaskService.OnUpdateActions = append(c.TaskService.OnUpdateActions, TaskServiceAction{
			Action: metaaction.Name,
			Config: &strobj.StringOrObject{
				Type:   strobj.String,
				StrVal: string(metaaction.ConfigTypeRestartTask),
			},
		})
	}

	if len(c.TaskService.OnDeleteActions) == 0 {
		c.TaskService.OnDeleteActions = append(c.TaskService.OnDeleteActions, TaskServiceAction{
			Action: metaaction.Name,
			Config: &strobj.StringOrObject{
				Type:   strobj.String,
				StrVal: string(metaaction.ConfigTypeStopTask),
			},
		})
	}

	if c.TaskService.CreateTimeout == nil {
		c.TaskService.CreateTimeout = &metav1.Duration{Duration: 2 * time.Minute}
	}

	if c.TaskService.DeleteTimeout == nil {
		c.TaskService.DeleteTimeout = &metav1.Duration{Duration: 5 * time.Minute}
	}

	if c.Webserver.BindAddress == nil {
		c.Webserver.BindAddress = ptr.To(":8888")
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
		c.Webserver.Network = ptr.To("tcp")
	}

	if c.Webserver.PublicBaseURL == nil {
		c.Webserver.PublicBaseURL = ptr.To("http://127.0.0.1:8888")
	}
}

func (c *TaskShimConfiguration) Validate() error {
	if len(c.TaskService.Actions) == 0 {
		return errors.New("missing task.actions")
	}

	if len(c.TaskService.OnCreateActions) == 0 {
		return errors.New("missing task.onCreateActions")
	}

	if len(c.TaskService.OnUpdateActions) == 0 {
		return errors.New("missing task.onUpdateActions")
	}

	if len(c.TaskService.OnDeleteActions) == 0 {
		return errors.New("missing task.onDeleteActions")
	}

	if c.TaskService.CreateTimeout == nil {
		return errors.New("missing task.createTimeout")
	}

	if c.TaskService.DeleteTimeout == nil {
		return errors.New("missing task.deleteTimeout")
	}

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
