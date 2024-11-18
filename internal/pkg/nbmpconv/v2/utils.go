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
	"strings"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

func parseKeyword(k string) (string, string) {
	i := strings.IndexRune(k, '=')
	if i > 0 && i != len(k)-1 {
		return k[:i], k[i+1:]
	}
	return k, ""
}

func encodeKeyword(k, v string) string {
	buf := strings.Builder{}
	buf.Grow(len(k) + 1 + len(v))
	buf.WriteString(k)
	if v != "" {
		buf.WriteRune('=')
		buf.WriteString(v)
	}
	return buf.String()
}

func engineWorkflowPhaseToNBMP(wp enginev1.WorkflowPhase) nbmpv2.State {
	switch wp {
	case enginev1.InitializingWorkflowPhase:
		return nbmpv2.InstantiatedState
	case enginev1.RunningWorkflowPhase, enginev1.AwaitingCompletionWorkflowPhase:
		return nbmpv2.RunningState
	case enginev1.SucceededWorkflowPhase:
		return nbmpv2.IdleState
	case enginev1.FailedWorkflowPhase:
		return nbmpv2.InErrorState
	default:
		return nbmpv2.InstantiatedState
	}
}

func engineTaskPhaseToNBMP(tp enginev1.TaskPhase) nbmpv2.State {
	switch tp {
	case enginev1.InitializingTaskPhase:
		return nbmpv2.InstantiatedState
	case enginev1.JobPendingTaskPhase:
		return nbmpv2.IdleState
	case enginev1.RunningTaskPhase:
		return nbmpv2.RunningState
	case enginev1.SucceededTaskPhase:
		return nbmpv2.IdleState
	case enginev1.FailedTaskPhase:
		return nbmpv2.InErrorState
	default:
		return nbmpv2.InstantiatedState
	}
}
