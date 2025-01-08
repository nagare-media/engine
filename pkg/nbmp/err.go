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

package nbmp

import "errors"

var (
	ErrNotFound        = errors.New("nbmp: resource not found")
	ErrAlreadyExists   = errors.New("nbmp: resource already exists")
	ErrUnsupported     = errors.New("nbmp: resource has unsupported descriptions")
	ErrInvalid         = errors.New("nbmp: resource is invalid")
	ErrRetryLater      = errors.New("nbmp: resource accept but retry later") // TODO: should include a retry-after field
	ErrUnexpected      = errors.New("nbmp: response is unexpected")
	ErrUnauthenticated = errors.New("nbmp: client is unauthenticated")
	ErrUnauthorized    = errors.New("nbmp: client is unauthorized")
)
