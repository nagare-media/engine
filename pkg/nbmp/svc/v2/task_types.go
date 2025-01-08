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

package v2

import (
	"context"

	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

// TaskService handles NBMP Task API requests.
type TaskService interface {
	Create(context.Context, *nbmpv2.Task) error
	Update(context.Context, *nbmpv2.Task) error
	Delete(context.Context, *nbmpv2.Task) error
	Retrieve(context.Context, *nbmpv2.Task) error
}

// TaskServiceMiddleware is a function that implements a TaskService middleware.
type TaskServiceMiddleware func(next TaskService) TaskService
