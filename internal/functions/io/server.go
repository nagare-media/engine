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

	"github.com/nagare-media/engine/pkg/registry"
	"github.com/nagare-media/engine/pkg/starter"
)

var (
	ApplicationProtocolUnknown = errors.New("io: application protocol unknown")
)

type Server interface {
	starter.Starter

	Protocol() string
}

// TODO: implement a builder solution that supports configuration
type ServerBuilder func(portNumber uint16) (Server, error)

var ServerBuilders = registry.New[ServerBuilder]()

func NewServerFor(p Port) (Server, error) {
	builder, ok := ServerBuilders.Get(p.Protocol())
	if !ok {
		return nil, ApplicationProtocolUnknown
	}
	return builder(p.PortNumber())
}
