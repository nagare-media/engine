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

package functions

import (
	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/models.go/base"
)

type SecretConfig struct {
	Workflow       Workflow                                `json:"workflow,omitempty"`
	Task           Task                                    `json:"task,omitempty"`
	MediaLocations map[string]enginev1.MediaLocationConfig `json:"mediaLocations,omitempty"`
	System         System                                  `json:"system,omitempty"`
}

type Workflow struct {
	Info   WorkflowInfo   `json:"info,omitempty"`
	Config map[string]any `json:"config,omitempty"`
}

type WorkflowInfo struct {
	Name string `json:"name,omitempty"`
}

type Task struct {
	Info   TaskInfo       `json:"info,omitempty"`
	Config map[string]any `json:"config,omitempty"`
}

type TaskInfo struct {
	Name string `json:"name,omitempty"`
}

type System struct {
	NATS NATS `json:"nats,omitempty"`
}

type NATS struct {
	URL      base.URI `json:"url,omitempty"`
	Username string   `json:"username,omitempty"`
	Password string   `json:"password,omitempty"`
	// TODO: add other auth methods
}

type GenericFunctionConfig struct {
	Inputs  map[string]any `json:"inputs,omitempty"`
	Outputs map[string]any `json:"outputs,omitempty"`
}

type MediaRef struct {
	URL string `json:"url,omitempty"`
}
