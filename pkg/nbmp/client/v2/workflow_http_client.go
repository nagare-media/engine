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

type httpWorkflowClient struct {
	Endpoint string
	HTTP     http.Client
}

var _ WorkflowClient = &httpWorkflowClient{}

func NewWorkflowClient(url string) *httpWorkflowClient {
	return &httpWorkflowClient{
		Endpoint: url,
		HTTP:     http.Client{},
	}
}

func (c *httpWorkflowClient) Create(ctx context.Context, wf *nbmpv2.Workflow) (*nbmpv2.Workflow, error) {
	return c.request(ctx, http.MethodPost, c.Endpoint, wf)
}

func (c *httpWorkflowClient) Update(ctx context.Context, wf *nbmpv2.Workflow) (*nbmpv2.Workflow, error) {
	return c.request(ctx, http.MethodPatch, fmt.Sprintf("%s/%s", c.Endpoint, wf.General.ID), wf)
}

func (c *httpWorkflowClient) Delete(ctx context.Context, id string) (*nbmpv2.Workflow, error) {
	return c.request(ctx, http.MethodDelete, fmt.Sprintf("%s/%s", c.Endpoint, id), nil)
}

func (c *httpWorkflowClient) Retrieve(ctx context.Context, id string) (*nbmpv2.Workflow, error) {
	return c.request(ctx, http.MethodGet, fmt.Sprintf("%s/%s", c.Endpoint, id), nil)
}

func (c *httpWorkflowClient) request(ctx context.Context, method, endpoint string, wf *nbmpv2.Workflow) (*nbmpv2.Workflow, error) {
	buf := bytes.Buffer{}
	if wf != nil {
		enc := json.NewEncoder(&buf)
		if err := enc.Encode(wf); err != nil {
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

	wf = &nbmpv2.Workflow{}
	dec := json.NewDecoder(resp)
	if errDec := dec.Decode(wf); errDec != nil {
		return nil, errDec
	}
	return wf, err
}

func (c *httpWorkflowClient) doRequest(ctx context.Context, method, endpoint string, body io.Reader) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", mime.ApplicationMPEG_NBMP_WDD_JSON)
	req.Header.Set("Content-Type", mime.ApplicationMPEG_NBMP_WDD_JSON)

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}

	err = handleStatusCode(resp.StatusCode)
	return resp.Body, err
}
