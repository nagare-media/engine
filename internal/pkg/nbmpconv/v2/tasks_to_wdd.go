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
	"errors"
	"fmt"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/internal/pkg/engineurl"
	nbmpconv "github.com/nagare-media/engine/internal/pkg/nbmpconv"
	"github.com/nagare-media/engine/pkg/nbmp"
	nbmputils "github.com/nagare-media/engine/pkg/nbmp/utils"
	"github.com/nagare-media/models.go/base"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

type tasksToWDDConverter struct {
	tsks []enginev1.Task

	streamIDs     map[string]struct{}
	inputMap      map[string]*nbmpv2.Input
	outputMap     map[string]*nbmpv2.Output
	connectionMap map[string]*nbmpv2.ConnectionMapping

	// config
	mediaConverter    nbmpconv.ResetConverter[*enginev1.Media, *nbmpv2.MediaParameter]
	metadataConverter nbmpconv.ResetConverter[*enginev1.Media, *nbmpv2.MetadataParameter]
	gpuResourceName   corev1.ResourceName
}

var _ nbmpconv.ResetConverter[[]enginev1.Task, *nbmpv2.Workflow] = &tasksToWDDConverter{}

func NewTasksToWDDConverter(gpuResourceName corev1.ResourceName, tsks []enginev1.Task) *tasksToWDDConverter {
	c := &tasksToWDDConverter{
		mediaConverter:    NewMediaToMediaParameterConverter(nil),
		metadataConverter: NewMediaToMetadataParameterConverter(nil),
		gpuResourceName:   gpuResourceName,
	}
	c.Reset(tsks)
	return c
}

func (c *tasksToWDDConverter) Reset(tsks []enginev1.Task) {
	c.tsks = tsks
	c.streamIDs = make(map[string]struct{})
	c.inputMap = make(map[string]*nbmpv2.Input)
	c.outputMap = make(map[string]*nbmpv2.Output)
	c.connectionMap = make(map[string]*nbmpv2.ConnectionMapping)
}

func (c *tasksToWDDConverter) Convert(wf *nbmpv2.Workflow) error {
	// $.inputs
	if err := c.convertInputs(wf); err != nil {
		return err
	}

	// $.outputs
	if err := c.convertOutputs(wf); err != nil {
		return err
	}

	// $.processing.connection-map
	if err := c.convertProcessingConnectionMap(wf); err != nil {
		return err
	}

	// $.processing.function-restrictions
	if err := c.convertProcessingFunctionRestrictions(wf); err != nil {
		return err
	}

	// $.processing.keywords: ignore (TODO: what does this specify?)
	// $.processing.image: ignore
	// $.processing.start-time: ignore

	return nil
}

func (c *tasksToWDDConverter) convertInputs(wf *nbmpv2.Workflow) error {
	for _, t := range c.tsks {
		for _, p := range t.Spec.InputPorts {
			if p.Input == nil {
				continue
			}

			if p.Input.URL == nil {
				return errors.New("convert: invalid input URL: nil")
			}

			// check if we have seen this input before

			if _, exists := c.streamIDs[p.Input.ID]; exists {
				continue
			}

			// check if stream is workflow input

			isWfInput := false
			u, err := engineurl.Parse(string(*p.Input.URL))
			if err == nil {
				_, isTaskStream := u.(*engineurl.TaskURL)
				isWfInput = !isTaskStream
			}
			if !isWfInput {
				continue
			}

			// add new input

			mediaParameters, metadataParameters, err := c.convertMediaToInputAndOutput(p.Input)
			if err != nil {
				return err
			}

			if err = c.observeStreamID(p.Input.ID); err != nil {
				return err
			}
			c.inputMap[p.Input.ID] = &nbmpv2.Input{
				MediaParameters:    mediaParameters,
				MetadataParameters: metadataParameters,
			}
		}
	}

	for _, i := range c.inputMap {
		wf.Input.MediaParameters = append(wf.Input.MediaParameters, i.MediaParameters...)
		wf.Input.MetadataParameters = append(wf.Input.MetadataParameters, i.MetadataParameters...)
	}

	return nil
}

func (c *tasksToWDDConverter) convertOutputs(wf *nbmpv2.Workflow) error {
	for _, t := range c.tsks {
		for _, p := range t.Spec.OutputPorts {
			if p.Output == nil {
				continue
			}

			if p.Output.URL == nil {
				return errors.New("convert: invalid output URL: nil")
			}

			// check if we have seen this output before

			if _, exists := c.streamIDs[p.Output.ID]; exists {
				continue
			}

			// check if stream is workflow output

			isWfOutput := false
			u, err := engineurl.Parse(string(*p.Output.URL))
			if err == nil {
				_, isTaskStream := u.(*engineurl.TaskURL)
				isWfOutput = !isTaskStream
			}
			if !isWfOutput {
				continue
			}

			// add new output

			mediaParameters, metadataParameters, err := c.convertMediaToInputAndOutput(p.Output)
			if err != nil {
				return err
			}

			if err = c.observeStreamID(p.Output.ID); err != nil {
				return err
			}
			c.outputMap[p.Output.ID] = &nbmpv2.Output{
				MediaParameters:    mediaParameters,
				MetadataParameters: metadataParameters,
			}
		}
	}

	for _, o := range c.outputMap {
		wf.Output.MediaParameters = append(wf.Input.MediaParameters, o.MediaParameters...)
		wf.Output.MetadataParameters = append(wf.Output.MetadataParameters, o.MetadataParameters...)
	}

	return nil
}

func (c *tasksToWDDConverter) convertProcessingConnectionMap(wf *nbmpv2.Workflow) error {
	for _, t := range c.tsks {
		for _, p := range t.Spec.InputPorts {
			if p.Input == nil {
				continue
			}

			if p.Input.URL == nil {
				return errors.New("convert: invalid input URL: nil")
			}

			inputStream, isWfInput := c.inputMap[p.Input.ID]

			// $.processing.connection-map[].connection-id
			// $.processing.connection-map[].to.id
			// $.processing.connection-map[].to.instance
			// $.processing.connection-map[].to.port-name
			cm := nbmpv2.ConnectionMapping{
				ConnectionID: p.Input.ID,
				To: nbmpv2.ConnectionMappingPort{
					ID:       t.Name,
					Instance: t.Name, // we create a function instance per task even if multiple tasks used the same function instance before
					PortName: p.ID,
				},
			}

			if isWfInput {
				// $.processing.connection-map[].from.id
				// $.processing.connection-map[].from.instance
				// $.processing.connection-map[].from.port-name
				cm.From = nbmpv2.ConnectionMappingPort{
					ID:       p.Input.ID,
					Instance: "",
					PortName: "",
				}
			} else {
				// can we determine the from task?
				if p.Input.Direction == nil || *p.Input.Direction == enginev1.PushMediaDirection {
					// we can't => leave connection mapping to the from task
					continue
				}

				// must be a nagare media engine URL
				u, err := engineurl.Parse(string(*p.Input.URL))
				if err != nil {
					return err
				}

				// must be a task URL
				tu, ok := u.(*engineurl.TaskURL)
				if !ok {
					return fmt.Errorf("convert: unexpected nagare-engine URL type: %T", tu)
				}

				// $.processing.connection-map[].from.id
				// $.processing.connection-map[].from.instance
				// $.processing.connection-map[].from.port-name
				cm.From = nbmpv2.ConnectionMappingPort{
					ID:       tu.TaskID,
					Instance: tu.TaskID,
					PortName: tu.PortName,
				}

				mediaParameters, metadataParameters, err := c.convertMediaToInputAndOutput(p.Input)
				if err != nil {
					return err
				}
				inputStream = &nbmpv2.Input{
					MediaParameters:    mediaParameters,
					MetadataParameters: metadataParameters,
				}
			}

			// $.processing.connection-map[].to.input-restrictions
			cm.To.InputRestrictions = inputStream
			// $.processing.connection-map[].from.output-restrictions
			cm.From.OutputRestrictions = &nbmpv2.Output{
				MediaParameters:    inputStream.MediaParameters,
				MetadataParameters: inputStream.MetadataParameters,
			}

			// $.processing.connection-map[].flowcontrol
			// $.processing.connection-map[].other-parameters
			fcr, op, err := c.convertProcessingConnectionMapFlowcontrolAndOtherParameters(p.Input.URL)
			if err != nil {
				return err
			}
			cm.Flowcontrol = fcr
			cm.OtherParameters = op

			// TODO: $.processing.connection-map[].co-located
			// TODO: $.processing.connection-map[].breakable

			key := c.connectionMappingKey(&cm)
			if _, exists := c.connectionMap[key]; !exists {
				c.connectionMap[key] = &cm
			}
		}

		for _, p := range t.Spec.OutputPorts {
			if p.Output == nil {
				continue
			}

			if p.Output.URL == nil {
				return errors.New("convert: invalid output URL: nil")
			}

			outputStream, isWfOutput := c.outputMap[p.Output.ID]

			// $.processing.connection-map[].connection-id
			// $.processing.connection-map[].from.id
			// $.processing.connection-map[].from.instance
			// $.processing.connection-map[].from.port-name
			cm := nbmpv2.ConnectionMapping{
				ConnectionID: p.Output.ID,
				From: nbmpv2.ConnectionMappingPort{
					ID:       t.Name,
					Instance: t.Name, // we create a function instance per task even if multiple tasks used the same function instance before
					PortName: p.ID,
				},
			}

			if isWfOutput {
				// $.processing.connection-map[].from.id
				// $.processing.connection-map[].from.instance
				// $.processing.connection-map[].from.port-name
				cm.To = nbmpv2.ConnectionMappingPort{
					ID:       p.Output.ID,
					Instance: "",
					PortName: "",
				}
			} else {
				// can we determine the to task?
				if p.Output.Direction == nil || *p.Output.Direction == enginev1.PullMediaDirection {
					// we can't => leave connection mapping to the to task
					continue
				}

				// must be a nagare media engine URL
				u, err := engineurl.Parse(string(*p.Output.URL))
				if err != nil {
					return err
				}

				// must be a task URL
				tu, ok := u.(*engineurl.TaskURL)
				if !ok {
					return fmt.Errorf("convert: unexpected nagare-engine URL type: %T", tu)
				}

				// $.processing.connection-map[].to.id
				// $.processing.connection-map[].to.instance
				// $.processing.connection-map[].to.port-name
				cm.To = nbmpv2.ConnectionMappingPort{
					ID:       tu.TaskID,
					Instance: tu.TaskID,
					PortName: tu.PortName,
				}

				mediaParameters, metadataParameters, err := c.convertMediaToInputAndOutput(p.Output)
				if err != nil {
					return err
				}
				outputStream = &nbmpv2.Output{
					MediaParameters:    mediaParameters,
					MetadataParameters: metadataParameters,
				}
			}

			// $.processing.connection-map[].to.input-restrictions
			cm.To.InputRestrictions = &nbmpv2.Input{
				MediaParameters:    outputStream.MediaParameters,
				MetadataParameters: outputStream.MetadataParameters,
			}
			// $.processing.connection-map[].from.output-restrictions
			cm.From.OutputRestrictions = outputStream

			// $.processing.connection-map[].flowcontrol
			// $.processing.connection-map[].other-parameters
			fcr, op, err := c.convertProcessingConnectionMapFlowcontrolAndOtherParameters(p.Output.URL)
			if err != nil {
				return err
			}
			cm.Flowcontrol = fcr
			cm.OtherParameters = op

			// TODO: $.processing.connection-map[].co-located
			// TODO: $.processing.connection-map[].breakable

			key := c.connectionMappingKey(&cm)
			if _, exists := c.connectionMap[key]; !exists {
				c.connectionMap[key] = &cm
			}
		}
	}

	// $.processing.connection-map
	wf.Processing.ConnectionMap = make([]nbmpv2.ConnectionMapping, 0, len(c.connectionMap))
	for _, cm := range c.connectionMap {
		wf.Processing.ConnectionMap = append(wf.Processing.ConnectionMap, *cm)
	}

	return nil
}

func (c *tasksToWDDConverter) convertProcessingFunctionRestrictions(wf *nbmpv2.Workflow) error {
	// we create a function instance per task even if multiple tasks used the same function instance before
	wf.Processing.FunctionRestrictions = make([]nbmpv2.FunctionRestriction, 0, len(c.tsks))
	for _, t := range c.tsks {
		// $.general.id
		if wf.General.ID == "" {
			wf.General.ID = t.Spec.WorkflowRef.Name
		}
		if wf.General.ID != t.Spec.WorkflowRef.Name {
			return fmt.Errorf("convert: unexpected workflow ID in task: '%s' != %s", wf.General.ID, t.Spec.WorkflowRef.Name)
		}

		// $.processing.function-restrictions.instance
		// $.processing.function-restrictions.general.state
		fr := nbmpv2.FunctionRestriction{
			Instance: t.Name,
			General:  &nbmpv2.General{},
		}

		// $.processing.function-restrictions.general

		// $.processing.function-restrictions.general.state
		fr.General.State = ptr.To(engineTaskPhaseToNBMP(t.Status.Phase))
		if t.DeletionTimestamp != nil {
			fr.General.State = ptr.To(nbmpv2.DestroyedState)
		}

		// $.processing.function-restrictions.general.id
		if t.Spec.FunctionRef != nil {
			fr.General = &nbmpv2.General{
				ID: fmt.Sprintf("%s/%s", t.Spec.FunctionRef.Kind, t.Spec.FunctionRef.Name),
			}
		}

		// TODO: $.processing.function-restrictions.general.*

		// $.processing.function-restrictions.processing.keywords
		if t.Spec.FunctionSelector != nil {
			fr.Processing = &nbmpv2.Processing{}
			for _, exp := range t.Spec.FunctionSelector.MatchExpressions {
				switch exp.Operator {
				case metav1.LabelSelectorOpExists:
					fr.Processing.Keywords = append(fr.Processing.Keywords, exp.Key)
				case metav1.LabelSelectorOpIn:
					if len(exp.Values) != 1 {
						return errors.New("convert: unsupported multi-value expression matching")
					}
					fr.Processing.Keywords = append(fr.Processing.Keywords, encodeKeyword(exp.Key, exp.Values[0]))
				case metav1.LabelSelectorOpNotIn, metav1.LabelSelectorOpDoesNotExist:
					fallthrough
				default:
					return fmt.Errorf("convert: expression matching operation not implemented: '%s'", exp.Operator)
				}
			}

			for k, v := range t.Spec.FunctionSelector.MatchLabels {
				fr.Processing.Keywords = append(fr.Processing.Keywords, encodeKeyword(k, v))
			}
		}

		// $.processing.function-restrictions.processing.start-time
		if t.Status.StartTime != nil {
			if fr.Processing == nil {
				fr.Processing = &nbmpv2.Processing{}
			}
			fr.Processing.StartTime = ptr.To(t.Status.StartTime.Time)
		}

		// TODO: $.processing.function-restrictions.processing.*

		// $.processing.function-restrictions.requirements

		// $.processing.function-restrictions.requirements.hardware.placement
		if t.Spec.MediaProcessingEntitySelector != nil &&
			t.Spec.MediaProcessingEntitySelector.MatchLabels != nil &&
			t.Spec.MediaProcessingEntitySelector.MatchLabels[enginev1.MediaProcessingEntityLocationLabel] != "" {
			placement := t.Spec.MediaProcessingEntitySelector.MatchLabels[enginev1.MediaProcessingEntityLocationLabel]
			fr.Requirements = &nbmpv2.Requirement{
				Hardware: &nbmpv2.HardwareRequirement{
					Placement: ptr.To(nbmpv2.HardwareRequirementPlacement(placement)),
				},
			}
		}

		// $.processing.function-restrictions.requirements.hardware
		if t.Spec.TemplatePatches != nil && len(t.Spec.TemplatePatches.Spec.Template.Spec.Containers) > 0 {
			// TODO: fix assumption that first container is main workload
			// TODO: is this approach ok for requests and limits?
			rl := t.Spec.TemplatePatches.Spec.Template.Spec.Containers[0].Resources.Requests
			if len(rl) == 0 {
				rl = t.Spec.TemplatePatches.Spec.Template.Spec.Containers[0].Resources.Limits
			}
			if len(rl) != 0 {
				if fr.Requirements == nil {
					fr.Requirements = &nbmpv2.Requirement{}
				}
				if fr.Requirements.Hardware == nil {
					fr.Requirements.Hardware = &nbmpv2.HardwareRequirement{}
				}
			}

			for res, quantity := range rl {
				switch res {
				case v1.ResourceCPU:
					// $.processing.function-restrictions.requirements.hardware.vcpu
					fr.Requirements.Hardware.VCPU = ptr.To(uint64(quantity.Value()))
				case c.gpuResourceName: // TODO: make gpuResourceName configurable per Task
					// $.processing.function-restrictions.requirements.hardware.vgpu
					fr.Requirements.Hardware.VGPU = ptr.To(uint64(quantity.Value()))
				case v1.ResourceMemory:
					// $.processing.function-restrictions.requirements.hardware.ram
					fr.Requirements.Hardware.RAM = ptr.To(uint64(quantity.ScaledValue(resource.Mega)))
				case v1.ResourceEphemeralStorage:
					// $.processing.function-restrictions.requirements.hardware.disk
					fr.Requirements.Hardware.Disk = ptr.To(uint64(quantity.ScaledValue(resource.Giga)))
				default:
					// NBMP doesn't support other hardware resource constraints
				}
			}
		}

		// TODO: $.processing.function-restrictions.requirements.flowcontrol
		// TODO: $.processing.function-restrictions.requirements.security
		// TODO: $.processing.function-restrictions.requirements.workflow-task
		// TODO: $.processing.function-restrictions.requirements.resource-estimators

		// $.processing.function-restrictions.configuration
		if len(t.Spec.Config) > 0 {
			fr.Configuration = &nbmpv2.Configuration{
				Parameters: make([]nbmpv2.Parameter, 0, len(t.Spec.Config)),
			}

			for k, v := range t.Spec.Config {
				fr.Configuration.Parameters = nbmputils.SetStringParameterValue(fr.Configuration.Parameters, k, v)
			}
		}

		// TODO: $.processing.function-restrictions.client-assistant
		// TODO: $.processing.function-restrictions.failover
		// TODO: $.processing.function-restrictions.monitoring
		// TODO: $.processing.function-restrictions.reporting
		// TODO: $.processing.function-restrictions.notification
		// TODO: $.processing.function-restrictions.step
		// TODO: $.processing.function-restrictions.security
		// TODO: $.processing.function-restrictions.blacklist

		// $.processing.function-restrictions.processing
		wf.Processing.FunctionRestrictions = append(wf.Processing.FunctionRestrictions, fr)
	}

	return nil
}

func (c *tasksToWDDConverter) convertMediaToInputAndOutput(m *enginev1.Media) ([]nbmpv2.MediaParameter, []nbmpv2.MetadataParameter, error) {
	var (
		err                error
		mediaParameters    []nbmpv2.MediaParameter
		metadataParameters []nbmpv2.MetadataParameter
	)

	switch m.Type {
	case enginev1.MediaMediaType:
		mp := nbmpv2.MediaParameter{}
		c.mediaConverter.Reset(m)
		if err := c.mediaConverter.Convert(&mp); err != nil {
			return nil, nil, err
		}
		// remove query parameters
		url, err := c.removeQueryParameters(&mp.CachingServerURL)
		if err != nil {
			return nil, nil, err
		}
		if url != nil {
			mp.CachingServerURL = *url
		}
		mediaParameters = append(mediaParameters, mp)

	case enginev1.MetadataMediaType:
		mp := nbmpv2.MetadataParameter{}
		c.metadataConverter.Reset(m)
		if err := c.metadataConverter.Convert(&mp); err != nil {
			return nil, nil, err
		}
		// remove query parameters
		mp.CachingServerURL, err = c.removeQueryParameters(mp.CachingServerURL)
		if err != nil {
			return nil, nil, err
		}
		metadataParameters = append(metadataParameters, mp)

	default:
		return nil, nil, fmt.Errorf("convert: unknown media type '%s'", m.Type)
	}

	return mediaParameters, metadataParameters, nil
}

func (c *tasksToWDDConverter) removeQueryParameters(u *base.URI) (*base.URI, error) {
	if u == nil {
		return nil, nil
	}

	url, err := u.URL()
	if err != nil {
		return u, err
	}

	url.RawQuery = ""

	return ptr.To(base.URI(url.String())), nil
}

func (c *tasksToWDDConverter) convertProcessingConnectionMapFlowcontrolAndOtherParameters(u *base.URI) (*nbmpv2.FlowcontrolRequirement, []nbmpv2.Parameter, error) {
	if u == nil {
		return nil, nil, nil
	}

	url, err := u.URL()
	if err != nil {
		return nil, nil, err
	}

	q := url.Query()
	fcr := nbmpv2.FlowcontrolRequirement{}
	op := make([]nbmpv2.Parameter, 0)
	for k := range q {
		v := q.Get(k)
		if v == "" {
			continue
		}

		// $.processing.connection-map[].flowcontrol.typical-delay
		// $.processing.connection-map[].flowcontrol.min-delay
		// $.processing.connection-map[].flowcontrol.max-delay
		// $.processing.connection-map[].flowcontrol.min-throughput
		// $.processing.connection-map[].flowcontrol.max-throughput
		// $.processing.connection-map[].flowcontrol.averaging-window
		// $.processing.connection-map[].other-parameters
		switch k {
		case nbmp.TypicalDelayFlowcontrolQueryParameterKey:
			i, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return nil, nil, err
			}
			fcr.TypicalDelay = &i

		case nbmp.MinDelayFlowcontrolQueryParameterKey:
			i, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return nil, nil, err
			}
			fcr.MinDelay = &i

		case nbmp.MaxDelayFlowcontrolQueryParameterKey:
			i, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return nil, nil, err
			}
			fcr.MaxDelay = &i

		case nbmp.MinThroughputFlowcontrolQueryParameterKey:
			i, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return nil, nil, err
			}
			fcr.MinThroughput = &i

		case nbmp.MaxThroughputFlowcontrolQueryParameterKey:
			i, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return nil, nil, err
			}
			fcr.MaxThroughput = &i

		case nbmp.AveragingWindowFlowcontrolQueryParameterKey:
			i, err := strconv.ParseUint(v, 10, 64)
			if err != nil {
				return nil, nil, err
			}
			fcr.AveragingWindow = &i

		default:
			op = nbmputils.SetStringParameterValue(op, k, v)
		}
	}

	return &fcr, op, nil
}

func (c *tasksToWDDConverter) connectionMappingKey(cm *nbmpv2.ConnectionMapping) string {
	if cm == nil {
		return ""
	}

	buf := strings.Builder{}
	buf.Grow(
		3 +
			len(cm.From.ID) +
			len(cm.From.PortName) +
			len(cm.To.ID) +
			len(cm.To.PortName),
	)

	buf.WriteString(cm.From.ID)
	buf.WriteString(cm.From.PortName)
	buf.WriteString(cm.To.ID)
	buf.WriteString(cm.To.PortName)

	return buf.String()
}

func (c *tasksToWDDConverter) observeStreamID(streamID string) error {
	if _, exists := c.streamIDs[streamID]; exists {
		return fmt.Errorf("convert: duplicate stream-id '%s'", streamID)
	}
	c.streamIDs[streamID] = struct{}{}
	return nil
}
