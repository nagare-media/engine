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

package v2

import (
	"context"
	"fmt"

	"github.com/nagare-media/engine/pkg/nbmp"
	nbmputils "github.com/nagare-media/engine/pkg/nbmp/utils"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

type workflowValidatorSpecLaxMiddleware struct {
	next WorkflowService
}

// WorkflowValidatorSpecLaxMiddleware validates given workflow according the NBMP specification allowing for some deviation.
func WorkflowValidatorSpecLaxMiddleware(next WorkflowService) WorkflowService {
	return &workflowValidatorSpecLaxMiddleware{
		next: next,
	}
}

var _ WorkflowServiceMiddleware = WorkflowValidatorSpecLaxMiddleware
var _ WorkflowService = &workflowValidatorSpecLaxMiddleware{}

func (m *workflowValidatorSpecLaxMiddleware) Create(ctx context.Context, w *nbmpv2.Workflow) error {
	m.common(ctx, w)

	if err := nbmputils.AcknowledgeStatusToErr(nbmputils.UpdateAcknowledgeStatus(w.Acknowledge)); err != nil {
		return err
	}

	return m.next.Create(ctx, w)
}

func (m *workflowValidatorSpecLaxMiddleware) Update(ctx context.Context, w *nbmpv2.Workflow) error {
	m.common(ctx, w)

	if err := nbmputils.AcknowledgeStatusToErr(nbmputils.UpdateAcknowledgeStatus(w.Acknowledge)); err != nil {
		return err
	}

	return m.next.Update(ctx, w)
}

func (m *workflowValidatorSpecLaxMiddleware) Delete(ctx context.Context, w *nbmpv2.Workflow) error {
	// no call to common as we only care about the workflow ID

	// workflows must have an ID
	if w.General.ID == "" {
		w.Acknowledge.Failed = append(w.Acknowledge.Failed, "$.general.id")
	}

	if err := nbmputils.AcknowledgeStatusToErr(nbmputils.UpdateAcknowledgeStatus(w.Acknowledge)); err != nil {
		return err
	}

	return m.next.Delete(ctx, w)
}

func (m *workflowValidatorSpecLaxMiddleware) Retrieve(ctx context.Context, w *nbmpv2.Workflow) error {
	// no call to common as we only care about the workflow ID

	// workflows must have an ID
	if w.General.ID == "" {
		w.Acknowledge.Failed = append(w.Acknowledge.Failed, "$.general.id")
	}

	if err := nbmputils.AcknowledgeStatusToErr(nbmputils.UpdateAcknowledgeStatus(w.Acknowledge)); err != nil {
		return err
	}

	return m.next.Retrieve(ctx, w)
}

func (m *workflowValidatorSpecLaxMiddleware) common(ctx context.Context, w *nbmpv2.Workflow) {
	//// Scheme

	// validate scheme URI
	if w.Scheme != nil && w.Scheme.URI != nbmpv2.SchemaURI {
		w.Acknowledge.Failed = append(w.Acknowledge.Failed, "$.scheme.uri")
	}

	//// General

	// workflows must have an ID
	// note: we allow setting an ID on workflow creation.
	if w.General.ID == "" {
		w.Acknowledge.Failed = append(w.Acknowledge.Failed, "$.general.id")
	}

	// workflows have no input-ports
	if w.General.InputPorts != nil || len(w.General.InputPorts) != 0 {
		w.Acknowledge.Failed = append(w.Acknowledge.Failed, "$.general.input-ports")
	}

	// workflows have no output-ports
	if w.General.OutputPorts != nil || len(w.General.OutputPorts) != 0 {
		w.Acknowledge.Failed = append(w.Acknowledge.Failed, "$.general.output-ports")
	}

	// workflows are no group
	if w.General.IsGroup != nil && *w.General.IsGroup {
		w.Acknowledge.Failed = append(w.Acknowledge.Failed, "$.general.is-group")
	}

	//// Reporting

	//// Input

	// NBMP spec does not allow workflows with no input. Let's be more flexible and allow in-workflow media generation.
	// if len(w.Input.MediaParameters) == 0 && len(w.Input.MetadataParameters) == 0 {
	// 	w.Acknowledge.Failed = append(w.Acknowledge.Failed, "$.input.media-parameters", "$.input.metadata-parameters")
	// }

	for i, mp := range w.Input.MediaParameters {
		// StreamID must be set
		if mp.StreamID == "" {
			w.Acknowledge.Failed = append(w.Acknowledge.Failed, fmt.Sprintf("$.input.media-parameters[%d].stream-id", i))
		}
	}

	for i, mp := range w.Input.MetadataParameters {
		// StreamID must be set
		if mp.StreamID == "" {
			w.Acknowledge.Failed = append(w.Acknowledge.Failed, fmt.Sprintf("$.input.metadata-parameters[%d].stream-id", i))
		}
	}

	//// Output

	// NBMP spec does not allow workflows with no output. Let's be more flexible and allow empty outputs.
	// if len(w.Output.MediaParameters) == 0 && len(w.Output.MetadataParameters) == 0 {
	// 	w.Acknowledge.Failed = append(w.Acknowledge.Failed, "$.output.media-parameters", "$.output.metadata-parameters")
	// }

	for i, mp := range w.Output.MediaParameters {
		// StreamID must be set
		if mp.StreamID == "" {
			w.Acknowledge.Failed = append(w.Acknowledge.Failed, fmt.Sprintf("$.output.media-parameters[%d].stream-id", i))
		}
	}

	for i, mp := range w.Output.MetadataParameters {
		// StreamID must be set
		if mp.StreamID == "" {
			w.Acknowledge.Failed = append(w.Acknowledge.Failed, fmt.Sprintf("$.output.metadata-parameters[%d].stream-id", i))
		}
	}

	//// Processing

	// Image (we will ignore this fields anyways)
	for i, img := range w.Processing.Image {
		// URL shall not be set
		if img.URL != "" {
			w.Acknowledge.Failed = append(w.Acknowledge.Failed, fmt.Sprintf("$.processing.image[%d].url", i))
		}
	}

	for i, fr := range w.Processing.FunctionRestrictions {
		if fr.General != nil {
			// functions must have an instance
			if fr.Instance == "" {
				w.Acknowledge.Failed = append(w.Acknowledge.Failed,
					fmt.Sprintf("$.processing.function-restrictions[%d].instance", i))
			}

			// functions must have an ID
			if fr.General.ID == "" {
				w.Acknowledge.Failed = append(w.Acknowledge.Failed,
					fmt.Sprintf("$.processing.function-restrictions[%d].general.id", i))
			}

			// we don't support (nested) function groups
			if fr.General.IsGroup != nil && *fr.General.IsGroup {
				w.Acknowledge.Unsupported = append(w.Acknowledge.Unsupported,
					fmt.Sprintf("$.processing.function-restrictions[%d].general.is-group", i))
			}

			// NBMP brand
			if fr.General.NBMPBrand != nil &&
				*fr.General.NBMPBrand != "" &&
				*fr.General.NBMPBrand != nbmp.BrandNagareMediaEngineV1 {
				w.Acknowledge.Unsupported = append(w.Acknowledge.Unsupported,
					fmt.Sprintf("$.processing.function-restrictions[%d].general.nbmp-brand", i))
			}
		}
	}

	//// ClientAssistant

	if w.ClientAssistant != nil {
		if w.ClientAssistant.ClientAssistanceFlag {
			w.Acknowledge.Unsupported = append(w.Acknowledge.Unsupported, "$.client-assistant.client-assistance-flag")
		}
	}

	//// Failover

	//// Monitoring

	//// Assertion

	//// Notification

	//// Security

	//// Scale

	//// Schedule

	// TODO: implement workflow validation
}
