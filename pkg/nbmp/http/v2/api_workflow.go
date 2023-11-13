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

package v2

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/gofiber/fiber/v2"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/pkg/http"
	nbmpsvcv2 "github.com/nagare-media/engine/pkg/nbmp/svc/v2"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

type workflowAPI struct {
	cfg *enginev1.WebserverConfiguration
	svc nbmpsvcv2.WorkflowService
}

var _ http.API = &workflowAPI{}

func WorkflowAPI(cfg *enginev1.WebserverConfiguration, svc nbmpsvcv2.WorkflowService) *workflowAPI {
	return &workflowAPI{
		cfg: cfg,
		svc: svc,
	}
}

func (api *workflowAPI) MountTo(r fiber.Router) {
	r.
		// middlewares
		Use(ValidateHeadersMiddleware(nbmpv2.WorkflowDescriptionDocumentMIMEType)).
		// API
		Post("", api.handleRequest(api.svc.Create)).
		Patch("/:id", api.handleRequest(api.svc.Update)).
		Put("/:id", api.handleRequest(api.svc.Update)).
		Delete("/:id", api.handleRequest(api.svc.Delete)).
		Get("/:id", api.handleRequest(api.svc.Retrieve))
}

func (api *workflowAPI) handleRequest(svcCall func(context.Context, *nbmpv2.Workflow) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := http.ContextFromFiberCtx(c)

		// parse request
		w := &nbmpv2.Workflow{}
		if http.IsReadRequest(c) || c.Method() == fiber.MethodDelete {
			// construct initial description
			w.General = nbmpv2.General{
				ID: c.Params("id"),
			}
		} else {
			// decode body
			dec := json.NewDecoder(bytes.NewReader(c.Body()))
			dec.DisallowUnknownFields()
			err := dec.Decode(w)
			if err != nil {
				return fiber.ErrBadRequest
			}
		}

		// service call
		err := svcCall(ctx, w)
		if err != nil {
			return handleErr(c, w, err, nbmpv2.WorkflowDescriptionDocumentMIMEType)
		}

		// create response
		respBody, err := json.Marshal(w)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		// TODO: the NBMP standard requires (SHOULD) a link object in the WDD response. The JSON schema definition does not
		//       specify a link object.
		var selfURL string
		if api.cfg.PublicBaseURL == nil {
			selfURL = c.BaseURL()
		} else {
			selfURL = *api.cfg.PublicBaseURL
		}
		selfURL += "/" + c.Path() + "/" + w.General.ID

		// set status and headers
		c.Set(fiber.HeaderContentType, nbmpv2.WorkflowDescriptionDocumentMIMEType)

		switch c.Method() {
		case fiber.MethodPost:
			// TODO: fiber.StatusAccepted may be more appropriate
			c.Status(fiber.StatusCreated)
			c.Set("Location", selfURL)

		case fiber.MethodPatch, fiber.MethodPut:
			// TODO: the NBMP standard specifies 201 as status code. This is probably a mistake as no new resource is created.
			// TODO: fiber.StatusAccepted may be more appropriate
			c.Status(fiber.StatusOK)

		case fiber.MethodDelete:
			c.Status(fiber.StatusOK)

		case fiber.MethodGet, fiber.MethodHead:
			// TODO: the NBMP standard specifies 201 as status code. This is probably a mistake as no new resource is created.
			c.Status(fiber.StatusOK)
		}

		return c.Send(respBody)
	}
}
