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
	"net/url"
	"strconv"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/internal/pkg/apis/utils"
	"github.com/nagare-media/engine/internal/pkg/engineurl"
	nbmpconv "github.com/nagare-media/engine/internal/pkg/nbmpconv"
	"github.com/nagare-media/engine/pkg/apis/meta"
	"github.com/nagare-media/engine/pkg/nbmp"
	nbmputils "github.com/nagare-media/engine/pkg/nbmp/utils"
	"github.com/nagare-media/models.go/base"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

type wddToTasksConverter struct {
	wdd *nbmpv2.Workflow

	streamIDs   map[string]struct{}
	inputMap    map[string]*enginev1.Media
	outputMap   map[string]*enginev1.Media
	instanceMap map[string]*enginev1.Task
	tskMap      map[string]*enginev1.Task

	// config
	mediaConverter    nbmpconv.ResetConverter[*nbmpv2.MediaParameter, *enginev1.Media]
	metadataConverter nbmpconv.ResetConverter[*nbmpv2.MetadataParameter, *enginev1.Media]
	schema            *runtime.Scheme
	gpuResourceName   corev1.ResourceName
}

var _ nbmpconv.ResetConverter[*nbmpv2.Workflow, *[]enginev1.Task] = &wddToTasksConverter{}

func NewWDDToTasksConverter(schema *runtime.Scheme, gpuResourceName corev1.ResourceName, wdd *nbmpv2.Workflow) *wddToTasksConverter {
	c := &wddToTasksConverter{
		mediaConverter:    NewMediaParameterToMediaConverter(nil),
		metadataConverter: NewMetadataParameterToMediaConverter(nil),
		schema:            schema,
		gpuResourceName:   gpuResourceName,
	}
	c.Reset(wdd)
	return c
}

func (c *wddToTasksConverter) Reset(wdd *nbmpv2.Workflow) {
	c.wdd = wdd
	c.streamIDs = make(map[string]struct{})
	c.inputMap = make(map[string]*enginev1.Media)
	c.outputMap = make(map[string]*enginev1.Media)
	c.instanceMap = make(map[string]*enginev1.Task)
	c.tskMap = make(map[string]*enginev1.Task)
}

func (c *wddToTasksConverter) Convert(tsks *[]enginev1.Task) error {
	// $.inputs
	if err := c.convertInputs(); err != nil {
		return err
	}

	// $.outputs
	if err := c.convertOutputs(); err != nil {
		return err
	}

	// $.processing.function-restrictions
	if err := c.convertProcessingFunctionRestrictions(); err != nil {
		return err
	}

	// $.processing.connection-map
	if err := c.convertProcessingConnectionMap(); err != nil {
		return err
	}

	// $.processing.keywords: ignore (TODO: what does this specify?)
	// $.processing.image: ignore
	// $.processing.start-time: ignore

	// gather all tasks
	*tsks = make([]enginev1.Task, 0, len(c.tskMap))
	for _, v := range c.tskMap {
		*tsks = append(*tsks, *v)
	}

	return nil
}

func (c *wddToTasksConverter) convertInputs() error {
	for _, mp := range c.wdd.Input.MediaParameters {
		m, err := c.convertMediaParametersToMedia(&mp)
		if err != nil {
			return err
		}
		c.inputMap[m.ID] = m
	}

	for _, mp := range c.wdd.Input.MetadataParameters {
		m, err := c.convertMetadataParametersToMedia(&mp)
		if err != nil {
			return err
		}
		c.inputMap[m.ID] = m
	}

	return nil
}

func (c *wddToTasksConverter) convertOutputs() error {
	for _, mp := range c.wdd.Output.MediaParameters {
		m, err := c.convertMediaParametersToMedia(&mp)
		if err != nil {
			return err
		}
		c.outputMap[m.ID] = m
	}

	for _, mp := range c.wdd.Output.MetadataParameters {
		m, err := c.convertMetadataParametersToMedia(&mp)
		if err != nil {
			return err
		}
		c.outputMap[m.ID] = m
	}

	return nil
}

func (c *wddToTasksConverter) convertProcessingFunctionRestrictions() error {
	for _, fr := range c.wdd.Processing.FunctionRestrictions {
		if _, exists := c.instanceMap[fr.Instance]; exists {
			return fmt.Errorf("convert: duplicate function instance '%s'", fr.Instance)
		}

		tsk := &enginev1.Task{}

		// set workflow reference
		tsk.Labels = map[string]string{
			enginev1.WorkflowNameLabel: c.wdd.General.ID,
		}
		wfRef := &meta.LocalObjectReference{Name: c.wdd.General.ID}
		if err := utils.NormalizeLocalRef(c.schema, wfRef, &enginev1.Workflow{}); err != nil {
			return err
		}
		tsk.Spec.WorkflowRef = *wfRef

		// set media processing entity selector
		// $.processing.function-restrictions.requirements.hardware.placement
		mpeSel := &metav1.LabelSelector{MatchLabels: make(map[string]string)}
		if fr.Requirements != nil && fr.Requirements.Hardware != nil && fr.Requirements.Hardware.Placement != nil {
			mpeSel.MatchLabels[enginev1.MediaProcessingEntityLocationLabel] = string(*fr.Requirements.Hardware.Placement)
		} else {
			mpeSel.MatchLabels[enginev1.IsDefaultMediaProcessingEntityAnnotation] = "true"
		}
		tsk.Spec.MediaProcessingEntitySelector = mpeSel

		// $.processing.function-restrictions.general
		if fr.General != nil {
			// $.processing.function-restrictions.general.id
			if fr.General.ID != "" {
				// we expect the format {kind}/{id}
				i := strings.IndexRune(fr.General.ID, '/')
				if i <= 0 || i == len(fr.General.ID)-1 {
					return fmt.Errorf("convert: unexpected function requirement: unexpected general.id format: '%s'", fr.General.ID)
				}
				ref := &meta.LocalObjectReference{
					APIVersion: enginev1.GroupVersion.Identifier(),
					Kind:       fr.General.ID[:i],
					Name:       fr.General.ID[i+1 : len(fr.General.ID)],
				}
				if err := utils.NormalizeLocalFunctionRef(c.schema, ref); err != nil {
					return err
				}
				tsk.Spec.FunctionRef = ref
			}
			// TODO: $.processing.function-restrictions.general.*
		}

		// $.processing.function-restrictions.processing
		if fr.Processing != nil {
			// $.processing.function-restrictions.processing.keywords
			tsk.Spec.FunctionSelector = &metav1.LabelSelector{}
			for _, s := range fr.Processing.Keywords {
				k, v := parseKeyword(s)
				var sel metav1.LabelSelectorRequirement
				if v == "" {
					sel = metav1.LabelSelectorRequirement{
						Key:      k,
						Operator: metav1.LabelSelectorOpExists,
					}
				} else {
					sel = metav1.LabelSelectorRequirement{
						Key:      k,
						Operator: metav1.LabelSelectorOpIn,
						Values:   []string{v},
					}
				}
				tsk.Spec.FunctionSelector.MatchExpressions = append(tsk.Spec.FunctionSelector.MatchExpressions, sel)
			}
			// TODO: $.processing.function-restrictions.processing.*
		}

		// $.processing.function-restrictions.requirements
		if fr.Requirements != nil {
			// TODO: $.processing.function-restrictions.requirements.flowcontrol
			rl := corev1.ResourceList{}

			// $.processing.function-restrictions.requirements.hardware
			if fr.Requirements.Hardware != nil {
				// $.processing.function-restrictions.requirements.hardware.vcpu
				if fr.Requirements.Hardware.VCPU != nil {
					q := resource.Quantity{Format: resource.DecimalSI}
					q.Set(int64(*fr.Requirements.Hardware.VCPU))
					rl[corev1.ResourceCPU] = q
				}

				// $.processing.function-restrictions.requirements.hardware.vgpu
				if fr.Requirements.Hardware.VGPU != nil {
					q := resource.Quantity{Format: resource.DecimalSI}
					q.Set(int64(*fr.Requirements.Hardware.VGPU))
					// TODO: make gpuResourceName configurable per Task
					rl[c.gpuResourceName] = q
				}

				// $.processing.function-restrictions.requirements.hardware.ram
				if fr.Requirements.Hardware.RAM != nil {
					q := resource.Quantity{Format: resource.BinarySI}
					q.SetScaled(int64(*fr.Requirements.Hardware.RAM), resource.Mega)
					rl[corev1.ResourceMemory] = q
				}

				// $.processing.function-restrictions.requirements.hardware.disk
				if fr.Requirements.Hardware.Disk != nil {
					q := resource.Quantity{Format: resource.BinarySI}
					q.SetScaled(int64(*fr.Requirements.Hardware.Disk), resource.Giga)
					rl[corev1.ResourceEphemeralStorage] = q
				}

				tsk.Spec.TemplatePatches = &batchv1.JobTemplateSpec{
					Spec: batchv1.JobSpec{
						Template: corev1.PodTemplateSpec{
							Spec: corev1.PodSpec{
								Containers: []corev1.Container{{
									Resources: corev1.ResourceRequirements{
										Requests: rl,
										Limits:   rl,
									},
								}},
							},
						},
					},
				}
			}

			// TODO: $.processing.function-restrictions.requirements.security
			// TODO: $.processing.function-restrictions.requirements.workflow-task
			// TODO: $.processing.function-restrictions.requirements.resource-estimators
		}

		// $.processing.function-restrictions.configuration
		if fr.Configuration != nil {
			tsk.Spec.Config = make(map[string]string, len(fr.Configuration.Parameters))
			for _, p := range fr.Configuration.Parameters {
				v, ok := nbmputils.ExtractStringParameterValue(p)
				if !ok {
					return fmt.Errorf("convert: unsupported parameter value for '%s'", p.Name)
				}
				tsk.Spec.Config[p.Name] = v
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

		// $.processing.function-restrictions.instance
		c.instanceMap[fr.Instance] = tsk
	}

	return nil
}

func (c *wddToTasksConverter) convertProcessingConnectionMap() error {
	for _, cm := range c.wdd.Processing.ConnectionMap {
		var (
			fromTsk, toTsk *enginev1.Task
			ok             bool
		)

		// $.processing.connection-map[].from.id
		// check if from is an input
		_, fromIsWfInput := c.inputMap[cm.From.ID]
		if !fromIsWfInput {
			// check if task is already known
			fromTsk, ok = c.tskMap[cm.From.ID]
			if !ok {
				// $.processing.connection-map[].from.instance
				// create task based on function instance
				instance, ok := c.instanceMap[cm.From.Instance]
				if !ok {
					return fmt.Errorf("convert: unknown function instance '%s'", cm.From.Instance)
				}
				fromTsk = instance.DeepCopy()
			}
		}

		// $.processing.connection-map[].to.id
		// check if to is an output
		_, toIsWfOutput := c.outputMap[cm.To.ID]
		if !toIsWfOutput {
			// check if task is already known
			toTsk, ok = c.tskMap[cm.To.ID]
			if !ok {
				// $.processing.connection-map[].to.instance
				// create task based on function instance
				instance, ok := c.instanceMap[cm.To.Instance]
				if !ok {
					return fmt.Errorf("convert: unknown function instance '%s'", cm.To.Instance)
				}
				toTsk = instance.DeepCopy()
			}
		}

		if fromIsWfInput && toIsWfOutput {
			return errors.New("convert: cannot directly connect workflow input to output")
		}

		if cm.From.ID == cm.To.ID {
			// TODO: this should be checked beforehand in the validator
			return errors.New("convert: cannot connect task to itself")
		}

		// initialize stream
		stream, err := c.createMediaStream(cm)
		if err != nil {
			return err
		}

		// attach output port to output stream
		if !fromIsWfInput {
			// $.processing.connection-map[].from.port-name
			var p *enginev1.OutputPortBindings
			for i := range fromTsk.Spec.OutputPorts {
				if fromTsk.Spec.OutputPorts[i].ID == cm.From.PortName {
					p = &fromTsk.Spec.OutputPorts[i]
					break
				}
			}

			if p == nil {
				fromTsk.Spec.OutputPorts = append(fromTsk.Spec.OutputPorts, enginev1.OutputPortBindings{ID: cm.From.PortName})
				p = &fromTsk.Spec.OutputPorts[len(fromTsk.Spec.OutputPorts)-1]
			}

			if p.Output != nil {
				return fmt.Errorf("convert: more than one output attached to port '%s' of task '%s'", cm.From.PortName, cm.From.ID)
			}
			p.Output = stream.DeepCopy()

			// add from task to tskMap
			c.tskMap[cm.From.ID] = fromTsk
		}

		// attach input port to input stream
		if !toIsWfOutput {
			// $.processing.connection-map[].to.port-name
			var p *enginev1.InputPortBinding
			for i := range toTsk.Spec.InputPorts {
				if toTsk.Spec.InputPorts[i].ID == cm.To.PortName {
					p = &toTsk.Spec.InputPorts[i]
					break
				}
			}

			if p == nil {
				toTsk.Spec.InputPorts = append(toTsk.Spec.InputPorts, enginev1.InputPortBinding{ID: cm.To.PortName})
				p = &toTsk.Spec.InputPorts[len(toTsk.Spec.InputPorts)-1]
			}

			if p.Input != nil {
				return fmt.Errorf("convert: more than one input attached to port '%s' of task '%s'", cm.To.PortName, cm.To.ID)
			}
			p.Input = stream.DeepCopy()

			// add to task to tskMap
			c.tskMap[cm.To.ID] = toTsk
		}
	}

	// TODO: $.processing.connection-map[].co-located
	// TODO: $.processing.connection-map[].breakable

	return nil
}

func (c *wddToTasksConverter) createMediaStream(cm nbmpv2.ConnectionMapping) (*enginev1.Media, error) {
	var (
		stream = &enginev1.Media{}
		err    error
	)

	// check for input or output stream
	inputStream, fromIsWfInput := c.inputMap[cm.From.ID]
	outputStream, toIsWfOutput := c.outputMap[cm.To.ID]
	if fromIsWfInput {
		stream = inputStream.DeepCopy()
	} else if toIsWfOutput {
		stream = outputStream.DeepCopy()
	}

	// check for from restrictions

	// $.processing.connection-map[].from.output-restrictions
	var fromRestrictions *enginev1.Media
	if cm.From.OutputRestrictions != nil {
		fromRestrictions, err = c.convertMediaOrMetadataRestrictionToMedia(cm.From.OutputRestrictions.MediaParameters, cm.From.OutputRestrictions.MetadataParameters)
		if err != nil {
			return nil, err
		}
	}

	// check for to restrictions

	// $.processing.connection-map[].to.input-restrictions
	var toRestrictions *enginev1.Media
	if cm.To.InputRestrictions != nil {
		toRestrictions, err = c.convertMediaOrMetadataRestrictionToMedia(cm.To.InputRestrictions.MediaParameters, cm.To.InputRestrictions.MetadataParameters)
		if err != nil {
			return nil, err
		}
	}

	// three way merge of stream, fromRestrictions and toRestrictions

	if err := c.mergeMediaRestrictions(stream, fromRestrictions); err != nil {
		return nil, err
	}

	if err := c.mergeMediaRestrictions(stream, toRestrictions); err != nil {
		return nil, err
	}

	// set defaults and validate values

	// Type
	if stream.Type == "" {
		stream.Type = enginev1.MediaMediaType
	}

	// ID
	if stream.ID == "" {
		// $.processing.connection-map[].connection-id
		stream.ID = cm.ConnectionID
	}

	// HumanReadable.Name
	if stream.HumanReadable == nil || stream.HumanReadable.Name == nil {
		// TODO: should we use the connection ID instead?
		stream.HumanReadable = &enginev1.HumanReadableMediaDescription{
			Name: ptr.To(fmt.Sprintf("%s --> %s", cm.From.ID, cm.To.ID)),
		}
	}

	// Direction
	if stream.Direction == nil {
		if toIsWfOutput {
			stream.Direction = ptr.To(enginev1.PullMediaDirection)
		} else {
			stream.Direction = ptr.To(enginev1.PushMediaDirection)
		}
	}

	// URL
	var streamURL *url.URL
	if stream.URL != nil {
		streamURL, err = url.Parse(string(*stream.URL))
		if err != nil {
			return nil, fmt.Errorf("convert: failed to parse stream URL: %w", err)
		}
	} else {
		if (fromIsWfInput && *stream.Direction == enginev1.PullMediaDirection) ||
			(toIsWfOutput && *stream.Direction == enginev1.PushMediaDirection) {
			return nil, errors.New("convert: unable to determine stream URL")
		}

		var originPort nbmpv2.ConnectionMappingPort
		if *stream.Direction == enginev1.PushMediaDirection {
			originPort = cm.To
		} else {
			originPort = cm.From
		}

		tu := engineurl.TaskURL{
			WorkflowID: c.wdd.General.ID,
			TaskID:     originPort.ID,
			PortName:   originPort.PortName,
			StreamID:   stream.ID,
		}
		streamURL = tu.URL()
	}

	query := streamURL.Query()

	// $.processing.connection-map[].flowcontrol
	if cm.Flowcontrol != nil {
		// $.processing.connection-map[].flowcontrol.typical-delay
		if cm.Flowcontrol.TypicalDelay != nil {
			query.Add(nbmp.TypicalDelayFlowcontrolQueryParameterKey, strconv.FormatUint(*cm.Flowcontrol.TypicalDelay, 10))
		}

		// $.processing.connection-map[].flowcontrol.min-delay
		if cm.Flowcontrol.MinDelay != nil {
			query.Add(nbmp.MinDelayFlowcontrolQueryParameterKey, strconv.FormatUint(*cm.Flowcontrol.MinDelay, 10))
		}

		// $.processing.connection-map[].flowcontrol.max-delay
		if cm.Flowcontrol.MaxDelay != nil {
			query.Add(nbmp.MaxDelayFlowcontrolQueryParameterKey, strconv.FormatUint(*cm.Flowcontrol.MaxDelay, 10))
		}

		// $.processing.connection-map[].flowcontrol.min-throughput
		if cm.Flowcontrol.MinThroughput != nil {
			query.Add(nbmp.MinThroughputFlowcontrolQueryParameterKey, strconv.FormatUint(*cm.Flowcontrol.MinThroughput, 10))
		}

		// $.processing.connection-map[].flowcontrol.max-throughput
		if cm.Flowcontrol.MaxThroughput != nil {
			query.Add(nbmp.MaxThroughputFlowcontrolQueryParameterKey, strconv.FormatUint(*cm.Flowcontrol.MaxThroughput, 10))
		}

		// $.processing.connection-map[].flowcontrol.averaging-window
		if cm.Flowcontrol.AveragingWindow != nil {
			query.Add(nbmp.AveragingWindowFlowcontrolQueryParameterKey, strconv.FormatUint(*cm.Flowcontrol.AveragingWindow, 10))
		}
	}

	// $.processing.connection-map[].other-parameters
	for _, p := range cm.OtherParameters {
		v, ok := nbmputils.ExtractStringParameterValue(p)
		if !ok {
			return nil, fmt.Errorf("convert: unsupported parameter value for '%s'", p.Name)
		}
		query.Add(p.Name, v)
	}
	streamURL.RawQuery = query.Encode()

	stream.URL = ptr.To(base.URI(streamURL.String()))

	// TODO: check if URL is valid (e.g. correct direction)

	// TODO: should we set a default Metadata.MimeType?

	// TODO: should we set a default Metadata.CodecType?

	return stream, nil
}

func (c *wddToTasksConverter) mergeMediaRestrictions(stream, restrictions *enginev1.Media) error {
	if restrictions != nil {
		// Type
		if restrictions.Type != "" {
			if stream.Type == "" {
				stream.Type = restrictions.Type
			} else if stream.Type != restrictions.Type {
				return errors.New("convert: unfulfilled restriction: type")
			}
		}

		// ID
		if restrictions.ID != "" {
			if stream.ID == "" {
				stream.ID = restrictions.ID
			} else if stream.ID != restrictions.ID {
				return errors.New("convert: unfulfilled restriction: ID")
			}
		}

		// HumanReadable.Name
		if restrictions.HumanReadable != nil && restrictions.HumanReadable.Name != nil {
			if stream.HumanReadable == nil || (stream.HumanReadable != nil && stream.HumanReadable.Name == nil) {
				stream.HumanReadable.Name = restrictions.HumanReadable.Name
			} else if stream.HumanReadable.Name != restrictions.HumanReadable.Name {
				return errors.New("convert: unfulfilled restriction: description")
			}
		}

		// Direction
		if restrictions.Direction != nil {
			if stream.Direction == nil {
				stream.Direction = restrictions.Direction
			} else if stream.Direction != restrictions.Direction {
				return errors.New("convert: unfulfilled restriction: mode")
			}
		}

		// URL
		if restrictions.URL != nil {
			if stream.URL == nil {
				stream.URL = restrictions.URL
			} else if stream.URL != restrictions.URL {
				return errors.New("convert: unfulfilled restriction: caching-server-url")
			}
		}

		// Labels
		for k, vr := range restrictions.Labels {
			if stream.Labels == nil {
				stream.Labels = make(map[string]string)
			}
			vs, ok := stream.Labels[k]
			if !ok {
				stream.Labels[k] = vr
			} else if vr != vs {
				return errors.New("convert: unfulfilled restriction: keyword")
			}
		}

		// Metadata.MimeType
		if restrictions.Metadata.MimeType != nil {
			if stream.Metadata.MimeType == nil {
				stream.Metadata.MimeType = restrictions.Metadata.MimeType
			} else if stream.Metadata.MimeType != restrictions.Metadata.MimeType {
				return errors.New("convert: unfulfilled restriction: mime-type")
			}
		}

		// Metadata.CodecType
		if restrictions.Metadata.CodecType != nil {
			if stream.Metadata.CodecType == nil {
				stream.Metadata.CodecType = restrictions.Metadata.CodecType
			} else if stream.Metadata.CodecType != restrictions.Metadata.CodecType {
				return errors.New("convert: unfulfilled restriction: codec-type")
			}
		}
	}

	return nil
}

func (c *wddToTasksConverter) convertMediaOrMetadataRestrictionToMedia(mediaRestrictions []nbmpv2.MediaParameter, metadataRestrictions []nbmpv2.MetadataParameter) (*enginev1.Media, error) {
	if len(mediaRestrictions) == 0 && len(metadataRestrictions) == 0 {
		return nil, nil
	}

	if len(mediaRestrictions)+len(metadataRestrictions) > 1 {
		return nil, fmt.Errorf("convert: invalid number of media or metadata restrictions '%d'", len(mediaRestrictions)+len(metadataRestrictions))
	}

	switch {
	case len(mediaRestrictions) == 1:
		return c.convertMediaParametersToMedia(&mediaRestrictions[0])
	case len(metadataRestrictions) == 1:
		return c.convertMetadataParametersToMedia(&metadataRestrictions[0])
	}

	return nil, fmt.Errorf("convert: illegal state reached: no restriction converted")
}

func (c *wddToTasksConverter) convertMediaParametersToMedia(mp *nbmpv2.MediaParameter) (*enginev1.Media, error) {
	if err := c.observeStreamID(mp.StreamID); err != nil {
		return nil, nil
	}

	c.mediaConverter.Reset(mp)
	m := &enginev1.Media{}
	if err := c.mediaConverter.Convert(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *wddToTasksConverter) convertMetadataParametersToMedia(mp *nbmpv2.MetadataParameter) (*enginev1.Media, error) {
	if err := c.observeStreamID(mp.StreamID); err != nil {
		return nil, nil
	}

	c.metadataConverter.Reset(mp)
	m := &enginev1.Media{}
	if err := c.metadataConverter.Convert(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *wddToTasksConverter) observeStreamID(streamID string) error {
	if _, exists := c.streamIDs[streamID]; exists {
		return fmt.Errorf("convert: duplicate stream-id '%s'", streamID)
	}
	c.streamIDs[streamID] = struct{}{}
	return nil
}
