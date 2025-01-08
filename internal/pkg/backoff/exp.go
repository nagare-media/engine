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

package backoff

import (
	"context"
	"math"
	"time"

	"github.com/cenkalti/backoff/v5"
)

type Func func() error

type BackOff = backoff.BackOff

var (
	WithBackOff        = backoff.WithBackOff
	WithNotify         = backoff.WithNotify
	WithMaxTries       = backoff.WithMaxTries
	WithMaxElapsedTime = backoff.WithMaxElapsedTime
)

func RetryFunc(ctx context.Context, f Func, opts ...backoff.RetryOption) error {
	op := func() (any, error) {
		err := f()
		return nil, err
	}
	_, err := Retry(ctx, op, opts...)
	return err
}

func Retry[T any](ctx context.Context, op backoff.Operation[T], opts ...backoff.RetryOption) (T, error) {
	opts = append([]backoff.RetryOption{
		backoff.WithBackOff(New()),
		backoff.WithMaxElapsedTime(math.MaxInt64),
	}, opts...)
	return backoff.Retry(
		ctx,
		op,
		opts...,
	)
}

func New() *backoff.ExponentialBackOff {
	b := &backoff.ExponentialBackOff{
		InitialInterval:     500 * time.Millisecond,
		RandomizationFactor: 0.25,
		Multiplier:          1.5,
		MaxInterval:         2 * time.Second,
	}
	b.Reset()
	return b
}
