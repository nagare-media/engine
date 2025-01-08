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

package http

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/gofiber/fiber/v2"
	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	engineio "github.com/nagare-media/engine/internal/functions/io"
	enginehttp "github.com/nagare-media/engine/internal/pkg/http"
	"github.com/nagare-media/engine/pkg/http"
	"github.com/nagare-media/engine/pkg/starter"
	"k8s.io/utils/ptr"
)

const (
	Protocol          = "http"
	DefaultPortNumber = uint16(8080)
	DefaultTimeout    = 30 * time.Minute
)

var (
	// TODO: make configurable
	DefaultWebserverConfig = enginev1.WebserverConfig{
		BindAddress:   ptr.To(fmt.Sprintf(":%d", DefaultPortNumber)),
		ReadTimeout:   &metav1.Duration{Duration: DefaultTimeout},
		WriteTimeout:  &metav1.Duration{Duration: DefaultTimeout},
		IdleTimeout:   &metav1.Duration{Duration: DefaultTimeout},
		Network:       ptr.To("tcp"),
		PublicBaseURL: ptr.To(fmt.Sprintf("http://127.0.0.1:%d", DefaultPortNumber)),
	}
)

type Server interface {
	engineio.Server

	Router() fiber.Router
}

type server struct {
	s starter.Starter
	r fiber.Router
}

var _ Server = &server{}

func (s *server) Start(ctx context.Context) error {
	return s.s.Start(ctx)
}

func (s *server) Protocol() string {
	return Protocol
}

func (s *server) Router() fiber.Router {
	return s.r
}

func NewServer(portNumber uint16) (engineio.Server, error) {
	cfg := &enginev1.WebserverConfig{
		BindAddress: ptr.To(fmt.Sprintf(":%d", portNumber)),
	}
	cfg.DefaultWithValuesFrom(DefaultWebserverConfig)

	s := enginehttp.NewServer(cfg)
	// global middlewares
	s.App.Use(http.TelemetryMiddleware())

	srv := &server{
		s: s,
		r: s.App,
	}

	return srv, nil
}

func init() {
	engineio.ServerBuilders.Register(Protocol, NewServer)
}
