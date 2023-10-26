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

package svc

import (
	"context"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	nbmpsvcv2 "github.com/nagare-media/engine/pkg/nbmp/svc/v2"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

type taskService struct {
	cfg *enginev1.TaskServiceConfiguration
}

var _ nbmpsvcv2.TaskService = &taskService{}

func NewTaskService(cfg *enginev1.TaskServiceConfiguration) *taskService {
	return &taskService{
		cfg: cfg,
	}
}

func (s *taskService) Create(ctx context.Context, t *nbmpv2.Task) error {
	panic("TODO: implement")
}

func (s *taskService) Update(ctx context.Context, t *nbmpv2.Task) error {
	panic("TODO: implement")
}

func (s *taskService) Delete(ctx context.Context, t *nbmpv2.Task) error {
	panic("TODO: implement")
}

func (s *taskService) Retrieve(ctx context.Context, t *nbmpv2.Task) error {
	panic("TODO: implement")
}
