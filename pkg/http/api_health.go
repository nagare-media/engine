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

import "github.com/gofiber/fiber/v2"

var DefaultHealthFunc = func() error { return nil }

type HealthFunc func() error

type healthAPI struct {
	HealthFunc HealthFunc
}

var _ API = &healthAPI{}

func HealthAPI(h HealthFunc) *healthAPI {
	return &healthAPI{
		HealthFunc: h,
	}
}

func (api *healthAPI) MountTo(r fiber.Router) {
	r.Get("/healthz", api.handleRequest).
		Get("/readyz", api.handleRequest)
}

func (api *healthAPI) handleRequest(c *fiber.Ctx) error {
	// TODO: implement something like https://datatracker.ietf.org/doc/html/draft-inadarei-api-health-check
	return api.HealthFunc()
}
