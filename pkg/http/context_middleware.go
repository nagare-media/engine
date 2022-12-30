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
)

type ctxLocalKey struct{}

type ContextLoader interface {
	LoadContext() context.Context
}

func ContextMiddleware(ctxLoader func() context.Context) fiber.Handler {
	return func(c *fiber.Ctx) error {
		c.Locals(ctxLocalKey{}, ctxLoader())
		return c.Next()
	}
}

func ContextFromFiberCtx(c *fiber.Ctx) context.Context {
	val := c.Locals(ctxLocalKey{})
	if val == nil {
		return context.Background()
	}

	if ctx, ok := val.(context.Context); ok {
		return ctx
	}
	return context.Background()
}

func ContextIntoFiberCtx(c *fiber.Ctx, ctx context.Context) {
	c.Locals(ctxLocalKey{}, ctx)
}
