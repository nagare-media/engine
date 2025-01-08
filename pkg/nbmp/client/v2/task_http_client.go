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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/nagare-media/engine/internal/pkg/mime"
	"github.com/nagare-media/engine/pkg/nbmp"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

type httpTaskClient struct {
	Endpoint string
	HTTP     http.Client
}

var _ TaskClient = &httpTaskClient{}

func NewTaskClient(url string) *httpTaskClient {
	return &httpTaskClient{
		Endpoint: url,
		HTTP:     http.Client{},
	}
}

func (c *httpTaskClient) Create(ctx context.Context, tsk *nbmpv2.Task) (*nbmpv2.Task, error) {
	return c.request(ctx, http.MethodPost, c.Endpoint, tsk)
}

func (c *httpTaskClient) Update(ctx context.Context, tsk *nbmpv2.Task) (*nbmpv2.Task, error) {
	return c.request(ctx, http.MethodPatch, fmt.Sprintf("%s/%s", c.Endpoint, tsk.General.ID), tsk)
}

func (c *httpTaskClient) Delete(ctx context.Context, id string) (*nbmpv2.Task, error) {
	return c.request(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", c.Endpoint, id), nil)
}

func (c *httpTaskClient) Retrieve(ctx context.Context, id string) (*nbmpv2.Task, error) {
	return c.request(ctx, http.MethodGet, fmt.Sprintf("%s/%s", c.Endpoint, id), nil)
}

func (c *httpTaskClient) request(ctx context.Context, method, endpoint string, tsk *nbmpv2.Task) (*nbmpv2.Task, error) {
	buf := bytes.Buffer{}
	if tsk != nil {
		enc := json.NewEncoder(&buf)
		if err := enc.Encode(tsk); err != nil {
			return nil, err
		}
	}

	resp, err := c.doRequest(ctx, method, endpoint, &buf)
	if err != nil &&
		// note: For these errors, we still have a response body. We will still return err laster.
		err != nbmp.ErrInvalid &&
		err != nbmp.ErrAlreadyExists &&
		err != nbmp.ErrUnsupported {
		return nil, err
	}

	tsk = &nbmpv2.Task{}
	dec := json.NewDecoder(resp)
	if errDec := dec.Decode(tsk); errDec != nil {
		return nil, errDec
	}
	return tsk, err
}

func (c *httpTaskClient) doRequest(ctx context.Context, method, endpoint string, body io.Reader) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", mime.ApplicationMPEG_NBMP_TDD_JSON)
	req.Header.Set("Content-Type", mime.ApplicationMPEG_NBMP_TDD_JSON)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}

	err = handleStatusCode(resp.StatusCode)
	return resp.Body, err
}
