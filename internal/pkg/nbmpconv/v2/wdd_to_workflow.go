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

package v2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	nbmpconv "github.com/nagare-media/engine/internal/pkg/nbmpconv"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

type wddToWorkflowConverter struct {
	wdd *nbmpv2.Workflow
}

var _ nbmpconv.ResetConverter[*nbmpv2.Workflow, *enginev1.Workflow] = &wddToWorkflowConverter{}

func NewWDDToWorkflowConverter(wdd *nbmpv2.Workflow) *wddToWorkflowConverter {
	c := &wddToWorkflowConverter{}
	c.Reset(wdd)
	return c
}

func (c *wddToWorkflowConverter) Reset(wdd *nbmpv2.Workflow) {
	c.wdd = wdd
}

func (c *wddToWorkflowConverter) Convert(wf *enginev1.Workflow) error {
	// $.scheme.uri: ignore

	// $.general.id
	wf.ObjectMeta = metav1.ObjectMeta{
		Name: c.wdd.General.ID,
	}

	// $.general.name
	// $.general.description
	wf.Spec = enginev1.WorkflowSpec{
		HumanReadable: &enginev1.HumanReadableWorkflowDescription{
			Name:        &c.wdd.General.Name,
			Description: &c.wdd.General.Description,
		},
	}

	// $.general.nbmp-brand: ignore
	// $.general.published-time: ignore
	// $.general.state: ignore
	// $.processing.start-time: ignore

	return nil
}
