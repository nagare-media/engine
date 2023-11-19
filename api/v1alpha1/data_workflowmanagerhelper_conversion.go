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

package v1alpha1

import (
	"fmt"
	"strconv"

	"github.com/nagare-media/engine/pkg/nbmp"
	nbmputils "github.com/nagare-media/engine/pkg/nbmp/utils"
	"github.com/nagare-media/models.go/base"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

func (d *WorkflowManagerHelperData) ConvertToNBMPTask(task *nbmpv2.Task) error {
	// initialize new Task

	*task = nbmpv2.Task{
		Scheme: &nbmpv2.Scheme{
			URI: nbmpv2.SchemaURI,
		},
		General: nbmpv2.General{
			// ID should be set during Task creation
			NBMPBrand: &nbmp.BrandNagareMediaEngineV1,
			State:     &nbmpv2.InstantiatedState,
		},
		Configuration: &nbmpv2.Configuration{
			Parameters: make([]nbmpv2.Parameter, 0, len(d.Task.Config)),
		},
		Failover: &nbmpv2.Failover{
			FailoverMode: nbmpv2.ExitFailoverMode,
		},
	}

	// General

	if d.Task.HumanReadable != nil && d.Task.HumanReadable.Name != nil {
		task.General.Name = *d.Task.HumanReadable.Name
	}

	if d.Task.HumanReadable != nil && d.Task.HumanReadable.Description != nil {
		task.General.Description = *d.Task.HumanReadable.Description
	}

	// Input

	for _, in := range d.Task.Inputs {
		for _, pb := range in.PortBindings {
			p := nbmpv2.Port{
				PortName: pb.ID,
				Bind: &nbmpv2.PortBinding{
					StreamID: in.ID,
				},
			}
			task.General.InputPorts = append(task.General.InputPorts, p)
		}

		switch in.Type {
		case MediaMediaType:
			mp := nbmpv2.MediaParameter{
				StreamID: in.ID,
				Mode:     (*nbmpv2.MediaAccessMode)(in.Direction),
			}

			if in.HumanReadable != nil && in.HumanReadable.Name != nil {
				mp.Name = *in.HumanReadable.Name
			}

			if in.URL != nil {
				mp.CachingServerURL = base.URI(*in.URL)
			}

			if in.Metadata.MimeType != nil {
				mp.MimeType = string(*in.Metadata.MimeType)
			}

			if in.Metadata.CodecType != nil {
				mp.MimeType = string(*in.Metadata.CodecType)
			}

			for k, v := range in.Labels {
				if v == "" {
					mp.Keywords = append(mp.Keywords, fmt.Sprintf("%s", k))
				} else {
					mp.Keywords = append(mp.Keywords, fmt.Sprintf("%s=%s", k, v))
				}
			}

			// TODO: handle multiple streams of the same type
			for _, s := range in.Metadata.Streams {
				switch {
				case s.Audio != nil:
					// TODO: implement
				case s.Video != nil:
					if s.Video.FrameRate != nil && s.Video.FrameRate.Average != nil {
						// support non-integer frame rates
						v := strconv.Itoa(int(*s.Video.FrameRate.Average))
						mp.VideoFormat = nbmputils.SetStringParameterValue(mp.VideoFormat, nbmp.VideoFormatFrameRateAverage, v)
					}

					if s.Duration != nil {
						mp.VideoFormat = nbmputils.SetStringParameterValue(mp.VideoFormat, nbmp.FormatFrameDuration, s.Duration.Duration.String())
					}
					// TODO: implement
				case s.Subtitle != nil:
					// TODO: implement
				case s.Data != nil:
					// TODO: implement
				}
			}

			task.Input.MediaParameters = append(task.Input.MediaParameters, mp)

		case MetadataMediaType:
			mp := nbmpv2.MetadataParameter{
				StreamID:         in.ID,
				Mode:             (*nbmpv2.MediaAccessMode)(in.Direction),
				CachingServerURL: (*base.URI)(in.URL),
			}

			if in.HumanReadable != nil && in.HumanReadable.Name != nil {
				mp.Name = *in.HumanReadable.Name
			}

			if in.Metadata.MimeType != nil {
				mp.MimeType = string(*in.Metadata.MimeType)
			}

			if in.Metadata.CodecType != nil {
				mp.MimeType = string(*in.Metadata.CodecType)
			}

			for k, v := range in.Labels {
				if v == "" {
					mp.Keywords = append(mp.Keywords, fmt.Sprintf("%s", k))
				} else {
					mp.Keywords = append(mp.Keywords, fmt.Sprintf("%s=%s", k, v))
				}
			}

			task.Input.MetadataParameters = append(task.Input.MetadataParameters, mp)

		default:
			return fmt.Errorf("unknown media type '%s'", in.Type)
		}
	}

	// Output

	for _, out := range d.Task.Outputs {
		for _, pb := range out.PortBindings {
			// check for duplicates?
			p := nbmpv2.Port{
				PortName: pb.ID,
				Bind: &nbmpv2.PortBinding{
					StreamID: out.ID,
				},
			}
			task.General.OutputPorts = append(task.General.OutputPorts, p)
		}

		switch out.Type {
		case MediaMediaType:
			mp := nbmpv2.MediaParameter{
				StreamID: out.ID,
				Mode:     (*nbmpv2.MediaAccessMode)(out.Direction),
			}

			if out.HumanReadable != nil && out.HumanReadable.Name != nil {
				mp.Name = *out.HumanReadable.Name
			}

			if out.URL != nil {
				mp.CachingServerURL = base.URI(*out.URL)
			}

			if out.Metadata.MimeType != nil {
				mp.MimeType = string(*out.Metadata.MimeType)
			}

			if out.Metadata.CodecType != nil {
				mp.MimeType = string(*out.Metadata.CodecType)
			}

			for k, v := range out.Labels {
				if v == "" {
					mp.Keywords = append(mp.Keywords, fmt.Sprintf("%s", k))
				} else {
					mp.Keywords = append(mp.Keywords, fmt.Sprintf("%s=%s", k, v))
				}
			}

			task.Output.MediaParameters = append(task.Output.MediaParameters, mp)

		case MetadataMediaType:
			mp := nbmpv2.MetadataParameter{
				StreamID:         out.ID,
				Mode:             (*nbmpv2.MediaAccessMode)(out.Direction),
				CachingServerURL: (*base.URI)(out.URL),
			}

			if out.HumanReadable != nil && out.HumanReadable.Name != nil {
				mp.Name = *out.HumanReadable.Name
			}

			if out.Metadata.MimeType != nil {
				mp.MimeType = string(*out.Metadata.MimeType)
			}

			if out.Metadata.CodecType != nil {
				mp.MimeType = string(*out.Metadata.CodecType)
			}

			for k, v := range out.Labels {
				if v == "" {
					mp.Keywords = append(mp.Keywords, fmt.Sprintf("%s", k))
				} else {
					mp.Keywords = append(mp.Keywords, fmt.Sprintf("%s=%s", k, v))
				}
			}

			task.Output.MetadataParameters = append(task.Output.MetadataParameters, mp)

		default:
			return fmt.Errorf("unknown media type '%s'", out.Type)
		}
	}

	// Configuration

	task.Configuration.Parameters = nbmputils.SetStringParameterValue(task.Configuration.Parameters, nbmp.EngineWorkflowIDParameterKey, d.Workflow.ID)
	task.Configuration.Parameters = nbmputils.SetStringParameterValue(task.Configuration.Parameters, nbmp.EngineTaskIDParameterKey, d.Task.ID)
	for k, v := range d.Task.Config {
		task.Configuration.Parameters = nbmputils.SetStringParameterValue(task.Configuration.Parameters, k, v)
	}

	return nil
}

// TODO: refactor and make dry
