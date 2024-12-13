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

package exec

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/nagare-media/engine/internal/task-shim/actions"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

// Action description
const (
	BaseName = "exec"
	Name     = actions.BuildInActionPrefix + BaseName
)

var (
	// TODO: make this configurable
	DefaultStopSignal             = os.Interrupt
	DefaultTerminationGracePeriod = 10 * time.Second
)

// action exec executes a command.
type action struct {
	cfg Config
}

var _ actions.Action = &action{}

// Config for exec action
type Config struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env"`
	WD      string            `json:"workingDirectory"`
}

// Exec exec action.
func (a *action) Exec(ctx actions.Context) (*nbmpv2.Task, error) {
	l := log.FromContext(ctx.Ctx).WithName(BaseName)

	// create cmd
	cmd := exec.CommandContext(ctx.Ctx, a.cfg.Command, a.cfg.Args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout // redirect all to stdout

	// after ctx is canceled we send an interrupt signal + kill after a timeout
	cmd.Cancel = func() error {
		l.Info("requested to terminate process")

		if runtime.GOOS == "windows" {
			return cmd.Process.Signal(os.Kill)
		}
		go func() {
			time.Sleep(DefaultTerminationGracePeriod)
			_ = cmd.Process.Signal(os.Kill)
		}()

		return cmd.Process.Signal(DefaultStopSignal)
	}

	// working directory
	if a.cfg.WD != "" {
		cmd.Dir = a.cfg.WD
	}

	// env
	env := os.Environ()
	for k, v := range a.cfg.Env {
		env = append(env, k+"="+v)
	}
	cmd.Env = env

	// start command
	l.Info("start process")
	if err := cmd.Run(); err != nil {
		if ec := cmd.ProcessState.ExitCode(); ec > 0 {
			l.Error(err, fmt.Sprintf("process terminated with error code %d", ec))
		} else {
			l.Error(err, "process terminated with error")
		}
		return ctx.Task, err
	} else {
		l.Info("process terminated successfully")
	}

	// TODO: allow for changes to task from cmd
	return ctx.Task, nil
}

// Validate exec action config.
func (c *Config) Validate() error {
	if c.Command == "" {
		return errors.New("exec: command not set")
	}

	return nil
}

// BuildAction instance of exec action.
func BuildAction(ctx actions.Context) (actions.Action, error) {
	if ctx.Config == nil {
		return nil, errors.New("exec: not configured")
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
