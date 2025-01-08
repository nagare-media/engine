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
	"net/http"

	"github.com/nagare-media/engine/pkg/nbmp"
)

func handleStatusCode(status int) error {
	switch status {
	case http.StatusAccepted: // 202
		return nbmp.ErrRetryLater

	case http.StatusBadRequest: // 400
		return nbmp.ErrInvalid

	case http.StatusUnauthorized: // 401
		return nbmp.ErrUnauthenticated

	case http.StatusForbidden: // 403
		return nbmp.ErrUnauthorized

	case http.StatusNotFound: // 404
		return nbmp.ErrNotFound

	case http.StatusConflict: // 409
		return nbmp.ErrAlreadyExists

	case http.StatusUnprocessableEntity: // 422
		return nbmp.ErrUnsupported
	}

	if 200 <= status && status <= 299 {
		return nil
	}

	return nbmp.ErrUnexpected
}
