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

package io

import (
	"errors"
	"strconv"
	"strings"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	nbmputils "github.com/nagare-media/engine/pkg/nbmp/utils"
	"github.com/nagare-media/engine/pkg/registry"
	"github.com/nagare-media/engine/pkg/starter"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

const (
	PortInstanceSeparationChar = "."
)

type PortType int

const (
	InputPortType PortType = iota
	OutputPortType
)

var (
	PortMissing                  = errors.New("io: port missing")
	PortNameUnknown              = errors.New("io: port name unknown")
	PortTypeUnknown              = errors.New("io: port type unknown")
	PortDirectionUnknown         = errors.New("io: port direction unknown")
	PortURLUnknown               = errors.New("io: port URL unknown")
	PortStreamNotFound           = errors.New("io: port stream not found")
	PortUnbound                  = errors.New("io: port unbound")
	ServerIncompatibleWithPort   = errors.New("io: server incompatible with port")
	ProtocolIncompatibleWithPort = errors.New("io: protocol incompatible with port")
)

type Port interface {
	starter.Starter
	Connection

	Name() string
	BaseName() string
	Protocol() string
	PortNumber() uint16
	Type() PortType
	Direction() enginev1.MediaDirection
	MountTo(Server) error
}

type InputPort interface {
	Port
	Connector
}

type OutputPort interface {
	Port
}

type InputPortBuilder func(nbmpv2.Port, nbmpv2.MediaOrMetadataParameter) (InputPort, error)

var InputPortBuilders = registry.New[InputPortBuilder]()

func NewInputPortForInputs(p nbmpv2.Port, i *nbmpv2.Input) (InputPort, error) {
	mp := nbmputils.FindMediaOrMetadataParameter(p, i)
	if mp == nil {
		if p.Bind == nil {
			return nil, PortUnbound
		}
		return nil, PortStreamNotFound
	}
	return NewInputPortFor(p, mp)
}

func NewInputPortFor(p nbmpv2.Port, mp nbmpv2.MediaOrMetadataParameter) (InputPort, error) {
	builder, ok := InputPortBuilders.Get(mp.GetProtocol())
	if !ok {
		return nil, ApplicationProtocolUnknown
	}
	return builder(p, mp)
}

type OutputPortBuilder func(nbmpv2.Port, nbmpv2.MediaOrMetadataParameter) (OutputPort, error)

var OutputPortBuilders = registry.New[OutputPortBuilder]()

func NewOutputPortForOutputs(p nbmpv2.Port, o *nbmpv2.Output) (OutputPort, error) {
	mp := nbmputils.FindMediaOrMetadataParameter(p, o)
	if mp == nil {
		if p.Bind == nil {
			return nil, PortUnbound
		}
		return nil, PortStreamNotFound
	}
	return NewOutputPortFor(p, mp)
}

func NewOutputPortFor(p nbmpv2.Port, mp nbmpv2.MediaOrMetadataParameter) (OutputPort, error) {
	builder, ok := OutputPortBuilders.Get(mp.GetProtocol())
	if !ok {
		return nil, ApplicationProtocolUnknown
	}
	return builder(p, mp)
}

func PortCommonBaseName(n string) string {
	i := strings.LastIndex(n, PortInstanceSeparationChar)
	if i < 0 {
		return n
	}
	inst := parsePortInstance(n[i+1:])
	if inst < 0 {
		return n
	}
	return n[:i]
}

func PortInstance(n string) int64 {
	i := strings.LastIndex(n, PortInstanceSeparationChar)
	if i < 0 {
		return -1
	}
	return parsePortInstance(n[i+1:])
}

func parsePortInstance(i string) int64 {
	inst, err := strconv.ParseInt(i, 10, 0)
	if err != nil {
		return -1
	}
	if inst < 0 {
		return -1
	}
	return inst
}

func IsServerPort(p Port) bool {
	return (p.Type() == InputPortType && p.Direction() == enginev1.PushMediaDirection) ||
		(p.Type() == OutputPortType && p.Direction() == enginev1.PullMediaDirection)
}

func IsClientPort(p Port) bool {
	return !IsServerPort(p)
}
