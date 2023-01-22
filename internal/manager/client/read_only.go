/*
Copyright 2022-2023 The nagare media authors

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

package client

import (
	"context"
	"errors"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ client.Client = &readOnlyClient{}

type readOnlyClient struct {
	client.Reader

	scheme     *runtime.Scheme
	restMapper meta.RESTMapper
}

func NewReadOnlyClient(reader client.Reader, scheme *runtime.Scheme, restMapper meta.RESTMapper) client.Client {
	return &readOnlyClient{
		Reader:     reader,
		scheme:     scheme,
		restMapper: restMapper,
	}
}

func (c *readOnlyClient) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	return errors.New("read only")
}

func (c *readOnlyClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	return errors.New("read only")
}

func (c *readOnlyClient) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	return errors.New("read only")
}

func (c *readOnlyClient) Patch(ctx context.Context, obj client.Object, patch client.Patch, opts ...client.PatchOption) error {
	return errors.New("read only")
}

func (c *readOnlyClient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return errors.New("read only")
}

func (c *readOnlyClient) Status() client.SubResourceWriter {
	return nil
}

func (c *readOnlyClient) SubResource(subResource string) client.SubResourceClient {
	return nil
}

func (c *readOnlyClient) Scheme() *runtime.Scheme {
	return c.scheme
}

func (c *readOnlyClient) RESTMapper() meta.RESTMapper {
	return c.restMapper
}
