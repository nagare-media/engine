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

package workflowmanagerhelper

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/log"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/pkg/starter"
)

type taskCtrl struct {
	cfg  *enginev1.WorkflowManagerHelperTaskControllerConfiguration
	data *enginev1.WorkflowManagerHelperData
}

var _ starter.Starter = &taskCtrl{}

func NewTaskController(cfg *enginev1.WorkflowManagerHelperTaskControllerConfiguration, data *enginev1.WorkflowManagerHelperData) starter.Starter {
	return &taskCtrl{
		cfg:  cfg,
		data: data,
	}
}

func (c *taskCtrl) Start(ctx context.Context) error {
	l := log.FromContext(ctx).WithName("task")
	ctx = log.IntoContext(ctx, l)

	panic("TODO: implement")
}
