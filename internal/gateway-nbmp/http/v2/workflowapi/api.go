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

package workflowapi

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/gofiber/fiber/v2"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/internal/gateway-nbmp/http/api"
	nbmpapiv2 "github.com/nagare-media/engine/internal/gateway-nbmp/http/v2"
	"github.com/nagare-media/engine/internal/gateway-nbmp/svc"
	"github.com/nagare-media/engine/pkg/http"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

type workflowapi struct {
	cfg   *enginev1.GatewayNBMPConfiguration
	wfsvc svc.WorkflowService
}

var _ api.API = &workflowapi{}

func New(cfg *enginev1.GatewayNBMPConfiguration, wfsvc svc.WorkflowService) *workflowapi {
	return &workflowapi{
		cfg:   cfg,
		wfsvc: wfsvc,
	}
}

func (wfapi *workflowapi) App() *fiber.App {
	app := fiber.New()
	app.
		Post("/", wfapi.handleRequest(wfapi.wfsvc.Create)).
		Patch("/:id", wfapi.handleRequest(wfapi.wfsvc.Update)).
		Put("/:id", wfapi.handleRequest(wfapi.wfsvc.Update)). // the NBMP standard does not define PUT
		Delete("/:id", wfapi.handleRequest(wfapi.wfsvc.Delete)).
		Get("/:id", wfapi.handleRequest(wfapi.wfsvc.Retrieve))
	return app
}

func (wfapi *workflowapi) handleRequest(svcCall func(ctx context.Context, wf *nbmpv2.Workflow) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := http.ContextFromFiberCtx(c)

		// parse request
		wf := &nbmpv2.Workflow{}
		if http.ReadRequest(c) {
			// construct initial description
			wf.General = nbmpv2.General{
				ID: c.Params("id"),
			}
		} else {
			// decode body
			dec := json.NewDecoder(bytes.NewReader(c.Body()))
			dec.DisallowUnknownFields()
			err := dec.Decode(wf)
			if err != nil {
				return fiber.ErrBadRequest
			}
		}

		// service call
		err := svcCall(ctx, wf)
		if err != nil {
			return nbmpapiv2.SvcErrorHandler(c, wf, err)
		}

		// create response
		respBody, err := json.Marshal(wf)
		if err != nil {
			return fiber.ErrInternalServerError
		}

		var selfURL string
		if wfapi.cfg.Webserver.PublicBaseURL == nil {
			selfURL = c.BaseURL()
		} else {
			selfURL = *wfapi.cfg.Webserver.PublicBaseURL
		}
		selfURL += "/" + c.Path() + "/" + wf.General.ID
		// TODO: the NBMP standard requires (SHOULD) a link object in the WDD response. The JSON schema definition does not
		//       specify a link object.

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
