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

package workflowmanagerhelper

import (
	"context"
	"encoding/json"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"sigs.k8s.io/controller-runtime/pkg/log"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	enginehttp "github.com/nagare-media/engine/internal/pkg/http"
	enginenats "github.com/nagare-media/engine/internal/pkg/nats"
	"github.com/nagare-media/engine/pkg/events"
	"github.com/nagare-media/engine/pkg/http"
	"github.com/nagare-media/engine/pkg/starter"
	"github.com/nagare-media/engine/pkg/updatable"
)

const (
	// TODO: make configurable
	ReportsNATSPublishTimeout = 5 * time.Second
)

type reportsCtrl struct {
	data *enginev1.WorkflowManagerHelperData

	http    starter.Starter
	eventCh chan cloudevents.Event
	nc      *nats.Conn
	js      jetstream.JetStream
}

var _ starter.Starter = &reportsCtrl{}

func NewReportsController(cfg *enginev1.WorkflowManagerHelperReportsControllerConfig, data updatable.Updatable[*enginev1.WorkflowManagerHelperData]) starter.Starter {
	s := enginehttp.NewServer(&cfg.Webserver)

	// Health API

	http.HealthAPI(http.DefaultHealthFunc).MountTo(s.App)

	// Event API

	eventCh := make(chan cloudevents.Event)
	events.API(eventCh).MountTo(s.App)

	return &reportsCtrl{
		data:    data.Get().Value, // assumption: System.NATS.URL, Workflow.ID and Task.ID will not change when data is updated
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
	var err error
	c.nc, c.js, err = enginenats.CreateJetStreamConn(ctx, string(c.data.System.NATS.URL))
	if err != nil {
		return err
	}
	defer c.nc.Close()

	// ensure stream exists for messages to persist
	// TODO: add support for streams created by user
	_, err = enginenats.CreateOrUpdateEngineStream(ctx, c.js)
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
			// Structured Content Mode
			// TODO: move to binary mode?
			d, err := json.Marshal(e)
			if err != nil {
				l.Error(err, "failed to encode event for publication")
				continue
			}

			// we ignore potentially canceled ctx and create new context with timeout
			ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), ReportsNATSPublishTimeout)
			defer cancel()

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
