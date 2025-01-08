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
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

func FindMediaOrMetadataParameter(p nbmpv2.Port, io nbmpv2.InputOrOutput) nbmpv2.MediaOrMetadataParameter {
	if p.Bind == nil {
		return nil
	}

	for _, mp := range io.GetMediaParameters() {
		if mp.StreamID == p.Bind.StreamID {
			return &mp
		}
	}

	for _, mp := range io.GetMetadataParameters() {
		if mp.StreamID == p.Bind.StreamID {
			return &mp
		}
	}

	return nil
}
