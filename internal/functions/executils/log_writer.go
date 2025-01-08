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

package executils

import (
	"bytes"
	"context"
	"io"
	"strings"
	"unicode"
	"unsafe"

	"github.com/go-logr/logr"
	"github.com/smallnest/ringbuffer"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	MaxLogLineLength = 2048
)

type LogFunc func(string)

func ContextLogFunc(ctx context.Context) LogFunc {
	l := log.FromContext(ctx)
	return LoggerLogFunc(l)
}

func LoggerLogFunc(l logr.Logger) LogFunc {
	return func(msg string) {
		l.Info(msg)
	}
}

type logWriter struct {
	log            LogFunc
	bufferLastLine bool
	buf            ringbuffer.RingBuffer
	sbuf           bytes.Buffer
}

var _ io.WriteCloser = &logWriter{}

func (lw *logWriter) Write(p []byte) (n int, err error) {

	// log line by line
	for {
		i := bytes.IndexByte(p, '\n')
		if i < 0 {
			break
		}

		line := dropCR(p[0:i])
		bufferedLine := lw.flushBuffer()

		lw.sbuf.Reset()
		lw.sbuf.Grow(len(bufferedLine) + len(line))

		// first write buffered line
		_, err = lw.sbuf.Write(bufferedLine)
		if err != nil {
			return
		}

		// next write new line
		_, err = lw.sbuf.Write(line)
		if err != nil {
			return
		}

		// log complete line
		lw.logBytes(lw.sbuf.Bytes())

		n += i + 1
		p = p[i+1:]
	}

	// should we buffer last line?
	if !lw.bufferLastLine {
		n += len(p)
		p := dropCR(p)
		lw.logBytes(p)
		return
	}

	// buffer remaining bytes for next write
	if len(p) > 0 {
		var n2 int
		for {
			n2, err = lw.buf.TryWrite(p)
			n += n2

			if err != ringbuffer.ErrTooMuchDataToWrite {
				return
			}

			// if buffer is full, we force a log output
			err = lw.Flush()
			if err != nil {
				return
			}

			p = p[n2:]
		}
	}

	return
}

func (lw *logWriter) logBytes(p []byte) {
	line := unsafe.String(unsafe.SliceData(p), len(p))
	line = strings.TrimRightFunc(line, unicode.IsSpace)
	if len(line) > 0 {
		lw.log(line)
	}
}

func (lw *logWriter) Close() error {
	return lw.Flush()
}

func (lw *logWriter) Flush() error {
	b := lw.flushBuffer()
	if len(b) > 0 {
		lw.logBytes(b)
	}
	return nil
}

func (lw *logWriter) flushBuffer() []byte {
	b := lw.buf.Bytes(nil)
	lw.buf.Reset()
	return b
}

func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

func NewLogWriter(l LogFunc, bufferLastLine bool) io.WriteCloser {
	return &logWriter{
		log:            l,
		bufferLastLine: bufferLastLine,
		buf: *ringbuffer.
			New(MaxLogLineLength).
			SetBlocking(true),
	}
}
