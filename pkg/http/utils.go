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

import "github.com/gofiber/fiber/v2"

func IsReadRequest(c *fiber.Ctx) bool {
	m := c.Method()
	return m == fiber.MethodGet ||
		m == fiber.MethodHead ||
		m == fiber.MethodOptions ||
		m == fiber.MethodTrace
}

func IsWriteRequest(c *fiber.Ctx) bool {
	return !IsReadRequest(c)
}
