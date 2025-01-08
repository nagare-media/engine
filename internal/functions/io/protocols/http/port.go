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

package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	"sigs.k8s.io/controller-runtime/pkg/log"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	engineio "github.com/nagare-media/engine/internal/functions/io"
	"github.com/nagare-media/engine/internal/pkg/backoff"
	nbmpconvv2 "github.com/nagare-media/engine/internal/pkg/nbmpconv/v2"
	"github.com/nagare-media/engine/pkg/starter"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

type port struct {
	mtx sync.Mutex

	pt     engineio.PortType
	d      enginev1.MediaDirection
	name   string
	number uint16
	url    *url.URL

	client *fasthttp.Client
	con    engineio.Connection
	pr     *io.PipeReader
	pw     *io.PipeWriter
}

var _ engineio.Port = &port{}
var _ engineio.InputPort = &port{}
var _ engineio.OutputPort = &port{}

func (p *port) Start(ctx context.Context) error {
	switch {
	case p.pt == engineio.InputPortType && p.d == enginev1.PushMediaDirection:
		//                  input (push)
		// ┌────────────┐           ┌───┐          ┌─────┐
		// │ handlePost │──Server──▶│ p │──Start──▶│ con │
		// └────────────┘           └───┘          └─────┘
		defer p.con.Close()
		_, err := io.Copy(p.con, p.pr)
		if err != nil {
			return err
		}

	case p.pt == engineio.InputPortType && p.d == enginev1.PullMediaDirection:
		//                  input (pull)
		// ┌────────────┐          ┌───┐          ┌─────┐
		// │ getRequest │──Start──▶│ p │──Start──▶│ con │
		// └────────────┘          └───┘          └─────┘
		defer p.con.Close()
		send := starter.Func(func(ctx context.Context) error {
			_, err := p.con.ReadFrom(p.pr)
			return err
		})
		get := starter.Func(p.getRequest)
		return starter.Manage(send, get).WaitForAllToTerminate().Start(ctx)

	case p.pt == engineio.OutputPortType && p.d == enginev1.PushMediaDirection:
		//                  output (push)
		// ┌─────┐           ┌───┐          ┌─────────────┐
		// │ str │──Stream──▶│ p │──Start──▶│ postRequest │
		// └─────┘           └───┘          └─────────────┘
		return p.postRequest(ctx)

	case p.pt == engineio.OutputPortType && p.d == enginev1.PullMediaDirection:
		//                  output (pull)
		// ┌─────┐           ┌───┐           ┌───────────┐
		// │ str │──Stream──▶│ p │──Server──▶│ handleGet │
		// └─────┘           └───┘           └───────────┘
		// nothing to do here
	}

	return nil
}

func (p *port) Connect(c engineio.Connection) {
	p.con = c
}

func (p *port) Write(b []byte) (n int, err error) {
	return p.pw.Write(b)
}

func (p *port) ReadFrom(r io.Reader) (n int64, err error) {
	n, err = io.Copy(p.pw, r)
	// we don't automatically close after EOF (i.e. err == nil)
	return
}

func (p *port) Read(b []byte) (int, error) {
	return p.pr.Read(b)
}

func (p *port) Close() error {
	return p.pw.Close()
}

func (p *port) Name() string {
	return p.name
}

func (p *port) BaseName() string {
	return engineio.PortCommonBaseName(p.Name())
}

func (p *port) PortNumber() uint16 {
	return p.number
}

func (p *port) Protocol() string {
	return Protocol
}

func (p *port) Type() engineio.PortType {
	return p.pt
}

func (p *port) Direction() enginev1.MediaDirection {
	return p.d
}

func (p *port) MountTo(srv engineio.Server) error {
	s, ok := srv.(Server)
	if !ok {
		return engineio.ServerIncompatibleWithPort
	}

	switch {
	case p.pt == engineio.InputPortType && p.d == enginev1.PushMediaDirection:
		s.Router().Post(p.url.EscapedPath(), p.handlePost)

	case p.pt == engineio.OutputPortType && p.d == enginev1.PullMediaDirection:
		s.Router().Get(p.url.EscapedPath(), p.handleGet)
	}

	return nil
}

func (p *port) getRequest(ctx context.Context) error {
	// input (pull)
	l := log.FromContext(ctx)

	op := func() error {
		req := fasthttp.AcquireRequest()
		defer fasthttp.ReleaseRequest(req)
		req.Header.SetMethod(fasthttp.MethodGet)
		req.SetRequestURI(p.url.String())
		req.SetConnectionClose()

		resp := fasthttp.AcquireResponse()
		resp.StreamBody = true
		defer fasthttp.ReleaseResponse(resp)

		err := p.client.Do(req, resp)
		if err != nil && !errors.Is(err, fasthttp.ErrConnectionClosed) {
			return err
		}

		if resp.StatusCode() != fasthttp.StatusOK {
			return fmt.Errorf("event: unexpected HTTP status code in response: %d", resp.StatusCode())
		}

		n, err := p.ReadFrom(resp.BodyStream())
		// TODO: deal with "n > 0"
		_ = n
		if err != nil {
			return err
		}

		// TODO: how do we know, that no additional requests are coming?
		p.Close()

		return nil
	}

	no := func(err error, t time.Duration) {
		l.Error(err, fmt.Sprintf("failed; retrying after %s", t))
	}

	return backoff.RetryFunc(ctx, op, backoff.WithNotify(no))
}

func (p *port) postRequest(ctx context.Context) error {
	// output (push)
	l := log.FromContext(ctx)

	op := func() error {
		req := fasthttp.AcquireRequest()
		defer fasthttp.ReleaseRequest(req)
		req.Header.SetMethod(fasthttp.MethodPost)
		req.SetBodyStream(io.NopCloser(p), -1)
		req.SetRequestURI(p.url.String())
		req.SetConnectionClose()

		resp := fasthttp.AcquireResponse()
		defer func() { _ = resp.CloseBodyStream() }()
		defer fasthttp.ReleaseResponse(resp)

		err := p.client.Do(req, resp)
		if err != nil && !errors.Is(err, fasthttp.ErrConnectionClosed) {
			return err
		}

		if resp.StatusCode() > 299 {
			return fmt.Errorf("event: unexpected HTTP status code in response: %d", resp.StatusCode())
		}

		return nil
	}

	no := func(err error, t time.Duration) {
		l.Error(err, fmt.Sprintf("failed; retrying after %s", t))
	}

	return backoff.RetryFunc(ctx, op, backoff.WithNotify(no))
}

func (p *port) handleGet(c *fiber.Ctx) error {
	// output (pull)

	// only one request at a time
	p.mtx.Lock()
	defer p.mtx.Unlock()

	// TODO: allow other content-types
	c.Response().Header.SetContentType("video/mp2t")
	c.Response().Header.Set(fiber.HeaderXContentTypeOptions, "nosniff")

	// flush headers for chunked transfer encoding
	c.Response().ImmediateHeaderFlush = true

	// [Close] is called by fiber
	c.Response().SetBodyStream(io.NopCloser(p), -1)

	c.Status(fiber.StatusOK)
	return nil
}

func (p *port) handlePost(c *fiber.Ctx) error {
	// input (push)

	// only one request at a time
	p.mtx.Lock()
	defer p.mtx.Unlock()

	var err error
	if c.Request().IsBodyStream() {
		defer func() {
			_ = c.Request().CloseBodyStream()
		}()
		_, err = p.ReadFrom(c.Request().BodyStream())
	} else {
		_, err = p.Write(c.Body())
	}

	if err != nil {
		return fiber.ErrInternalServerError
	}

	// TODO: how do we know, that no additional requests are coming?
	p.Close()

	c.Status(fiber.StatusCreated)
	return nil
}

func NewInputPortFor(p nbmpv2.Port, mp nbmpv2.MediaOrMetadataParameter) (engineio.InputPort, error) {
	in, err := newPortFor(p, mp)
	if err != nil {
		return nil, err
	}
	in.pt = engineio.InputPortType
	return in, nil
}

func NewOutputPortFor(p nbmpv2.Port, mp nbmpv2.MediaOrMetadataParameter) (engineio.OutputPort, error) {
	out, err := newPortFor(p, mp)
	if err != nil {
		return nil, err
	}
	out.pt = engineio.OutputPortType
	return out, nil
}

func newPortFor(p nbmpv2.Port, mp nbmpv2.MediaOrMetadataParameter) (*port, error) {
	// protocol
	if Protocol != mp.GetProtocol() {
		return nil, engineio.ProtocolIncompatibleWithPort
	}

	// direction
	if mp.GetMode() == nil {
		return nil, engineio.PortDirectionUnknown
	}
	d := nbmpconvv2.MediaAccessModeToMediaDirection(*mp.GetMode())
	if d == "" {
		return nil, engineio.PortDirectionUnknown
	}

	// URL
	if mp.GetCachingServerURL() == nil {
		return nil, engineio.PortURLUnknown
	}
	u, err := mp.GetCachingServerURL().URL()
	if err != nil {
		return nil, err
	}

	// port number
	// TODO: port number should not be derived from caching-server-url (= Service port) but from function configuration (i.e. the Job template)
	pn := DefaultPortNumber
	pnstr := u.Port()
	if pnstr != "" {
		pnp, err := strconv.ParseUint(pnstr, 10, 16)
		if err != nil {
			return nil, err
		}
		pn = uint16(pnp)
	}

	pr, pw := engineio.Pipe()
	port := &port{
		name:   p.PortName,
		d:      d,
		url:    u,
		number: pn,
		client: &fasthttp.Client{},
		pr:     pr,
		pw:     pw,
	}

	return port, nil
}

func init() {
	engineio.InputPortBuilders.Register(Protocol, NewInputPortFor)
	engineio.OutputPortBuilders.Register(Protocol, NewOutputPortFor)
}
