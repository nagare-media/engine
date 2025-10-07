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

package meta

import (
	"encoding/json"
	"errors"

	"k8s.io/utils/ptr"

	"github.com/nagare-media/engine/internal/task-shim/actions"
)

// Action description
const (
	BaseName = "meta"
	Name     = actions.BuildInActionPrefix + BaseName
)

// ConfigType of the meta action.
type ConfigType string

var _ json.Unmarshaler = ptr.To(ConfigType(""))

const (
	StartTaskConfigType   = ConfigType("start-task")
	RestartTaskConfigType = ConfigType("restart-task")
	StopTaskConfigType    = ConfigType("stop-task")
)

func (ct *ConfigType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	switch c := ConfigType(s); c {
	case StartTaskConfigType,
		RestartTaskConfigType,
		StopTaskConfigType:
		*ct = c
		return nil
	}

	return errors.New("meta: unknown config type")
}
