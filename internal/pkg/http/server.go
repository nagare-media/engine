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

package http

import (
	"context"

	"github.com/gofiber/fiber/v2"

	"sigs.k8s.io/controller-runtime/pkg/log"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/pkg/http"
)

type server struct {
	App *fiber.App

	cfg *enginev1.WebserverConfiguration
	ctx context.Context
}

func NewServer(cfg *enginev1.WebserverConfiguration) *server {
	s := &server{
		cfg: cfg,
		App: fiber.New(fiber.Config{
			ServerHeader:          "nme",
			AppName:               "nagare media engine",
			DisableStartupMessage: true,

			CaseSensitive:                true,
			DisableDefaultContentType:    true, // we set Content-Type ourselves
			DisablePreParseMultipartForm: true, // we don't expect multipart/form-data requests

			ReadTimeout:  cfg.ReadTimeout.Duration,
			WriteTimeout: cfg.WriteTimeout.Duration,
			IdleTimeout:  cfg.IdleTimeout.Duration,
			Network:      *cfg.Network,

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
	s.App.
		Use(http.ContextMiddleware(func() context.Context { return s.ctx })).
		Use(http.RequestIDMiddleware())

	return s
}

func (s *server) Start(ctx context.Context) error {
	l := log.FromContext(ctx).WithName("http")
	s.ctx = log.IntoContext(ctx, l)

	l.Info("start webserver", "bind-address", *s.cfg.BindAddress)

	var err error
	listenDone := make(chan struct{})
	go func() {
		err = s.App.Listen(*s.cfg.BindAddress)
		if err != nil {
			l.Error(err, "unable to start webserver")
		}
		close(listenDone)
	}()

	select {
	case <-s.ctx.Done():
		l.Info("stop webserver")
		err = s.App.Shutdown()
		if err != nil {
			l.Error(err, "unable to stop webserver gracefully")
		}
	case <-listenDone:
	}

	l.Info("webserver stopped")

	return err
}
