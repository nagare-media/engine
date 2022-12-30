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

package v2

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"

	"github.com/nagare-media/engine/internal/gateway-nbmp/svc"
)

func SvcErrorHandler(c *fiber.Ctx, obj interface{}, svcErr error) error {
	var s int
	switch svcErr {
	// responses without a body
	default:
		return fiber.ErrInternalServerError
	case svc.ErrNotFound:
		return fiber.ErrNotFound

	// responses with a body including error details
	case svc.ErrUnsupported:
		s = fiber.StatusUnprocessableEntity
	case svc.ErrInvalid:
		s = fiber.StatusBadRequest
	}

	c.Status(s)
	respBody, err := json.Marshal(obj)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	return c.Send(respBody)
}
