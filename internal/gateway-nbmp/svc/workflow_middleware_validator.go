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
	"context"

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
		next: next,
	}
}

var _ nbmpsvcv2.WorkflowServiceMiddleware = WorkflowValidatorMiddleware
var _ nbmpsvcv2.WorkflowService = &workflowValidatorMiddleware{}

func (m *workflowValidatorMiddleware) Create(ctx context.Context, w *nbmpv2.Workflow) error {
	m.common(ctx, w)

	if err := nbmputils.AcknowledgeStatusToErr(nbmputils.UpdateAcknowledgeStatus(w.Acknowledge)); err != nil {
		return err
	}

	return m.next.Create(ctx, w)
}

func (m *workflowValidatorMiddleware) Update(ctx context.Context, w *nbmpv2.Workflow) error {
	m.common(ctx, w)

	if err := nbmputils.AcknowledgeStatusToErr(nbmputils.UpdateAcknowledgeStatus(w.Acknowledge)); err != nil {
		return err
	}

	return m.next.Update(ctx, w)
}

func (m *workflowValidatorMiddleware) common(ctx context.Context, w *nbmpv2.Workflow) {
	//// Scheme

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

	//// ClientAssistant

	//// Failover

	//// Monitoring

	//// Assertion

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

	// TODO: implement full validation
}
