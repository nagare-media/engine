/*
Copyright 2022 The nagare media authors

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

package http

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"sigs.k8s.io/controller-runtime/pkg/log"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	nbmpv2 "github.com/nagare-media/engine/internal/gateway-nbmp/http/v2"
	"github.com/nagare-media/engine/internal/gateway-nbmp/http/v2/workflowapi"
	"github.com/nagare-media/engine/pkg/http"
)

type server struct {
	cfg *enginev1.WebserverConfiguration
	app *fiber.App
	ctx context.Context
}

func NewServer(cfg enginev1.WebserverConfiguration) *server {
	cfgCopy := cfg.DeepCopy()
	cfgCopy.Default()

	s := &server{
		cfg: cfgCopy,
		app: fiber.New(fiber.Config{
			ServerHeader:          "nagare media engine",
			AppName:               "nagare media engine gateway-nbmp",
			DisableStartupMessage: true,

			CaseSensitive:                true,
			DisableDefaultContentType:    true, // we set Content-Type ourselves
			DisablePreParseMultipartForm: true, // we don't expect multipart/form-data requests

			ReadTimeout:  *cfgCopy.ReadTimeout,
			WriteTimeout: *cfgCopy.WriteTimeout,
			IdleTimeout:  *cfgCopy.IdleTimeout,
			Network:      *cfgCopy.Network,

			RequestMethods: []string{
				fiber.MethodGet,
				fiber.MethodPost,
				fiber.MethodDelete,
				fiber.MethodPatch,

				// These methods are not used in the NBMP specification, but can be mapped to corresponding methods
				fiber.MethodHead, // HEAD is mapped to GET without a body
				fiber.MethodPut,  // PUT is mapped to PATCH
			},
		}),
	}

	// global middlewares
	s.app.
		Use(http.ContextMiddleware(func() context.Context { return s.ctx })).
		Use(http.TelemetryMiddleware()).
		Use(http.RequestIDMiddleware())

	// Health API
	s.app.Get("/healthz", http.HealthRequestHandler())
	s.app.Get("/readyz", http.HealthRequestHandler())

	// NBMP 2nd edition APIs
	s.app.Group("/v2").
		// middlewares
		Use(nbmpv2.HandleRequest).
		// APIs
		Mount("/workflows", workflowapi.New(cfgCopy).App())

	return s
}

func (s *server) Start(ctx context.Context) error {
	l := log.FromContext(ctx).WithName("http")
	s.ctx = log.IntoContext(ctx, l)

	l.Info("start webserver")

	var err error
	listenDone := make(chan struct{})
	go func() {
		err = s.app.Listen(*s.cfg.BindAddress)
		if err != nil {
			l.Error(err, "problem starting listener")
		}
		close(listenDone)
	}()

	select {
	case <-s.ctx.Done():
		l.Info("stop webserver")
		err = s.app.Shutdown()
		if err != nil {
			l.Error(err, "problem stopping webserver")
		}
	case <-listenDone:
	}

	l.Info("webserver stopped")

	return err
}
