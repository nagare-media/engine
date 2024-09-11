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

package events

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cenkalti/backoff/v4"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/nagare-media/engine/internal/pkg/mime"
)

type Client interface {
	Send(context.Context, cloudevents.Event) error
	SendAsyncAck(context.Context, cloudevents.Event) error
}

// DiscardClient will discard all events.
type DiscardClient struct{}

var _ Client = &DiscardClient{}

func (c *DiscardClient) Send(context.Context, cloudevents.Event) error {
	return nil
}

func (c *DiscardClient) SendAsyncAck(context.Context, cloudevents.Event) error {
	return nil
}

// HTTPClient will deliver messages using HTTP.
type HTTPClient struct {
	URL    string
	Client *http.Client
}

var _ Client = &HTTPClient{}

func (c *HTTPClient) Send(ctx context.Context, e cloudevents.Event) error {
	return c.doSend(ctx, e, false)
}

func (c *HTTPClient) SendAsyncAck(ctx context.Context, e cloudevents.Event) error {
	return c.doSend(ctx, e, true)
}

func (c *HTTPClient) doSend(ctx context.Context, e cloudevents.Event, async bool) error {
	buf := &bytes.Buffer{}
	enc := json.NewEncoder(buf)
	if err := enc.Encode(e); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.URL, buf)
	if err != nil {
		return err
	}

	if async {
		q := req.URL.Query()
		q.Add("async", "true")
		req.URL.RawQuery = q.Encode()
	}

	req.Header.Set("Content-Type", mime.ApplicationCloudEventsJSON)

	if c.Client == nil {
		c.Client = http.DefaultClient
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode > 299 {
		return fmt.Errorf("event: unexpected HTTP status code in response: %d", resp.StatusCode)
	}

	return nil
}

type backoffClient struct {
	c Client
	b backoff.BackOff
}

var _ Client = &backoffClient{}

func ClientWithBackoff(c Client, b backoff.BackOff) *backoffClient {
	return &backoffClient{
		c: c,
		b: b,
	}
}

func (c *backoffClient) Send(ctx context.Context, e cloudevents.Event) error {
	c.b.Reset()
	op := func() error { return c.c.Send(ctx, e) }
	return backoff.Retry(op, backoff.WithContext(c.b, ctx))
}

func (c *backoffClient) SendAsyncAck(ctx context.Context, e cloudevents.Event) error {
	c.b.Reset()
	op := func() error { return c.c.SendAsyncAck(ctx, e) }
	return backoff.Retry(op, backoff.WithContext(c.b, ctx))
}
