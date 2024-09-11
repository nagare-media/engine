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

package v2

import (
	"encoding/json"

	"github.com/gofiber/fiber/v2"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/nagare-media/engine/pkg/nbmp"
)

const DefaultRetryAfterSeconds = "10"

func handleErr(c *fiber.Ctx, obj any, svcErr error, contentType string) error {
	var s int
	switch {
	case svcErr == nbmp.ErrRetryLater:
		c.Status(fiber.StatusAccepted) // 202
		c.Set(fiber.HeaderRetryAfter, DefaultRetryAfterSeconds)
		// respond without a body
		return nil

	case svcErr == nbmp.ErrInvalid:
		s = fiber.StatusBadRequest // 400

	case svcErr == nbmp.ErrNotFound,
		apierrors.IsNotFound(svcErr),
		apierrors.IsGone(svcErr):
		// respond without a body
		return fiber.ErrNotFound // 404

	case svcErr == nbmp.ErrAlreadyExists,
		apierrors.IsAlreadyExists(svcErr):
		s = fiber.StatusConflict // 409

	case svcErr == nbmp.ErrUnsupported:
		s = fiber.StatusUnprocessableEntity // 422

	case apierrors.IsServiceUnavailable(svcErr),
		apierrors.IsUnexpectedServerError(svcErr),
		apierrors.IsUnexpectedObjectError(svcErr):
		s = fiber.StatusBadGateway // 502

	case apierrors.IsTimeout(svcErr),
		apierrors.IsServerTimeout(svcErr):
		s = fiber.StatusGatewayTimeout // 504

	default:
		// respond without a body
		return fiber.ErrInternalServerError // 500
	}

	c.Status(s)
	c.Set(fiber.HeaderContentType, contentType)
	respBody, err := json.Marshal(obj)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	return c.Send(respBody)
}
