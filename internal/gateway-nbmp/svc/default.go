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

import nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"

func PopulateWorkflowDefaults(wf *nbmpv2.Workflow) {
	// Scheme

	if wf.Scheme == nil {
		wf.Scheme = &nbmpv2.Scheme{
			URI: nbmpv2.SchemaURI,
		}
	}

	// Acknowledge

	// reset Acknowledge for response
	wf.Acknowledge = &nbmpv2.Acknowledge{}
	if wf.Acknowledge.Unsupported == nil {
		wf.Acknowledge.Unsupported = make([]string, 0)
	}
	if wf.Acknowledge.Failed == nil {
		wf.Acknowledge.Failed = make([]string, 0)
	}
	if wf.Acknowledge.Partial == nil {
		wf.Acknowledge.Partial = make([]string, 0)
	}
}
