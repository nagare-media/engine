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

package web

import (
	"context"

	"github.com/gofiber/fiber/v2"
	"sigs.k8s.io/controller-runtime/pkg/log"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
)

type server struct {
	cfg enginev1.WebserverConfiguration
}

func NewServer(cfg enginev1.WebserverConfiguration) *server {
	cfgCopy := cfg.DeepCopy()
	cfgCopy.Default()
	return &server{
		cfg: *cfgCopy,
	}
}

func (s *server) Start(ctx context.Context) error {
	l := log.FromContext(ctx).WithName("web")

	l.Info("start webserver")

	app := fiber.New(fiber.Config{
		ServerHeader:          "nagare media engine",
		AppName:               "nagare media engine gateway-nbmp",
		DisableStartupMessage: true,

		StrictRouting:                true,
		CaseSensitive:                true,
		DisableDefaultContentType:    true, // we set Content-Type ourselves
		DisablePreParseMultipartForm: true, // we don't expect multipart/form-data requests

		ReadTimeout:  *s.cfg.ReadTimeout,
		WriteTimeout: *s.cfg.WriteTimeout,
		IdleTimeout:  *s.cfg.IdleTimeout,
		Network:      *s.cfg.Network,

		// TODO: ErrorHandler???

		RequestMethods: []string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodDelete,
			fiber.MethodPatch,

			// These methods are not used in the NBMP specification, but can be mapped to corresponding methods
			fiber.MethodHead, // HEAD is mapped to GET without a body
			fiber.MethodPut,  // PUT is mapped to PATCH
		},
	})

	// TODO: define routes

	var err error
	listenDone := make(chan struct{})
	go func() {
		err = app.Listen(*s.cfg.BindAddress)
		if err != nil {
			l.Error(err, "problem starting listener")
		}
		close(listenDone)
	}()

	select {
	case <-ctx.Done():
		l.Info("stop webserver")
		err = app.Shutdown()
		if err != nil {
			l.Error(err, "problem stopping webserver")
		}
	case <-listenDone:
	}

	l.Info("webserver stopped")

	return err
}
