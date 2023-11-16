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

package nats

import (
	"context"
	"fmt"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/log"

	backoff "github.com/cenkalti/backoff/v4"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

const (
	// TODO: make configurable
	NATSConnectionTimeout = 30 * time.Second
)

const (
	StreamName    = "nagare-media-engine"
	SubjectPrefix = "nme"
)

func CreateJetStreamConn(ctx context.Context, url string) (*nats.Conn, jetstream.JetStream, error) {
	l := log.FromContext(ctx)

	var (
		nc *nats.Conn
		js jetstream.JetStream
	)

	op := func() error {
		l.Info("establishing connection to NATS")
		var err error

		// cleanup if there was an error
		defer func() {
			if err != nil {
				nc.Close()
			}
		}()

		nc, err = nats.Connect(url)
		if err != nil {
			return err
		}

		js, err = jetstream.New(nc)
		if err != nil {
			return err
		}

		l.Info("connected to NATS")
		return nil
	}

	no := func(err error, t time.Duration) {
		l.Error(err, fmt.Sprintf("failed; retrying after %s", t))
	}

	expBackOff := &backoff.ExponentialBackOff{
		InitialInterval:     500 * time.Millisecond,
		RandomizationFactor: 0.25,
		Multiplier:          1.5,
		MaxInterval:         5 * time.Second,
		MaxElapsedTime:      0, // = indefinitely (we use contexts for that)
		Stop:                backoff.Stop,
		Clock:               backoff.SystemClock,
	}
	expBackOff.Reset()

	ctx, cancel := context.WithTimeout(ctx, NATSConnectionTimeout)
	defer cancel()

	err := backoff.RetryNotify(op, backoff.WithContext(expBackOff, ctx), no)
	return nc, js, err
}

func CreateOrUpdateEngineStream(ctx context.Context, js jetstream.JetStream) (jetstream.Stream, error) {
	return js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:         StreamName,
		Subjects:     []string{Subject(SubjectPrefix, ">")},
		Retention:    jetstream.LimitsPolicy,
		MaxMsgs:      -1, // unlimited
		MaxBytes:     -1, // unlimited
		MaxAge:       0,  // unlimited
		AllowRollup:  true,
		DenyDelete:   false,
		DenyPurge:    false,
		MaxConsumers: -1, // unlimited
		Storage:      jetstream.FileStorage,
		Replicas:     1,
	})
}

func Subject(tokens ...string) string {
	return strings.Join(tokens, ".")
}
