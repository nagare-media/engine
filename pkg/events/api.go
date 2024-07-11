/*
Copyright 2022-2024 The nagare media authors

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

package events

import (
	"encoding/json"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/gofiber/fiber/v2"

	"github.com/nagare-media/engine/internal/pkg/mime"
	"github.com/nagare-media/engine/pkg/http"
)

type eventAPI struct {
	h chan<- cloudevents.Event
}

var _ http.API = &eventAPI{}

func API(handler chan<- cloudevents.Event) *eventAPI {
	return &eventAPI{
		h: handler,
	}
}

func (api *eventAPI) MountTo(r fiber.Router) {
	r.Post("/events", api.handleRequest)
}

func (api *eventAPI) handleRequest(c *fiber.Ctx) error {
	var body []byte
	var events []cloudevents.Event

	// handle request based on Content-Type header
	// TODO: must this be split on ';'?
	ct := string(c.Request().Header.ContentType())
loop:
	for {
		switch ct {
		case "":
			// decide Content-Type based on body
			// TODO: should we be more strict?
			body = c.Body()
			if body[0] == '{' {
				ct = mime.ApplicationCloudEventsJSON
			} else if body[0] == '[' {
				ct = mime.ApplicationCloudEventsBatchJSON
			} else {
				ct = mime.ApplicationOctetStream
			}
			c.Request().Header.SetContentType(ct)

		case cloudevents.ApplicationCloudEventsJSON:
			// Structured Content Mode
			if len(body) == 0 {
				body = c.Body()
			}
			e := cloudevents.Event{}
			if err := json.Unmarshal(body, &e); err != nil {
				return fiber.ErrBadRequest
			}
			events = append(events, e)
			break loop

		case cloudevents.ApplicationCloudEventsBatchJSON:
			// Batched Content Mode
			if len(body) == 0 {
				body = c.Body()
			}
			if err := json.Unmarshal(body, &events); err != nil {
				return fiber.ErrBadRequest
			}
			break loop

		default:
			// TODO: implement support for Binary Content Mode
			return fiber.ErrUnsupportedMediaType
		}
	}

	// emit events event handler for each event
	if c.QueryBool("async", false) {
		go api.emitEvents(events)
		return c.SendStatus(fiber.StatusAccepted)
	}
	api.emitEvents(events)
	return c.SendStatus(fiber.StatusOK)
}

// emitEvents to event handler
func (api *eventAPI) emitEvents(events []cloudevents.Event) {
	for _, e := range events {
		api.h <- e
	}
}
