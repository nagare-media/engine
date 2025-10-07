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
	"context"
	"errors"
	"io"

	"github.com/smallnest/ringbuffer"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	engineio "github.com/nagare-media/engine/internal/functions/io"
	"github.com/nagare-media/engine/pkg/engineurl"
	"github.com/nagare-media/engine/pkg/starter"
	"github.com/nagare-media/models.go/base"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

const (
	Protocol = "buffered"

	DefaultBufferSize = 10 * 1024 * 1024 // 10MB
)

type port struct {
	p   engineio.Port
	con engineio.Connection
	buf *ringbuffer.RingBuffer
}

var _ engineio.InputPort = &port{}
var _ engineio.OutputPort = &port{}

func (p *port) Start(ctx context.Context) error {
	mgr := starter.Manage(p.p, starter.Func(p.startBuffer)).WaitForAllToTerminate()
	return mgr.Start(ctx)
}

func (p *port) startBuffer(ctx context.Context) (err error) {
	switch p.p.Type() {
	case engineio.InputPortType:
		defer p.con.Close()
		_, err = p.con.ReadFrom(p.buf)

	case engineio.OutputPortType:
		defer p.p.Close()
		_, err = p.p.ReadFrom(p.buf)

	default:
		err = engineio.PortTypeUnknown
	}

	return
}

func (p *port) Connect(c engineio.Connection) {
	p.con = c
}

func (p *port) Write(b []byte) (n int, err error) {
	return p.buf.Write(b)
}

func (p *port) ReadFrom(r io.Reader) (n int64, err error) {
	n, err = p.buf.ReadFrom(r)
	// we don't automatically close after EOF (i.e. err == nil)
	return
}

func (p *port) Close() error {
	p.buf.CloseWriter()
	return nil
}

func (p *port) Name() string {
	return p.p.Name()
}

func (p *port) BaseName() string {
	return p.p.BaseName()
}

func (p *port) Protocol() string {
	return p.p.Protocol()
}

func (p *port) PortNumber() uint16 {
	return p.p.PortNumber()
}

func (p *port) Type() engineio.PortType {
	return p.p.Type()
}

func (p *port) Direction() enginev1.MediaDirection {
	return p.p.Direction()
}

func (p *port) MountTo(srv engineio.Server) error {
	return p.p.MountTo(srv)
}

func NewInputPortFor(p nbmpv2.Port, mp nbmpv2.MediaOrMetadataParameter) (engineio.InputPort, error) {
	sp, err := newInputSubPortFor(p, mp)
	if err != nil {
		return nil, err
	}
	// TODO: implement buffer size query argument
	return NewInputPortForPort(sp)
}

func NewInputPortForPort(p engineio.InputPort) (engineio.InputPort, error) {
	return NewInputPortForPortSize(p, DefaultBufferSize)
}

func NewInputPortForPortSize(p engineio.InputPort, size int) (engineio.InputPort, error) {
	port := newPortForPortSize(p, size)
	p.Connect(port)
	return port, nil
}

func newInputSubPortFor(p nbmpv2.Port, mp nbmpv2.MediaOrMetadataParameter) (engineio.InputPort, error) {
	mp, err := rewriteMediaOrMetadataParameter(mp)
	if err != nil {
		return nil, err
	}
	return engineio.NewInputPortFor(p, mp)
}

func NewOutputPortFor(p nbmpv2.Port, mp nbmpv2.MediaOrMetadataParameter) (engineio.OutputPort, error) {
	sp, err := newOutputSubPortFor(p, mp)
	if err != nil {
		return nil, err
	}
	// TODO: implement buffer size query argument
	return NewOutputPortForPort(sp)
}

func NewOutputPortForPort(p engineio.OutputPort) (engineio.OutputPort, error) {
	return NewOutputPortForPortSize(p, DefaultBufferSize)
}

func NewOutputPortForPortSize(p engineio.OutputPort, size int) (engineio.OutputPort, error) {
	return newPortForPortSize(p, size), nil
}

func newOutputSubPortFor(p nbmpv2.Port, mp nbmpv2.MediaOrMetadataParameter) (engineio.OutputPort, error) {
	mp, err := rewriteMediaOrMetadataParameter(mp)
	if err != nil {
		return nil, err
	}
	return engineio.NewOutputPortFor(p, mp)
}

func newPortForPortSize(p engineio.OutputPort, size int) *port {
	return &port{
		p:   p,
		buf: ringbuffer.New(size).SetBlocking(true),
	}
}

func rewriteMediaOrMetadataParameter(mp nbmpv2.MediaOrMetadataParameter) (nbmpv2.MediaOrMetadataParameter, error) {
	if mp.GetCachingServerURL() == nil {
		return nil, engineio.PortURLUnknown
	}

	u, err := mp.GetCachingServerURL().URL()
	if err != nil {
		return nil, err
	}

	q := u.Query()
	protocol := q.Get(engineurl.BufferedProtocolQueryKey)
	if protocol == "" {
		return nil, errors.New("buffered: missing protocol query key")
	}
	q.Del(engineurl.BufferedProtocolQueryKey)

	u.Scheme = protocol
	u.RawQuery = q.Encode()

	u2 := base.URI(u.String())
	mp.SetCachingServerURL(&u2)
	mp.SetProtocol(protocol)
	return mp, nil
}

func init() {
	engineio.InputPortBuilders.Register(Protocol, NewInputPortFor)
	engineio.OutputPortBuilders.Register(Protocol, NewOutputPortFor)
}
