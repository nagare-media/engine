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

package file

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/nagare-media/engine/internal/task-shim/actions"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

// Action description
const (
	BaseName = "file"
	Name     = actions.BuildInActionPrefix + BaseName
)

var (
	// TODO: make this configurable
	DefaultMode = os.FileMode(0640)
)

// action file manipulates a file on the filesystem.
type action struct {
	cfg Config
}

// Config for file action
type Config struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

var _ actions.Action = &action{}

// Exec file action.
func (a *action) Exec(ctx actions.Context) (*nbmpv2.Task, error) {
	err := os.WriteFile(a.cfg.Path, []byte(a.cfg.Content), DefaultMode)
	if err != nil {
		return ctx.Task, err
	}
	return ctx.Task, nil
}

// Validate file action config.
func (c *Config) Validate() error {
	if c.Path == "" {
		return errors.New("file: path is missing")
	}

	return nil
}

// BuildAction instance of file action.
func BuildAction(ctx actions.Context) (actions.Action, error) {
	if ctx.Config == nil {
		return nil, errors.New("file: not configured")
	}

	c := Config{}
	if err := json.Unmarshal(ctx.Config, &c); err != nil {
		return nil, err
	}

	if err := c.Validate(); err != nil {
		return nil, err
	}

	return &action{cfg: c}, nil
}

func init() {
	actions.ActionBuilders.Register(Name, BuildAction)
}
