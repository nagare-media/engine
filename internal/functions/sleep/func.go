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

package sleep

import (
	"context"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/nagare-media/engine/internal/functions"
	"github.com/nagare-media/engine/pkg/nbmp"
	"github.com/nagare-media/engine/pkg/nbmp/utils"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

// Function description
const (
	Name = "sleep"

	DurationParameterKey = "sleep.engine.nagare.media/duration"
)

// Default configuration
var (
	DefaultDuration = 30 * time.Second
)

// function sleeps for a certain amount of time. It can be used for debugging.
type function struct {
	duration time.Duration
}

var _ nbmp.Function = &function{}

// Exec sleep function.
func (f *function) Exec(ctx context.Context) error {
	l := log.FromContext(ctx).
		WithName(Name).
		WithValues("duration", f.duration)

	l.Info("going to sleep")
	select {
	case <-ctx.Done():
		l.Error(ctx.Err(), "sleep disrupted")
	case <-time.After(f.duration):
		l.Info("woke up")
	}

	return nil
}

// BuildTask from sleep function.
func BuildTask(ctx context.Context, t *nbmpv2.Task) (nbmp.Function, error) {
	f := &function{
		duration: DefaultDuration,
	}

	if d, ok := utils.GetStringParameterValue(t.Configuration.Parameters, DurationParameterKey); ok {
		var err error
		f.duration, err = time.ParseDuration(d)
		if err != nil {
			return nil, err
		}
	}

	return f, nil
}

func init() {
	functions.TaskBuilders.Register(Name, BuildTask)
}
