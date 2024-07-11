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
	Webserver   WebserverConfiguration           `json:"webserver"`
	TaskService TaskShimTaskServiceConfiguration `json:"task"`
}

type TaskShimTaskServiceConfiguration struct {
	Actions         []TaskServiceAction `json:"actions"`
	OnCreateActions []TaskServiceAction `json:"onCreate"`
	OnUpdateActions []TaskServiceAction `json:"onUpdate"`
	OnDeleteActions []TaskServiceAction `json:"onDelete"`
}

type TaskServiceAction struct {
	Name   string                 `json:"name"`
	Action string                 `json:"action"`
	Config *strobj.StringOrObject `json:"config"`
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

	if len(c.TaskService.Actions) == 0 {
		return errors.New("missing task actions")
	}

	return nil
}
