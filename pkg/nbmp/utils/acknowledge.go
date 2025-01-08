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

package utils

import (
	"github.com/nagare-media/engine/pkg/nbmp"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

func UpdateAcknowledgeStatus(ack *nbmpv2.Acknowledge) nbmpv2.AcknowledgeStatus {
	ack.Status = nbmpv2.FulfilledAcknowledgeStatus
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

func AcknowledgeStatusToErr(s nbmpv2.AcknowledgeStatus) error {
	switch s {
	case nbmpv2.NotSupportedAcknowledgeStatus:
		return nbmp.ErrUnsupported
	case nbmpv2.FailedAcknowledgeStatus:
		return nbmp.ErrInvalid
	}
	return nil
}
