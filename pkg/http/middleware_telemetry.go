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

package http

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"

	"sigs.k8s.io/controller-runtime/pkg/log"
)

func TelemetryMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctx := ContextFromFiberCtx(c)
		l := log.FromContext(ctx)

		responseStart := time.Now()
		err := c.Next()
		responseTime := time.Since(responseStart)

		// request
		remoteAddr := c.Context().RemoteAddr()
		hostname := c.Hostname()
		method := c.Method()
		path := c.Path()
		referer := c.Request().Header.Referer()
		userAgent := string(c.Request().Header.UserAgent())

		// response
		status := c.Response().StatusCode()
		if err != nil {
			if e, ok := err.(*fiber.Error); ok {
				status = e.Code
			}
		}

		// both
		requestID := RequestIDFromFiberCtx(c)

		// logging
		l.V(2).Info(
			fmt.Sprintf("%d %s %s%s", status, method, hostname, path),

			// request
			"remoteAddr", remoteAddr,
			"hostname", hostname,
			"method", method,
			"path", path,
			"referer", referer,
			"userAgent", userAgent,

			// response
			"status", status,
			"responseTime", responseTime,

			// both
			"requestID", requestID,
		)

		// TODO: metrics
		// TODO: tracing
		return err
	}
}
