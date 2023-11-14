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

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"sigs.k8s.io/controller-runtime/pkg/log"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	enginehttp "github.com/nagare-media/engine/internal/pkg/http"
	eventshttp "github.com/nagare-media/engine/pkg/events/http"
	"github.com/nagare-media/engine/pkg/http"
	"github.com/nagare-media/engine/pkg/starter"
)

type reportsCtrl struct {
	cfg  *enginev1.WorkflowManagerHelperReportsControllerConfiguration
	data *enginev1.WorkflowManagerHelperData

	http     starter.Starter
	httpDone chan struct{}
	eventCh  chan cloudevents.Event
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
		cfg:      cfg,
		http:     s,
		httpDone: make(chan struct{}),
		eventCh:  eventCh,
	}
}

func (c *reportsCtrl) Start(ctx context.Context) error {
	l := log.FromContext(ctx).WithName("reports")
	ctx = log.IntoContext(ctx, l)

	// start http server
	var httpErr error
	go func() {
		httpErr = c.http.Start(ctx)
		close(c.httpDone)
	}()

	// TODO: connect to NATS

	// event loop
	for {
		select {
		case <-c.httpDone:
			l.Error(httpErr, "webserver terminated unexpectedly")
			return httpErr

		case <-ctx.Done():
			l.Info("termination requested")
			<-c.httpDone
			return nil

		case e := <-c.eventCh:
			if err := c.handleEvent(ctx, e); err != nil {
				l.Error(err, "handling event failed")
			}
		}
	}
}

func (c *reportsCtrl) handleEvent(ctx context.Context, event cloudevents.Event) error {
	// l := log.FromContext(ctx)
	// TODO: send event to NATS
	return nil
}
