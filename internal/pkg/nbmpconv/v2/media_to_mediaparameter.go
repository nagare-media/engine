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

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	nbmpconv "github.com/nagare-media/engine/internal/pkg/nbmpconv"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
	"k8s.io/utils/ptr"
)

type mediaToMediaParameterConverter struct {
	m *enginev1.Media
}

var _ nbmpconv.ResetConverter[*enginev1.Media, *nbmpv2.MediaParameter] = &mediaToMediaParameterConverter{}

func NewMediaToMediaParameterConverter(m *enginev1.Media) *mediaToMediaParameterConverter {
	c := &mediaToMediaParameterConverter{}
	c.Reset(m)
	return c
}

func (c *mediaToMediaParameterConverter) Reset(m *enginev1.Media) {
	c.m = m
}

func (c *mediaToMediaParameterConverter) Convert(mp *nbmpv2.MediaParameter) error {
	if c.m.Type != enginev1.MediaMediaType {
		return fmt.Errorf("convert: unexpected media type '%s'", c.m.Type)
	}

	// $.stream-id
	mp.StreamID = c.m.ID

	// $.name
	if c.m.HumanReadable != nil && c.m.HumanReadable.Name != nil {
		mp.Name = *c.m.HumanReadable.Name
	}

	// $.keywords
	for k, v := range c.m.Labels {
		mp.Keywords = append(mp.Keywords, encodeKeyword(k, v))
	}

	// $.mime-type
	if c.m.Metadata.MimeType != nil {
		mp.MimeType = string(*c.m.Metadata.MimeType)
	}

	// TODO: $.video-format
	// TODO: $.audio-format
	// TODO: $.image-format

	// $.codec-type
	if c.m.Metadata.CodecType != nil {
		mp.CodecType = (*string)(c.m.Metadata.CodecType)
	}

	// $.mode
	if c.m.Direction != nil {
		mp.Mode = ptr.To(MediaDirectionToMediaAccessMode(*c.m.Direction))
	}

	// TODO: $.throughput
	// TODO: $.buffersize
	// TODO: $.availability-duration
	// TODO: $.timeout

	// $.caching-server-url
	// $.protocol
	if c.m.URL != nil {
		mp.CachingServerURL = *c.m.URL
		u, err := c.m.URL.URL()
		if err != nil {
			return err
		}
		mp.Protocol = u.Scheme
	}

	// TODO: $.completion-timeout

	return nil
}
