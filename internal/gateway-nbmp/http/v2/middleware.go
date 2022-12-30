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
	"github.com/gofiber/fiber/v2"

	"github.com/nagare-media/engine/pkg/http"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

func HandleRequest(c *fiber.Ctx) error {
	// check Accepts header
	accepts := c.Accepts(
		nbmpv2.FunctionDescriptionDocumentMIMEType,
		nbmpv2.TaskDescriptionDocumentMIMEType,
		nbmpv2.WorkflowDescriptionDocumentMIMEType,
	)
	if accepts == "" {
		return fiber.ErrNotAcceptable
	}

	// check Content-Type header
	if http.WriteRequest(c) {
		ct := string(c.Request().Header.ContentType())
		switch ct {
		// if no Content-Types is set, we assume the body is still a correct NBMP description document
		case "":
			c.Request().Header.SetContentType(fiber.MIMEApplicationJSONCharsetUTF8)

		case
			// correct Content-Types (TODO: properly parse MIME type including charset)
			nbmpv2.FunctionDescriptionDocumentMIMEType,
			nbmpv2.TaskDescriptionDocumentMIMEType,
			nbmpv2.WorkflowDescriptionDocumentMIMEType,
			// we allow generic JSON Content-Type types and check during unmarshal
			fiber.MIMEApplicationJSON,
			fiber.MIMEApplicationJSONCharsetUTF8:

		default:
			return fiber.ErrUnsupportedMediaType
		}
	}

	return c.Next()
}
