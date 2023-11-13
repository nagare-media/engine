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

package taskshim

import (
	"context"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	enginehttp "github.com/nagare-media/engine/internal/pkg/http"
	"github.com/nagare-media/engine/internal/task-shim/svc"
	"github.com/nagare-media/engine/pkg/http"
	nbmphttpv2 "github.com/nagare-media/engine/pkg/nbmp/http/v2"
	nbmpsvcv2 "github.com/nagare-media/engine/pkg/nbmp/svc/v2"
	"github.com/nagare-media/engine/pkg/starter"
)

func New(ctx context.Context, terminateFunc func(), cfg *enginev1.TaskShimConfiguration) starter.Starter {
	s := enginehttp.NewServer(&cfg.Webserver)

	// Health API

	http.HealthAPI(http.DefaultHealthFunc).MountTo(s.App)

	// NBMP 2nd edition APIs

	svc :=
		nbmpsvcv2.TaskDefaulterMiddleware(
			nbmpsvcv2.TaskValidatorSpecLaxMiddleware(
				svc.TaskValidatorMiddleware(
					svc.NewTaskService(ctx, terminateFunc, &cfg.TaskService),
				),
			),
		)

	r := s.App.Group("/v2")
	// middlewares
	r.Use(http.TelemetryMiddleware())
	// APIs
	nbmphttpv2.TaskAPI(&cfg.Webserver, svc).MountTo(r.Group("/tasks"))

	return s
}
