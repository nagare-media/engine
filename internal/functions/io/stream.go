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

	"github.com/nagare-media/engine/pkg/starter"
)

type Stream interface {
	Connection
	Connector
}

type stream struct {
	con []Connection
}

var _ Stream = &stream{}

func NewStream() Stream {
	return &stream{}
}

func (s *stream) Connect(c Connection) {
	s.con = append(s.con, c)
}

func (s *stream) Write(p []byte) (n int, err error) {
	for _, c := range s.con {
		n, err = c.Write(p)
		if err != nil {
			return
		}
		if n != len(p) {
			err = io.ErrShortWrite
			return
		}
	}
	return len(p), nil
}

func (s *stream) ReadFrom(r io.Reader) (written int64, err error) {
	// see [io.Copy] implementation

	// fast paths
	if len(s.con) == 0 {
		// [io.Discard] implements [io.ReaderFrom]
		return io.Copy(io.Discard, r)
	}
	if len(s.con) == 1 {
		if rf, ok := s.con[0].(io.ReaderFrom); ok {
			return rf.ReadFrom(r)
		}
	}

	// slow copy path
	buf := *AcquireBuffer()
	defer ReleaseBuffer(&buf)
	for {
		nr, er := r.Read(buf)
		if nr > 0 {
			nw, ew := s.Write(buf[0:nr])
			if nw < 0 || nw > nr {
				nw = 0
				if ew == nil {
					ew = errors.New("invalid write result")
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}

	// we don't automatically close after EOF (i.e. err == nil)

	return written, err
}

func (s *stream) Close() (err error) {
	for _, c := range s.con {
		ec := c.Close()
		if ec != nil && err == nil {
			// TODO: should we return a composite error?
			err = ec
		}
	}
	return
}

type StreamProcessor interface {
	starter.Starter
}

type StreamProcessorFunc func(context.Context) error

func (f StreamProcessorFunc) Start(ctx context.Context) error {
	return f(ctx)
}

var _ StreamProcessor = StreamProcessorFunc(func(context.Context) error { return nil })
