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
	"encoding/json"
	"fmt"
	"time"

	backoff "github.com/cenkalti/backoff/v4"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"sigs.k8s.io/controller-runtime/pkg/log"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	enginehttp "github.com/nagare-media/engine/internal/pkg/http"
	enginenats "github.com/nagare-media/engine/internal/pkg/nats"
	eventshttp "github.com/nagare-media/engine/pkg/events/http"
	"github.com/nagare-media/engine/pkg/http"
	"github.com/nagare-media/engine/pkg/starter"
)

const (
	// TODO: make configurable
	ReportNATSConnectionTimeout = 30 * time.Second
	ReportNATSFlushTimeout      = 10 * time.Second
)

type reportsCtrl struct {
	cfg  *enginev1.WorkflowManagerHelperReportsControllerConfiguration
	data *enginev1.WorkflowManagerHelperData

	http    starter.Starter
	eventCh chan cloudevents.Event
	nc      *nats.Conn
	js      jetstream.JetStream
}

var _ starter.Starter = &reportsCtrl{}

func NewReportsController(cfg *enginev1.WorkflowManagerHelperReportsControllerConfiguration, data *enginev1.WorkflowManagerHelperData) starter.Starter {
	s := enginehttp.NewServer(&cfg.Webserver)

	// Health API

	http.HealthAPI(http.DefaultHealthFunc).MountTo(s.App)

	// Event API

	eventCh := make(chan cloudevents.Event)
	eventshttp.API(eventCh).MountTo(s.App)

	return &reportsCtrl{
		cfg:     cfg,
		data:    data,
		http:    s,
		eventCh: eventCh,
	}
}

func (c *reportsCtrl) Start(ctx context.Context) error {
	l := log.FromContext(ctx).WithName("reports")
	ctx = log.IntoContext(ctx, l)

	// start http server
	var (
		httpErr  error
		httpDone = make(chan struct{})
	)
	go func() { httpErr = c.http.Start(ctx); close(httpDone) }()

	// connect to NATS
	if err := c.createJetStream(ctx); err != nil {
		return err
	}
	defer func() {
		if err := c.nc.FlushTimeout(ReportNATSFlushTimeout); err != nil {
			l.Error(err, "failed to flush undelivered messages")
		}
	}()
	defer c.nc.Close()

	// ensure stream exists for messages to persist
	// TODO: add support for streams created by user
	_, err := enginenats.CreateOrUpdateEngineStream(ctx, c.js)
	if err != nil {
		return err
	}

	subjectName := enginenats.Subject(enginenats.SubjectPrefix, c.data.Workflow.ID, c.data.Task.ID)

	// event loop
	for {
		select {
		case <-httpDone:
			if httpErr != nil {
				l.Error(httpErr, "webserver terminated unexpectedly")
			}
			return httpErr

		case <-ctx.Done():
			l.Info("termination requested")
			<-httpDone
			return nil

		case e := <-c.eventCh:
			d, err := json.Marshal(e)
			if err != nil {
				l.Error(err, "failed to encode event for publication")
				continue
			}

			_, err = c.js.PublishMsg(ctx, &nats.Msg{
				Subject: subjectName,
				Data:    d,
			})
			if err != nil {
				l.Error(err, "failed to publish event")
				continue
			}
		}
	}
}

func (c *reportsCtrl) createJetStream(ctx context.Context) error {
	l := log.FromContext(ctx)

	op := func() error {
		l.Info("establishing connection to NATS")
		var (
			err error
			nc  *nats.Conn
			js  jetstream.JetStream
		)

		// cleanup if there was an error
		defer func() {
			if err != nil {
				nc.Close()
			}
		}()

		nc, err = nats.Connect(string(c.cfg.NATS.URL))
		if err != nil {
			return err
		}

		js, err = jetstream.New(nc)
		if err != nil {
			return err
		}

		l.Info("connected to NATS")
		c.nc = nc
		c.js = js
		return nil
	}

	no := func(err error, t time.Duration) {
		l.Error(err, fmt.Sprintf("failed; retrying after %s", t))
	}

	ctx, cancle := context.WithTimeout(ctx, ReportNATSConnectionTimeout)
	defer cancle()
	return backoff.RetryNotify(op, newBackOffWithContext(ctx), no)
}
