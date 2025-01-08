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

package v2

import (
	"k8s.io/utils/ptr"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	nbmpconv "github.com/nagare-media/engine/internal/pkg/nbmpconv"
	"github.com/nagare-media/engine/pkg/nbmp"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

type workflowToWDDConverter struct {
	wf *enginev1.Workflow
}

var _ nbmpconv.ResetConverter[*enginev1.Workflow, *nbmpv2.Workflow] = &workflowToWDDConverter{}

func NewWorkflowToWDDConverter(wf *enginev1.Workflow) *workflowToWDDConverter {
	c := &workflowToWDDConverter{}
	c.Reset(wf)
	return c
}

func (c *workflowToWDDConverter) Reset(wf *enginev1.Workflow) {
	c.wf = wf
}

func (c *workflowToWDDConverter) Convert(wdd *nbmpv2.Workflow) error {
	// $.scheme.uri
	wdd.Scheme = &nbmpv2.Scheme{URI: nbmpv2.SchemaURI}

	// $.general.id
	wdd.General.ID = c.wf.Name

	if c.wf.Spec.HumanReadable != nil {
		// $.general.name
		if c.wf.Spec.HumanReadable.Name != nil {
			wdd.General.Name = *c.wf.Spec.HumanReadable.Name
		}

		// $.general.description
		if c.wf.Spec.HumanReadable.Description != nil {
			wdd.General.Description = *c.wf.Spec.HumanReadable.Description
		}
	}

	// $.general.nbmp-brand
	wdd.General.NBMPBrand = ptr.To(nbmp.BrandNagareMediaEngineV1)

	// $.general.published-time
	wdd.General.PublishedTime = ptr.To(c.wf.CreationTimestamp.Time)

	// $.general.state
	wdd.General.State = ptr.To(engineWorkflowPhaseToNBMP(c.wf.Status.Phase))
	if !c.wf.DeletionTimestamp.IsZero() {
		wdd.General.State = ptr.To(nbmpv2.DestroyedState)
	}

	// $.processing.start-time
	if !c.wf.Status.StartTime.IsZero() {
		wdd.Processing.StartTime = ptr.To(c.wf.Status.StartTime.Time)
	}

	return nil
}
