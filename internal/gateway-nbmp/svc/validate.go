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
	"fmt"

	"github.com/nagare-media/models.go/base"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

var NBMPBrandNagareEnginev1alpha1 = base.URI("urn:nagare-media:engine:schema:nbmp:v1apha1")

func ValidateWorkflowCommon(wf *nbmpv2.Workflow) error {
	//// Scheme

	// validate scheme URI
	if wf.Scheme != nil && wf.Scheme.URI != nbmpv2.SchemaURI {
		wf.Acknowledge.Failed = append(wf.Acknowledge.Failed, "$.scheme.uri")
	}

	//// General

	// workflows have no input-ports
	if wf.General.InputPorts != nil || len(wf.General.InputPorts) != 0 {
		wf.Acknowledge.Failed = append(wf.Acknowledge.Failed, "$.general.input-ports")
	}

	// workflows have no output-ports
	if wf.General.OutputPorts != nil || len(wf.General.OutputPorts) != 0 {
		wf.Acknowledge.Failed = append(wf.Acknowledge.Failed, "$.general.output-ports")
	}

	// workflows are no group
	if wf.General.IsGroup != nil && *wf.General.IsGroup {
		wf.Acknowledge.Failed = append(wf.Acknowledge.Failed, "$.general.is-group")
	}

	// NBMP brand
	if wf.General.NBMPBrand != nil &&
		*wf.General.NBMPBrand != "" &&
		*wf.General.NBMPBrand != NBMPBrandNagareEnginev1alpha1 {
		wf.Acknowledge.Unsupported = append(wf.Acknowledge.Unsupported, "$.general.nbmp-brand")
	}

	//// Repository

	// we don't support external repositories
	if wf.Repository != nil && wf.Repository.Mode != nil && *wf.Repository.Mode != nbmpv2.AvailableRepositoryMode {
		wf.Acknowledge.Unsupported = append(wf.Acknowledge.Unsupported, "$.repository.mode")
	}

	//// Input

	// NBMP spec does not allow workflows with no input. Let's be more flexible and allow in-workflow media generation.
	// if len(wf.Input.MediaParameters) == 0 && len(wf.Input.MetadataParameters) == 0 {
	// 	wf.Acknowledge.Failed = append(wf.Acknowledge.Failed, "$.input.media-parameters", "$.input.metadata-parameters")
	// }

	for i, mp := range wf.Input.MediaParameters {
		// StreamID must be set
		if mp.StreamID == "" {
			wf.Acknowledge.Failed = append(wf.Acknowledge.Failed, fmt.Sprintf("$.input.media-parameters[%d].stream-id", i))
		}

		// TODO: we currently only support pull based inputs
		if mp.Mode != nil && *mp.Mode != nbmpv2.PullMediaAccessMode {
			wf.Acknowledge.Unsupported = append(wf.Acknowledge.Unsupported,
				fmt.Sprintf("$.input.media-parameters[%d].mode", i))
		}
	}

	for i, mp := range wf.Input.MetadataParameters {
		// StreamID must be set
		if mp.StreamID == "" {
			wf.Acknowledge.Failed = append(wf.Acknowledge.Failed, fmt.Sprintf("$.input.metadata-parameters[%d].stream-id", i))
		}

		// TODO: we currently only support pull based inputs
		if mp.Mode != nil && *mp.Mode != nbmpv2.PullMediaAccessMode {
			wf.Acknowledge.Unsupported = append(wf.Acknowledge.Unsupported,
				fmt.Sprintf("$.input.metadata-parameters[%d].mode", i))
		}
	}

	//// Output

	// NBMP spec does not allow workflows with no output. Let's be more flexible and allow empty outputs.
	// if len(wf.Output.MediaParameters) == 0 && len(wf.Output.MetadataParameters) == 0 {
	// 	wf.Acknowledge.Failed = append(wf.Acknowledge.Failed, "$.output.media-parameters", "$.input.metadata-parameters")
	// }

	for i, mp := range wf.Output.MediaParameters {
		// StreamID must be set
		if mp.StreamID == "" {
			wf.Acknowledge.Failed = append(wf.Acknowledge.Failed, fmt.Sprintf("$.output.media-parameters[%d].stream-id", i))
		}
	}

	for i, mp := range wf.Output.MetadataParameters {
		// StreamID must be set
		if mp.StreamID == "" {
			wf.Acknowledge.Failed = append(wf.Acknowledge.Failed, fmt.Sprintf("$.output.metadata-parameters[%d].stream-id", i))
		}
	}

	//// Processing

	// Image (we will ignore this fields anyways)
	for i, img := range wf.Processing.Image {
		// URL shall not be set
		if img.URL != "" {
			wf.Acknowledge.Failed = append(wf.Acknowledge.Failed, fmt.Sprintf("$.processing.image[%d].url", i))
		}
	}

	for i, fr := range wf.Processing.FunctionRestrictions {
		if fr.General != nil {
			// functions must have an instance
			if fr.Instance == "" {
				wf.Acknowledge.Failed = append(wf.Acknowledge.Failed,
					fmt.Sprintf("$.processing.function-restrictions[%d].instance", i))
			}

			// functions must have an ID
			if fr.General.ID == "" {
				wf.Acknowledge.Failed = append(wf.Acknowledge.Failed,
					fmt.Sprintf("$.processing.function-restrictions[%d].general.id", i))
			}

			// we don't support (nested) function groups
			if fr.General.IsGroup != nil && *fr.General.IsGroup {
				wf.Acknowledge.Unsupported = append(wf.Acknowledge.Unsupported,
					fmt.Sprintf("$.processing.function-restrictions[%d].general.is-group", i))
			}

			// NBMP brand
			if fr.General.NBMPBrand != nil &&
				*fr.General.NBMPBrand != "" &&
				*fr.General.NBMPBrand != NBMPBrandNagareEnginev1alpha1 {
				wf.Acknowledge.Unsupported = append(wf.Acknowledge.Unsupported,
					fmt.Sprintf("$.processing.function-restrictions[%d].general.nbmp-brand", i))
			}
		}
	}

	//// ClientAssistant

	if wf.ClientAssistant != nil {
		if wf.ClientAssistant.ClientAssistanceFlag {
			wf.Acknowledge.Unsupported = append(wf.Acknowledge.Unsupported, "$.client-assistant.client-assistance-flag")
		}
	}

	//// Failover

	//// Monitoring

	//// Assertion

	//// Reporting

	//// Notification

	//// Security

	// We don't support enforced security constrains. MPE admins and function developers are responsible.
	if wf.Security != nil {
		wf.Acknowledge.Unsupported = append(wf.Acknowledge.Unsupported, "$.security")
	}

	//// Scale

	//// Schedule

	return acknowledgeStatusToErr(updateAcknowledgeStatus(wf.Acknowledge))
}

func ValidateWorkflowCreate(wf *nbmpv2.Workflow) error {
	// ignore err and return one ourselves
	_ = ValidateWorkflowCommon(wf)

	// the ID should initially be empty
	if wf.General.ID != "" {
		wf.Acknowledge.Failed = append(wf.Acknowledge.Failed, "$.general.id")
	}

	return acknowledgeStatusToErr(updateAcknowledgeStatus(wf.Acknowledge))
}

func ValidateWorkflowUpdate(wf *nbmpv2.Workflow) error {
	// ignore err and return one ourselves
	_ = ValidateWorkflowCommon(wf)

	// workflows must have an ID
	if wf.General.ID == "" {
		wf.Acknowledge.Failed = append(wf.Acknowledge.Failed, "$.general.id")
	}

	return acknowledgeStatusToErr(updateAcknowledgeStatus(wf.Acknowledge))
}

func ValidateWorkflowDelete(wf *nbmpv2.Workflow) error {
	// workflows must have an ID
	if wf.General.ID == "" {
		wf.Acknowledge.Failed = append(wf.Acknowledge.Failed, "$.general.id")
	}

	return acknowledgeStatusToErr(updateAcknowledgeStatus(wf.Acknowledge))
}

func ValidateWorkflowRetrieve(wf *nbmpv2.Workflow) error {
	// workflows must have an ID
	if wf.General.ID == "" {
		wf.Acknowledge.Failed = append(wf.Acknowledge.Failed, "$.general.id")
	}

	return acknowledgeStatusToErr(updateAcknowledgeStatus(wf.Acknowledge))
}

func acknowledgeStatusToErr(s nbmpv2.AcknowledgeStatus) error {
	switch s {
	case nbmpv2.NotSupportedAcknowledgeStatus:
		return ErrUnsupported
	case nbmpv2.FailedAcknowledgeStatus:
		return ErrInvalid
	}
	return nil
}

func updateAcknowledgeStatus(ack *nbmpv2.Acknowledge) nbmpv2.AcknowledgeStatus {
	if len(ack.Partial) > 0 {
		ack.Status = nbmpv2.PartiallyFulfilledAcknowledgeStatus
	}
	if len(ack.Unsupported) > 0 {
		ack.Status = nbmpv2.NotSupportedAcknowledgeStatus
	}
	if len(ack.Failed) > 0 {
		ack.Status = nbmpv2.FailedAcknowledgeStatus
	}
	return ack.Status
}
