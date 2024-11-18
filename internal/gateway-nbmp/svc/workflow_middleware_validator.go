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

package svc

import (
	"context"
	"fmt"

	"github.com/nagare-media/engine/pkg/nbmp"
	nbmpsvcv2 "github.com/nagare-media/engine/pkg/nbmp/svc/v2"
	nbmputils "github.com/nagare-media/engine/pkg/nbmp/utils"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

type workflowValidatorMiddleware struct {
	nbmpsvcv2.WorkflowDummyMiddleware

	next nbmpsvcv2.WorkflowService
}

// WorkflowValidatorMiddleware validates given workflow according the nagare media engine.
func WorkflowValidatorMiddleware(next nbmpsvcv2.WorkflowService) nbmpsvcv2.WorkflowService {
	return &workflowValidatorMiddleware{
		WorkflowDummyMiddleware: nbmpsvcv2.WorkflowDummyMiddleware{
			Next: next,
		},
		next: next,
	}
}

var _ nbmpsvcv2.WorkflowServiceMiddleware = WorkflowValidatorMiddleware
var _ nbmpsvcv2.WorkflowService = &workflowValidatorMiddleware{}

func (m *workflowValidatorMiddleware) Create(ctx context.Context, w *nbmpv2.Workflow) error {
	m.common(w)

	if err := nbmputils.AcknowledgeStatusToErr(nbmputils.UpdateAcknowledgeStatus(w.Acknowledge)); err != nil {
		return err
	}

	return m.next.Create(ctx, w)
}

func (m *workflowValidatorMiddleware) Update(ctx context.Context, w *nbmpv2.Workflow) error {
	m.common(w)

	if err := nbmputils.AcknowledgeStatusToErr(nbmputils.UpdateAcknowledgeStatus(w.Acknowledge)); err != nil {
		return err
	}

	return m.next.Update(ctx, w)
}

func (m *workflowValidatorMiddleware) common(w *nbmpv2.Workflow) {
	//// General

	// we only support nagare media engine NBMP brand
	if w.General.NBMPBrand != nil &&
		*w.General.NBMPBrand != "" &&
		*w.General.NBMPBrand != nbmp.BrandNagareMediaEngineV1 {
		w.Acknowledge.Unsupported = append(w.Acknowledge.Unsupported, "$.general.nbmp-brand")
	}

	//// Repository

	// we don't support external repositories
	if w.Repository != nil &&
		w.Repository.Mode != nil && *w.Repository.Mode != nbmpv2.AvailableRepositoryMode {
		w.Acknowledge.Unsupported = append(w.Acknowledge.Unsupported, "$.repository.mode")
	}

	//// Input

	//// Output

	//// Processing

	for i, fr := range w.Processing.FunctionRestrictions {
		if fr.General != nil {
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

	//// Requirement

	//// Step

	// we don't support Step
	if w.Step != nil {
		w.Acknowledge.Unsupported = append(w.Acknowledge.Unsupported, "$.Step")
	}

	//// ClientAssistant

	// we don't support ClientAssistant
	if w.ClientAssistant != nil {
		w.Acknowledge.Unsupported = append(w.Acknowledge.Unsupported, "$.ClientAssistant")
	}

	//// Failover

	//// Monitoring

	// we don't support Monitoring
	if w.Monitoring != nil {
		w.Acknowledge.Unsupported = append(w.Acknowledge.Unsupported, "$.Monitoring")
	}

	//// Assertion

	// we don't support Assertion
	if w.Assertion != nil {
		w.Acknowledge.Unsupported = append(w.Acknowledge.Unsupported, "$.Assertion")
	}

	//// Reporting

	// we don't support Reporting
	if w.Reporting != nil {
		w.Acknowledge.Unsupported = append(w.Acknowledge.Unsupported, "$.Reporting")
	}

	//// Notification

	// we don't support Notification
	if w.Notification != nil {
		w.Acknowledge.Unsupported = append(w.Acknowledge.Unsupported, "$.Notification")
	}

	//// Security

	// we don't support Security
	if w.Security != nil {
		w.Acknowledge.Unsupported = append(w.Acknowledge.Unsupported, "$.Security")
	}

	//// Scale

	// we don't support Scale
	if w.Scale != nil {
		w.Acknowledge.Unsupported = append(w.Acknowledge.Unsupported, "$.Scale")
	}

	//// Schedule

	// we don't support Schedule
	if w.Schedule != nil {
		w.Acknowledge.Unsupported = append(w.Acknowledge.Unsupported, "$.Schedule")
	}

	// TODO: implement workflow validation
}
