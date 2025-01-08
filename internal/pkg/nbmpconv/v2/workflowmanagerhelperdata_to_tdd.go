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
	"fmt"

	"k8s.io/utils/ptr"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	nbmpconv "github.com/nagare-media/engine/internal/pkg/nbmpconv"
	"github.com/nagare-media/engine/pkg/nbmp"
	nbmputils "github.com/nagare-media/engine/pkg/nbmp/utils"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

type workflowManagerHelperDataToTDDConverter struct {
	data *enginev1.WorkflowManagerHelperData

	// config
	mediaConverter    nbmpconv.ResetConverter[*enginev1.Media, *nbmpv2.MediaParameter]
	metadataConverter nbmpconv.ResetConverter[*enginev1.Media, *nbmpv2.MetadataParameter]
}

var _ nbmpconv.ResetConverter[*enginev1.WorkflowManagerHelperData, *nbmpv2.Task] = &workflowManagerHelperDataToTDDConverter{}

func NewWorkflowManagerHelperDataToTDDConverter(data *enginev1.WorkflowManagerHelperData) *workflowManagerHelperDataToTDDConverter {
	c := &workflowManagerHelperDataToTDDConverter{
		mediaConverter:    NewMediaToMediaParameterConverter(nil),
		metadataConverter: NewMediaToMetadataParameterConverter(nil),
	}
	c.Reset(data)
	return c
}

func (c *workflowManagerHelperDataToTDDConverter) Reset(data *enginev1.WorkflowManagerHelperData) {
	c.data = data
}

func (c *workflowManagerHelperDataToTDDConverter) Convert(tdd *nbmpv2.Task) error {
	// $.scheme.uri
	tdd.Scheme = &nbmpv2.Scheme{URI: nbmpv2.SchemaURI}

	// $.general.id: is not mapped (set automatically during task creation)

	if c.data.Task.HumanReadable != nil {
		// $.general.name
		if c.data.Task.HumanReadable.Name != nil {
			tdd.General.Name = *c.data.Task.HumanReadable.Name
		}

		// $.general.description
		if c.data.Task.HumanReadable.Description != nil {
			tdd.General.Description = *c.data.Task.HumanReadable.Description
		}
	}

	// $.general.nbmp-brand
	tdd.General.NBMPBrand = ptr.To(nbmp.BrandNagareMediaEngineV1)

	// $.general.state
	tdd.General.State = ptr.To(nbmpv2.InstantiatedState)

	// $.input
	// $.general.input-ports
	tdd.General.InputPorts = make([]nbmpv2.Port, 0, len(c.data.Task.InputPorts))
	tdd.Input = nbmpv2.Input{}
	for _, ip := range c.data.Task.InputPorts {
		// $.general.input-ports[].bind
		tdd.General.InputPorts = append(tdd.General.InputPorts, nbmpv2.Port{
			PortName: ip.ID,
		})
		p := &tdd.General.InputPorts[len(tdd.General.InputPorts)-1]

		if ip.Input == nil {
			continue
		}

		// $.general.input-ports[].bind.stream-id
		p.Bind = &nbmpv2.PortBinding{
			StreamID: ip.Input.ID,
		}

		// $.general.input-ports[].bind.name
		if ip.Input.HumanReadable != nil {
			p.Bind.Name = ip.Input.HumanReadable.Name
		}

		// $.general.input-ports[].bind.keywords
		for k, v := range ip.Input.Labels {
			p.Bind.Keywords = append(p.Bind.Keywords, encodeKeyword(k, v))
		}

		// $.input.media-parameters
		// $.input.metadata-parameters
		// TODO: filter out duplicate inputs
		if err := c.convertMediaToInputAndOutput(ip.Input, &tdd.Input); err != nil {
			return err
		}
	}

	// $.output
	// $.general.output-ports
	tdd.General.OutputPorts = make([]nbmpv2.Port, 0, len(c.data.Task.OutputPorts))
	tdd.Output = nbmpv2.Output{}
	for _, op := range c.data.Task.OutputPorts {
		// $.general.output-ports[].bind
		tdd.General.OutputPorts = append(tdd.General.OutputPorts, nbmpv2.Port{
			PortName: op.ID,
		})
		p := &tdd.General.OutputPorts[len(tdd.General.OutputPorts)-1]

		if op.Output == nil {
			continue
		}

		// $.general.output-ports[].bind.stream-id
		p.Bind = &nbmpv2.PortBinding{
			StreamID: op.Output.ID,
		}

		// $.general.output-ports[].bind.name
		if op.Output.HumanReadable != nil {
			p.Bind.Name = op.Output.HumanReadable.Name
		}

		// $.general.output-ports[].bind.keywords
		for k, v := range op.Output.Labels {
			p.Bind.Keywords = append(p.Bind.Keywords, encodeKeyword(k, v))
		}

		// $.output.media-parameters
		// $.output.metadata-parameters
		// TODO: filter out duplicate outputs
		if err := c.convertMediaToInputAndOutput(op.Output, &tdd.Output); err != nil {
			return err
		}
	}

	// $.configuration
	tdd.Configuration = &nbmpv2.Configuration{
		Parameters: make([]nbmpv2.Parameter, 0, 6+len(c.data.Task.Config)),
	}

	tdd.Configuration.Parameters = nbmputils.SetStringParameterValue(tdd.Configuration.Parameters, nbmp.EngineWorkflowIDParameterKey, c.data.Workflow.ID)
	if c.data.Workflow.HumanReadable != nil {
		if c.data.Workflow.HumanReadable.Name != nil {
			tdd.Configuration.Parameters = nbmputils.SetStringParameterValue(tdd.Configuration.Parameters, nbmp.EngineWorkflowNameParameterKey, *c.data.Workflow.HumanReadable.Name)
		}
		if c.data.Workflow.HumanReadable.Description != nil {
			tdd.Configuration.Parameters = nbmputils.SetStringParameterValue(tdd.Configuration.Parameters, nbmp.EngineWorkflowDescriptionParameterKey, *c.data.Workflow.HumanReadable.Description)
		}
	}

	tdd.Configuration.Parameters = nbmputils.SetStringParameterValue(tdd.Configuration.Parameters, nbmp.EngineTaskIDParameterKey, c.data.Task.ID)
	if c.data.Task.HumanReadable != nil {
		if c.data.Task.HumanReadable.Name != nil {
			tdd.Configuration.Parameters = nbmputils.SetStringParameterValue(tdd.Configuration.Parameters, nbmp.EngineTaskNameParameterKey, *c.data.Task.HumanReadable.Name)
		}
		if c.data.Task.HumanReadable.Description != nil {
			tdd.Configuration.Parameters = nbmputils.SetStringParameterValue(tdd.Configuration.Parameters, nbmp.EngineTaskDescriptionParameterKey, *c.data.Task.HumanReadable.Description)
		}
	}

	for k, v := range c.data.Task.Config {
		tdd.Configuration.Parameters = nbmputils.SetStringParameterValue(tdd.Configuration.Parameters, k, v)
	}

	// TODO: $.processing
	// TODO: $.requirement
	// TODO: $.step
	// TODO: $.startup-delay
	// TODO: $.client-assistant
	// TODO: $.failover
	// TODO: $.monitoring
	// TODO: $.assertion
	// TODO: $.reporting
	// TODO: $.notification
	// TODO: $.security
	// TODO: $.scale
	// TODO: $.schedule

	return nil
}

func (c *workflowManagerHelperDataToTDDConverter) convertMediaToInputAndOutput(m *enginev1.Media, io nbmpv2.InputOrOutput) error {
	switch m.Type {
	case enginev1.MediaMediaType:
		mp := nbmpv2.MediaParameter{}
		c.mediaConverter.Reset(m)
		if err := c.mediaConverter.Convert(&mp); err != nil {
			return err
		}
		io.SetMediaParameters(append(io.GetMediaParameters(), mp))

	case enginev1.MetadataMediaType:
		mp := nbmpv2.MetadataParameter{}
		c.metadataConverter.Reset(m)
		if err := c.metadataConverter.Convert(&mp); err != nil {
			return err
		}
		io.SetMetadataParameters(append(io.GetMetadataParameters(), mp))

	default:
		return fmt.Errorf("convert: unknown media type '%s'", m.Type)
	}

	return nil
}
