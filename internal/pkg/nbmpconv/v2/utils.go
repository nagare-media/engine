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

package v2

import (
	"fmt"
	"strings"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	nbmputils "github.com/nagare-media/engine/pkg/nbmp/utils"
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

func EngineWorkflowPhaseToNBMP(wp enginev1.WorkflowPhase) nbmpv2.State {
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

func EngineTaskPhaseToNBMP(tp enginev1.TaskPhase) nbmpv2.State {
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

func ParametersToMap(p []nbmpv2.Parameter) (map[string]string, error) {
	m := make(map[string]string, len(p))
	for _, e := range p {
		v, ok := nbmputils.ExtractStringParameterValue(e)
		if !ok {
			return nil, fmt.Errorf("convert: unsupported parameter value for '%s'", e.Name)
		}
		m[e.Name] = v
	}
	return m, nil
}

func MapToParameters(m map[string]string) []nbmpv2.Parameter {
	p := make([]nbmpv2.Parameter, 0, len(m))
	for k, v := range m {
		p = nbmputils.SetStringParameterValue(p, k, v)
	}
	return p
}
