/*
Copyright 2023 The nagare media authors

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

package utils

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
)

func GetMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func FilterIsNotFoundErr(err error) error {
	agErr, ok := err.(kerrors.Aggregate)
	if !ok {
		return err
	}
	var errs []error
	for _, err := range agErr.Errors() {
		if !apierrors.IsNotFound(err) {
			errs = append(errs, err)
		}
	}
	return kerrors.NewAggregate(errs)
}

func AppendAggregateErr(err1, err2 error) error {
	agErr, ok := err1.(kerrors.Aggregate)
	if !ok {
		return kerrors.NewAggregate([]error{err1, err2})
	}
	return kerrors.NewAggregate(append(agErr.Errors(), err2))
}
