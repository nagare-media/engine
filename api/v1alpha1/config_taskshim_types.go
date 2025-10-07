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
	"errors"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	metaaction "github.com/nagare-media/engine/internal/task-shim/actions/meta"
	"github.com/nagare-media/engine/pkg/apis/strobj"
)

var (
	DefaultTaskShimConfig = TaskShimConfig{
		TaskShimConfigSpec: TaskShimConfigSpec{
			TaskService: TaskShimTaskServiceConfig{
				OnCreateActions: []TaskServiceAction{{
					Action: metaaction.Name,
					Config: &strobj.StringOrObject{
						Type:   strobj.String,
						StrVal: string(metaaction.StartTaskConfigType),
					},
				}},
				OnUpdateActions: []TaskServiceAction{{
					Action: metaaction.Name,
					Config: &strobj.StringOrObject{
						Type:   strobj.String,
						StrVal: string(metaaction.RestartTaskConfigType),
					},
				}},
				OnDeleteActions: []TaskServiceAction{{
					Action: metaaction.Name,
					Config: &strobj.StringOrObject{
						Type:   strobj.String,
						StrVal: string(metaaction.StopTaskConfigType),
					},
				}},
				CreateTimeout: &metav1.Duration{Duration: 1 * time.Minute},
				DeleteTimeout: &metav1.Duration{Duration: 1 * time.Minute},
			},
			Webserver: WebserverConfig{
				BindAddress:   ptr.To(":8888"),
				WriteTimeout:  &metav1.Duration{}, // = unlimited
				IdleTimeout:   &metav1.Duration{}, // = unlimited
				PublicBaseURL: ptr.To("http://127.0.0.1:8888"),
			},
		},
	}
)

type TaskShimConfigSpec struct {
	// TaskService configuration.
	TaskService TaskShimTaskServiceConfig `json:"task"`

	// Webserver configuration.
	Webserver WebserverConfig `json:"webserver"`
}

type TaskShimTaskServiceConfig struct {
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
	// to "1m".
	// +optional
	CreateTimeout *metav1.Duration `json:"createTimeout,omitempty"`

	// DeleteTimeout is the duration after which the process will terminate if the task has stopped and no Delete request
	// is ever made. Defaults to "1m".
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

// TaskShimConfig defines the configuration for nagare media engine task-shim.
type TaskShimConfig struct {
	metav1.TypeMeta `json:",inline"`

	TaskShimConfigSpec `json:",inline"`
}

func (c *TaskShimConfig) Default() {
	c.doDefaultWithValuesFrom(DefaultTaskShimConfig)
}

func (c *TaskShimConfig) DefaultWithValuesFrom(d TaskShimConfig) {
	c.doDefaultWithValuesFrom(d)
	c.doDefaultWithValuesFrom(DefaultTaskShimConfig)
}

func (c *TaskShimConfig) doDefaultWithValuesFrom(d TaskShimConfig) {
	if len(c.TaskService.OnCreateActions) == 0 {
		c.TaskService.OnCreateActions = d.TaskService.OnCreateActions
	}
	if len(c.TaskService.OnUpdateActions) == 0 {
		c.TaskService.OnUpdateActions = d.TaskService.OnUpdateActions
	}
	if len(c.TaskService.OnDeleteActions) == 0 {
		c.TaskService.OnDeleteActions = d.TaskService.OnDeleteActions
	}
	if c.TaskService.CreateTimeout == nil {
		c.TaskService.CreateTimeout = d.TaskService.CreateTimeout
	}
	if c.TaskService.DeleteTimeout == nil {
		c.TaskService.DeleteTimeout = d.TaskService.DeleteTimeout
	}
	c.Webserver.DefaultWithValuesFrom(d.Webserver)
}

func (c *TaskShimConfig) Validate() error {
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
	if err := c.Webserver.Validate("webserver."); err != nil {
		return err
	}
	return nil
}

func init() {
	SchemeBuilder.Register(&TaskShimConfig{})
}
