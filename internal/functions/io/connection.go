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

import "io"

type Connector interface {
	Connect(Connection)
}

type Connection interface {
	io.WriteCloser
	io.ReaderFrom
}

type connection struct {
	w io.Writer
}

var _ Connection = &connection{}

func NewConnection(w io.Writer) Connection {
	return &connection{w: w}
}

func (c *connection) Write(p []byte) (n int, err error) {
	return c.w.Write(p)
}

func (c *connection) ReadFrom(r io.Reader) (n int64, err error) {
	// fast path
	if rf, ok := c.w.(io.ReaderFrom); ok {
		return rf.ReadFrom(r)
	}

	// slow copy path
	buf := *AcquireBuffer()
	defer ReleaseBuffer(&buf)
	return io.CopyBuffer(c.w, r, buf)
}

func (c *connection) Close() error {
	return nil
}
