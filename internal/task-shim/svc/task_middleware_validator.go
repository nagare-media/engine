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

package svc

import (
	nbmpsvcv2 "github.com/nagare-media/engine/pkg/nbmp/svc/v2"
)

type taskValidatorMiddleware struct {
	nbmpsvcv2.TaskDummyMiddleware

	next nbmpsvcv2.TaskService
}

// TaskValidatorMiddleware validates given task according the nagare media engine.
func TaskValidatorMiddleware(next nbmpsvcv2.TaskService) nbmpsvcv2.TaskService {
	return &taskValidatorMiddleware{
		next: next,
	}
}

var _ nbmpsvcv2.TaskServiceMiddleware = TaskValidatorMiddleware
var _ nbmpsvcv2.TaskService = &taskValidatorMiddleware{}

// TODO: implement task validation