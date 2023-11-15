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

package workflowmanagerhelper

import (
	"context"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
)

func newBackOffWithContext(ctx context.Context) backoff.BackOff {
	expBackOff := &backoff.ExponentialBackOff{
		InitialInterval:     500 * time.Millisecond,
		RandomizationFactor: 0.25,
		Multiplier:          1.5,
		MaxInterval:         10 * time.Second,
		MaxElapsedTime:      0, // = indefinitely (we use contexts for that)
		Stop:                backoff.Stop,
		Clock:               backoff.SystemClock,
	}
	expBackOff.Reset()
	return backoff.WithContext(expBackOff, ctx)
}
