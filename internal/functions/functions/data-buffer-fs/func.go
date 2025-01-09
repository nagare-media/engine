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

package function

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/nagare-media/engine/internal/functions"
	"github.com/nagare-media/engine/pkg/nbmp"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

// Function description
const (
	Name = "data-buffer-fs"
)

// function buffers input stream.
type function struct {
}

var _ nbmp.Function = &function{}

// Exec data-buffer-fs function.
func (f *function) Exec(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	l := log.FromContext(ctx).WithName(Name)
	ctx = log.IntoContext(ctx, l)

	_ = ctx // TODO: implement

	return nil
}

// BuildTask from data-buffer-fs function.
func BuildTask(ctx context.Context, t *nbmpv2.Task) (nbmp.Function, error) {
	f := &function{}
	return f, nil
}

func init() {
	functions.TaskBuilders.Register(Name, BuildTask)
}
