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
	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

func MediaAccessModeToMediaDirection(ma nbmpv2.MediaAccessMode) enginev1.MediaDirection {
	switch ma {
	case nbmpv2.PullMediaAccessMode:
		return enginev1.PullMediaDirection
	case nbmpv2.PushMediaAccessMode:
		return enginev1.PushMediaDirection
	}
	return ""
}

func MediaDirectionToMediaAccessMode(md enginev1.MediaDirection) nbmpv2.MediaAccessMode {
	switch md {
	case enginev1.PullMediaDirection:
		return nbmpv2.PullMediaAccessMode
	case enginev1.PushMediaDirection:
		return nbmpv2.PushMediaAccessMode
	}
	return ""
}
