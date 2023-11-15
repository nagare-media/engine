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
	"strings"

	"github.com/nats-io/nats.go/jetstream"
)

const (
	StreamName    = "nagare-media-engine"
	SubjectPrefix = "nme"
)

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
