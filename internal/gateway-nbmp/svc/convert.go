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
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/mitchellh/mapstructure"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/pkg/apis/functions"
	"github.com/nagare-media/engine/pkg/apis/meta"
	"github.com/nagare-media/engine/pkg/maps"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

const (
	IsNBMPLabel = "engine.nagare.media/is-nbmp"

	MediaLocationDefaultStreaming = "default-streaming"
	MediaLocationDefaultStep      = "default-step"
)

func (s *workflowService) wddToWorkflow(nbmpWf *nbmpv2.Workflow) (*enginev1.Workflow, error) {
	wf := &enginev1.Workflow{
		ObjectMeta: metav1.ObjectMeta{
			Name: nbmpWf.General.ID,
			Annotations: map[string]string{
				IsNBMPLabel: "true",
			},
		},
		Spec: enginev1.WorkflowSpec{
			HumanReadable: &enginev1.HumanReadableWorkflowDescription{
				Name:        ptr.To[string](strings.Clone(nbmpWf.General.Name)),
				Description: ptr.To[string](strings.Clone(nbmpWf.General.Description)),
			},
		},
	}

	wfInputMediaCfg := make(map[string]any)
	for _, mp := range nbmpWf.Input.MediaParameters {
		c := map[string]any{
			"name":      mp.Name,
			"mime-type": mp.MimeType,
			"protocol":  mp.Protocol,
			"url":       mp.CachingServerURL,
		}
		wfInputMediaCfg[mp.StreamID] = c
	}

	wfInputMetadataCfg := make(map[string]any)
	for _, mp := range nbmpWf.Input.MetadataParameters {
		c := map[string]any{
			"name":      mp.Name,
			"mime-type": mp.MimeType,
			"protocol":  mp.Protocol,
			"url":       mp.CachingServerURL,
		}
		wfInputMetadataCfg[mp.StreamID] = c
	}

	wfOutputMediaCfg := make(map[string]any)
	for _, mp := range nbmpWf.Output.MediaParameters {
		c := map[string]any{
			"name":      mp.Name,
			"mime-type": mp.MimeType,
			"protocol":  mp.Protocol,
			"url":       mp.CachingServerURL,
		}
		wfOutputMediaCfg[mp.StreamID] = c
	}

	wfOutputMetadataCfg := make(map[string]any)
	for _, mp := range nbmpWf.Output.MetadataParameters {
		c := map[string]any{
			"name":      mp.Name,
			"mime-type": mp.MimeType,
			"protocol":  mp.Protocol,
			"url":       mp.CachingServerURL,
		}
		wfOutputMetadataCfg[mp.StreamID] = c
	}

	wfCfg := map[string]any{
		"inputs": map[string]any{
			"media":    wfInputMediaCfg,
			"metadata": wfInputMetadataCfg,
		},
		"outputs": map[string]any{
			"media":    wfOutputMediaCfg,
			"metadata": wfOutputMetadataCfg,
		},
	}

	cfgRaw, err := json.Marshal(wfCfg)
	if err != nil {
		return nil, err
	}
	wf.Spec.Config = &apiextensionsv1.JSON{Raw: cfgRaw}

	// TODO: add media locations

	return wf, nil
}

func (s *workflowService) wddToTasks(nbmpWf *nbmpv2.Workflow, wf *enginev1.Workflow) ([]*enginev1.Task, error) {
	tasks := make([]*enginev1.Task, 0)

	// check execution mode

	defaultExecMode := nbmpv2.StreamingExecutionMode
	if nbmpWf.Requirement.WorkflowTask != nil && nbmpWf.Requirement.WorkflowTask.ExecutionMode != nil {
		if *nbmpWf.Requirement.WorkflowTask.ExecutionMode != nbmpv2.HybridExecutionMode {
			defaultExecMode = *nbmpWf.Requirement.WorkflowTask.ExecutionMode
		}
	}

	taskIDToExecutionMode := make(map[string]nbmpv2.ExecutionMode)
	for i, fr := range nbmpWf.Processing.FunctionRestrictions {
		if _, ok := taskIDToExecutionMode[fr.Instance]; ok {
			// task seen multiple times
			nbmpWf.Acknowledge.Failed = append(nbmpWf.Acknowledge.Failed,
				fmt.Sprintf("$.processing.function-restrictions[%d].instance", i))
		} else {
			taskIDToExecutionMode[fr.Instance] = defaultExecMode
			if fr.Requirements != nil &&
				fr.Requirements.WorkflowTask != nil &&
				fr.Requirements.WorkflowTask.ExecutionMode != nil {
				if *fr.Requirements.WorkflowTask.ExecutionMode == nbmpv2.HybridExecutionMode {
					// what does this exec mode mean in the context of a single task?
					nbmpWf.Acknowledge.Unsupported = append(nbmpWf.Acknowledge.Unsupported,
						fmt.Sprintf("$.processing.function-restrictions[%d].requirements.workflow-task.execution-mode", i))
				}
				taskIDToExecutionMode[fr.Instance] = *fr.Requirements.WorkflowTask.ExecutionMode
			}
		}
	}

	// check connection map

	taskIDToFunctionID := make(map[string]string)
	taskIDToInputToPortMapping := make(map[string]map[string]*nbmpv2.ConnectionMappingPort)
	taskIDToOutputToExecutionMode := make(map[string]map[string]nbmpv2.ExecutionMode)

	for i, con := range nbmpWf.Processing.ConnectionMap {
		// update function ID map
		if fid, ok := taskIDToFunctionID[con.From.Instance]; ok {
			if fid != con.From.ID {
				nbmpWf.Acknowledge.Failed = append(nbmpWf.Acknowledge.Failed,
					fmt.Sprintf("$.processing.connection-map[%d].from.id", i))
			}
		} else {
			taskIDToFunctionID[con.From.Instance] = con.From.ID
		}
		if fid, ok := taskIDToFunctionID[con.To.Instance]; ok {
			if fid != con.To.ID {
				nbmpWf.Acknowledge.Failed = append(nbmpWf.Acknowledge.Failed,
					fmt.Sprintf("$.processing.connection-map[%d].to.id", i))
			}
		} else {
			taskIDToFunctionID[con.To.Instance] = con.To.ID
		}

		// update input port map
		if _, ok := taskIDToInputToPortMapping[con.To.Instance]; !ok {
			taskIDToInputToPortMapping[con.To.Instance] = make(map[string]*nbmpv2.ConnectionMappingPort)
		}
		if _, ok := taskIDToInputToPortMapping[con.To.Instance][con.To.PortName]; ok {
			// input is already connected
			nbmpWf.Acknowledge.Failed = append(nbmpWf.Acknowledge.Failed,
				fmt.Sprintf("$.processing.connection-map[%d].to.port-name", i))
		} else {
			taskIDToInputToPortMapping[con.To.Instance][con.To.PortName] = &con.From
		}

		// update output execution mode map
		if _, ok := taskIDToOutputToExecutionMode[con.From.Instance]; !ok {
			taskIDToOutputToExecutionMode[con.From.Instance] = make(map[string]nbmpv2.ExecutionMode)
		}
		if mode, ok := taskIDToOutputToExecutionMode[con.From.Instance][con.From.PortName]; ok &&
			mode != taskIDToExecutionMode[con.To.Instance] {
			// output of From is already sent to a task that has a different execution mode -> we don't support that
			nbmpWf.Acknowledge.Unsupported = append(nbmpWf.Acknowledge.Unsupported,
				fmt.Sprintf("$.processing.connection-map[%d].to.instance", i))
		} else {
			taskIDToOutputToExecutionMode[con.From.Instance][con.From.PortName] = taskIDToExecutionMode[con.To.Instance]
		}
	}

	// check function restrictions

	for i, fr := range nbmpWf.Processing.FunctionRestrictions {
		// function ID
		fid, ok := taskIDToFunctionID[fr.Instance]
		if !ok {
			nbmpWf.Acknowledge.Failed = append(nbmpWf.Acknowledge.Failed,
				fmt.Sprintf("$.processing.function-restrictions[%d].instance", i))
		}

		// template patches
		tplPatches, err := s.functionRestrictionToJobTemplateSpec(&fr)
		if err != nil {
			return nil, err
		}

		// MPE selector
		mpeSel, err := s.functionRestrictionToMediaProcessingEntitySelector(&fr)
		if err != nil {
			return nil, err
		}

		// task config
		ioCfg := functions.GenericFunctionConfig{
			Inputs:  make(map[string]any),
			Outputs: make(map[string]any),
		}

		// execution mode
		execMode, ok := taskIDToExecutionMode[fr.Instance]
		if !ok {
			// this should not happen
			return nil, errors.New("wddToTasks: could not determine execution mode")
		}
		inputMediaLocation := MediaLocationDefaultStreaming
		if execMode == nbmpv2.StepExecutionMode {
			inputMediaLocation = MediaLocationDefaultStep
		}

		// task inputs
		for portName, from := range taskIDToInputToPortMapping[fr.Instance] {
			ioCfg.Inputs[portName] = functions.MediaRef{
				URL: fmt.Sprintf("nagare://%s/%s/%s", inputMediaLocation, from.Instance, from.PortName),
			}
		}

		// workflow inputs
		for j, port := range fr.General.InputPorts {
			if port.Bind.StreamID == nil {
				continue
			}

			if _, ok := ioCfg.Inputs[port.PortName]; ok {
				// already connected to a task input
				nbmpWf.Acknowledge.Failed = append(nbmpWf.Acknowledge.Failed,
					fmt.Sprintf("$.processing.function-restrictions[%d].general.input-ports[%d].port-name", i, j))
				continue
			}

			// media or metadata?
			var streamType string
			if containsStreamID(*port.Bind.StreamID, nbmpWf.Input.MediaParameters) {
				streamType = "media"
			} else if containsStreamID(*port.Bind.StreamID, nbmpWf.Input.MetadataParameters) {
				streamType = "metadata"
			} else {
				nbmpWf.Acknowledge.Failed = append(nbmpWf.Acknowledge.Failed,
					fmt.Sprintf("$.processing.function-restrictions[%d].general.input-ports[%d].bind.stream-id", i, j))
				continue
			}

			ioCfg.Inputs[port.PortName] = functions.MediaRef{
				URL: fmt.Sprintf("{{ wf.cfg.inputs.%s['%s'].url }}", streamType, *port.Bind.StreamID),
			}
		}

		// task output
		for portName, em := range taskIDToOutputToExecutionMode[fr.Instance] {
			outputMediaLocation := MediaLocationDefaultStreaming
			if em == nbmpv2.StepExecutionMode {
				outputMediaLocation = MediaLocationDefaultStep
			}
			ioCfg.Outputs[portName] = functions.MediaRef{
				URL: fmt.Sprintf("nagare://%s/%s/%s", outputMediaLocation, fr.Instance, portName),
			}
		}

		// workflow output
		for j, port := range fr.General.OutputPorts {
			if port.Bind.StreamID == nil {
				continue
			}

			// media or metadata?
			var streamType string
			if containsStreamID(*port.Bind.StreamID, nbmpWf.Output.MediaParameters) {
				streamType = "media"
			} else if containsStreamID(*port.Bind.StreamID, nbmpWf.Output.MetadataParameters) {
				streamType = "metadata"
			} else {
				nbmpWf.Acknowledge.Failed = append(nbmpWf.Acknowledge.Failed,
					fmt.Sprintf("$.processing.function-restrictions[%d].general.output-ports[%d].bind.stream-id", i, j))
				continue
			}

			ioCfg.Outputs[port.PortName] = functions.MediaRef{
				URL: fmt.Sprintf("{{ wf.cfg.outputs.%s['%s'].url }}", streamType, *port.Bind.StreamID),
			}
		}

		// construct config map
		var cfgMap map[string]any
		err = mapstructure.Decode(ioCfg, &cfgMap)
		if err != nil {
			return nil, err
		}

		// create config for task
		taskCfg := make(map[string]any)
		// TODO: fill taskCfg

		// merge config maps
		cfgMap = maps.Merge(cfgMap, taskCfg)

		// create
		rawCfg, err := json.Marshal(cfgMap)
		if err != nil {
			return nil, err
		}

		// create task
		tasks = append(tasks, &enginev1.Task{
			ObjectMeta: metav1.ObjectMeta{
				Name: fr.Instance,
				Annotations: map[string]string{
					IsNBMPLabel: "true",
				},
			},
			Spec: enginev1.TaskSpec{
				HumanReadable: &enginev1.HumanReadableTaskDescription{
					Name:        ptr.To[string](strings.Clone(fr.General.Name)),
					Description: ptr.To[string](strings.Clone(fr.General.Description)),
				},
				MediaProcessingEntitySelector: mpeSel,
				WorkflowRef:                   meta.LocalObjectReference{Name: wf.Name},
				FunctionRef:                   &meta.LocalObjectReference{Name: fid},
				Config:                        &apiextensionsv1.JSON{Raw: rawCfg},
				TemplatePatches:               tplPatches,
				JobFailurePolicy:              &enginev1.JobFailurePolicy{DefaultAction: &enginev1.JobFailurePolicyActionFailWorkflow},
			},
		})
	}

	err := acknowledgeStatusToErr(updateAcknowledgeStatus(nbmpWf.Acknowledge))
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *workflowService) functionRestrictionToMediaProcessingEntitySelector(fr *nbmpv2.FunctionRestriction) (*metav1.LabelSelector, error) {
	sel := &metav1.LabelSelector{
		MatchLabels: make(map[string]string),
	}

	if fr.Requirements != nil && fr.Requirements.Hardware != nil && fr.Requirements.Hardware.Placement != nil {
		sel.MatchLabels[enginev1.BetaMediaProcessingEntityLocationLabel] = string(*fr.Requirements.Hardware.Placement)
	} else {
		sel.MatchLabels[enginev1.BetaIsDefaultMediaProcessingEntityAnnotation] = "true"
	}

	return sel, nil
}

func (s *workflowService) functionRestrictionToJobTemplateSpec(fr *nbmpv2.FunctionRestriction) (*batchv1.JobTemplateSpec, error) {
	var js *batchv1.JobTemplateSpec

	if fr.Requirements != nil && fr.Requirements.Hardware != nil {
		rl := corev1.ResourceList{}

		// TODO: detect overflow or change uint64 to something more reasonable in the NBMP model

		// number of CPU cores
		if fr.Requirements.Hardware.VCPU != nil {
			n := *fr.Requirements.Hardware.VCPU
			rl[corev1.ResourceCPU] = *resource.NewMilliQuantity(int64(1000*n), resource.DecimalSI)
		}

		// number of GPUs
		if fr.Requirements.Hardware.VGPU != nil {
			n := *fr.Requirements.Hardware.VGPU
			// TODO: make this configurable per Task.
			rl[s.cfg.Services.DefaultKubernetesGPUResource] = *resource.NewQuantity(int64(n), resource.DecimalSI)
		}

		// RAM in megabytes
		if fr.Requirements.Hardware.RAM != nil {
			n := *fr.Requirements.Hardware.RAM
			rl[corev1.ResourceMemory] = *resource.NewQuantity(int64(n*1024*1024), resource.BinarySI)
		}

		// disk in gigabytes
		if fr.Requirements.Hardware.Disk != nil {
			n := *fr.Requirements.Hardware.Disk
			rl[corev1.ResourceStorage] = *resource.NewQuantity(int64(n*1024*1024*1024), resource.BinarySI)
			rl[corev1.ResourceEphemeralStorage] = rl[corev1.ResourceStorage]
		}

		js = &batchv1.JobTemplateSpec{
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

	return js, nil
}

func containsStreamID(streamID string, obj any) bool {
	switch mps := obj.(type) {
	case []nbmpv2.MediaParameter:
		for _, mp := range mps {
			if mp.StreamID == streamID {
				return true
			}
		}
	case []nbmpv2.MetadataParameter:
		for _, mp := range mps {
			if mp.StreamID == streamID {
				return true
			}
		}
	}
	return false
}
