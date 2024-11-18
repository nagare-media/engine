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

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// Media describes an input or output.
type Media struct {
	// ID of this media.
	ID string `json:"id"`

	// Human readable description of this media.
	// +optional
	HumanReadable *HumanReadableMediaDescription `json:"humanReadable,omitempty"`

	// Type of this media.
	Type MediaType `json:"type"`

	// Direction this media is streamed in.
	// +optional
	Direction *MediaDirection `json:"direction,omitempty"`

	PortBindings []MediaPortBinding `json:"portBindings"`

	// URL of this media.
	// +optional
	URL *string `json:"url,omitempty"`

	// Labels of this media.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Metadata of this media.
	Metadata MediaMetadata `json:"metadata"`
}

type HumanReadableMediaDescription struct {
	// Human readable name of this media.
	// +optional
	Name *string `json:"name,omitempty"`
}

// +kubebuilder:validation:Enum=media;metadata
type MediaType string

const (
	MediaMediaType    = MediaType("media")
	MetadataMediaType = MediaType("metadata")
)

// +kubebuilder:validation:Enum=push;pull
type MediaDirection string

const (
	PushMediaDirection = MediaDirection("push")
	PullMediaDirection = MediaDirection("pull")
)

type MediaPortBinding struct {
	// ID of the port.
	ID string `json:"id"`
}

type MediaMetadata struct {
	// +optional
	MimeType *MimeType `json:"mimeType,omitempty"`

	// +optional
	CodecType *CodecType `json:"codecType,omitempty"`

	// +optional
	Duration *metav1.Duration `json:"duration,omitempty"`

	// BitRate in bit/s
	// +optional
	BitRate *BitRate `json:"bitRate,omitempty"`

	// Size in bytes.
	// +optional
	Size *Size `json:"size,omitempty"`

	// +optional
	Checksums []Checksum `json:"checksums,omitempty"`

	// +optional
	Container *MediaContainer `json:"container,omitempty"`

	// +optional
	Streams []MediaStream `json:"streams,omitempty"`
}

type MimeType string

type CodecType string

type BitRate int64

type Size int64

type Checksum struct {
	Alg string `json:"alg"`
	Sum string `json:"sum"`
}

type MediaContainer struct {
	Name string `json:"name"`
}

type MediaStream struct {
	ID string `json:"id"`

	// +optional
	Codec *MediaCodec `json:"codec,omitempty"`

	// BitRate in bit/s
	// +optional
	BitRate *BitRate `json:"bitRate,omitempty"`

	// +optional
	Duration *metav1.Duration `json:"duration,omitempty"`

	// Properties additionally associated with this stream.
	// +optional
	Properties []MediaProperty `json:"properties,omitempty"`

	MediaStreamType `json:",inline"`
}

type MediaCodec struct {
	Name string `json:"name"`

	// +optional
	Profile *string `json:"profile,omitempty"`

	// +optional
	Level *string `json:"level,omitempty"`
}

type MediaProperty struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// +kubebuilder:validation:MinProperties=1
// +kubebuilder:validation:MaxProperties=1
type MediaStreamType struct {
	// +optional
	Audio *AudioMediaStream `json:"audio,omitempty"`

	// +optional
	Video *VideoMediaStream `json:"video,omitempty"`

	// +optional
	Subtitle *SubtitleMediaStream `json:"subtitle,omitempty"`

	// +optional
	Data *DataMediaStream `json:"data,omitempty"`
}

type AudioMediaStream struct {
	// +optional
	BitDepth *uint8 `json:"bitDepth,omitempty"`

	// +optional
	Channels *uint8 `json:"channels,omitempty"`

	// +optional
	ChannelLayout *string `json:"channelLayout,omitempty"`

	// +optional
	SamplingRate *int32 `json:"samplingRate,omitempty"`
}

type VideoMediaStream struct {
	// +optional
	BitDepth *uint8 `json:"bitDepth,omitempty"`

	// +optional
	Resolution *VideoResolution `json:"resolution,omitempty"`

	// +optional
	FrameRate *FrameRate `json:"frameRate,omitempty"`

	// +kubebuilder:default=unknown
	// +optional
	FieldOrder *FieldOrder `json:"fieldOrder,omitempty"`

	// +optional
	Color *VideoColor `json:"color,omitempty"`

	// TODO: add displayMatrix infos
}

type VideoResolution struct {
	Width int32 `json:"width"`

	Hight int32 `json:"hight"`

	// +kubebuilder:default="1.0"
	// +optional
	SAR *string `json:"sar,omitempty"`
}

type FrameRate struct {
	// +optional
	Average *string `json:"average,omitempty"`

	// +optional
	LowestCommon *string `json:"lowestCommon,omitempty"`
}

// +kubebuilder:validation:Enum=unknown;progressive;top-field-first;bottom-field-first
type FieldOrder string

const (
	UnknownFieldOrder          = FieldOrder("unknown")
	ProgressiveFieldOrder      = FieldOrder("progressive")
	TopFieldFirstFieldOrder    = FieldOrder("top-field-first")
	BottomFieldFirstFieldOrder = FieldOrder("bottom-field-first")
)

type VideoColor struct {
	// +optional
	Model *string `json:"model,omitempty"`

	// +optional
	ChromaSubsampling *ChromaSubsampling `json:"chromaSubsampling,omitempty"`

	// +kubebuilder:default=limited
	// +optional
	Range *ColorRange `json:"range,omitempty"`

	// +optional
	Space *string `json:"space,omitempty"`

	// +optional
	Primaries *string `json:"primaries,omitempty"`

	// +optional
	Transfer *string `json:"transfer,omitempty"`
}

// +kubebuilder:validation:Enum={"3:1:1","4:1:0","4:1:1","4:2:0","4:2:2","4:4:0","4:4:4"}
type ChromaSubsampling string

const (
	ChromaSubsampling311 = ChromaSubsampling("3:1:1")
	ChromaSubsampling410 = ChromaSubsampling("4:1:0")
	ChromaSubsampling411 = ChromaSubsampling("4:1:1")
	ChromaSubsampling420 = ChromaSubsampling("4:2:0")
	ChromaSubsampling422 = ChromaSubsampling("4:2:2")
	ChromaSubsampling440 = ChromaSubsampling("4:4:0")
	ChromaSubsampling444 = ChromaSubsampling("4:4:4")
)

// +kubebuilder:validation:Enum=full;limited
type ColorRange string

const (
	FullColorRange    = ColorRange("full")
	LimitedColorRange = ColorRange("limited")
)

type SubtitleMediaStream struct {
	// +optional
	Language *Language `json:"language,omitempty"`
}

type Language string

type DataMediaStream struct {
}
