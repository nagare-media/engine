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
	nbmpconv "github.com/nagare-media/engine/internal/pkg/nbmpconv"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
	"k8s.io/utils/ptr"
)

type mediaParameterToMediaConverter struct {
	mp *nbmpv2.MediaParameter
}

var _ nbmpconv.ResetConverter[*nbmpv2.MediaParameter, *enginev1.Media] = &mediaParameterToMediaConverter{}

func NewMediaParameterToMediaConverter(mp *nbmpv2.MediaParameter) *mediaParameterToMediaConverter {
	c := &mediaParameterToMediaConverter{}
	c.Reset(mp)
	return c
}

func (c *mediaParameterToMediaConverter) Reset(mp *nbmpv2.MediaParameter) {
	c.mp = mp
}

func (c *mediaParameterToMediaConverter) Convert(m *enginev1.Media) error {
	m.Type = enginev1.MediaMediaType

	// $.stream-id
	m.ID = c.mp.StreamID

	// $.name
	if c.mp.Name != "" {
		m.HumanReadable = &enginev1.HumanReadableMediaDescription{
			Name: &c.mp.Name,
		}
	}

	// $.keywords
	for _, s := range c.mp.Keywords {
		k, v := parseKeyword(s)
		m.Labels[k] = v
	}

	// $.mime-type
	if c.mp.MimeType != "" {
		m.Metadata.MimeType = (*enginev1.MimeType)(&c.mp.MimeType)
	}

	// TODO: $.video-format
	// TODO: $.audio-format
	// TODO: $.image-format

	// $.codec-type
	if c.mp.CodecType != nil {
		m.Metadata.CodecType = (*enginev1.CodecType)(c.mp.CodecType)
	}

	// $.mode
	if c.mp.Mode != nil {
		switch *c.mp.Mode {
		case nbmpv2.PullMediaAccessMode:
			m.Direction = ptr.To(enginev1.PullMediaDirection)
		case nbmpv2.PushMediaAccessMode:
			m.Direction = ptr.To(enginev1.PushMediaDirection)
		}
	}

	// TODO: $.throughput
	// TODO: $.buffersize
	// TODO: $.availability-duration
	// TODO: $.timeout

	// $.caching-server-url
	// $.protocol: ignore (should be given with caching-server-url)
	if c.mp.CachingServerURL != "" {
		m.URL = ptr.To(c.mp.CachingServerURL)
	}

	// TODO: $.completion-timeout

	return nil
}
