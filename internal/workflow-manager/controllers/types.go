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

package controllers

import (
	"context"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type EnsureFunc[T any] func(context.Context, T) (Result, error)

type Result struct {
	Requeue      bool
	RequeueAfter time.Duration
	Stop         bool
}

func (r *Result) ReconcileResult() reconcile.Result {
	return reconcile.Result{
		Requeue:      r.Requeue,
		RequeueAfter: r.RequeueAfter,
	}
}

func ApplyEnsureFuncs[T any](ctx context.Context, obj T, funcs []EnsureFunc[T]) (Result, error) {
	for _, f := range funcs {
		res, err := f(ctx, obj)
		if err != nil || res.Stop {
			return res, err
		}
	}
	return Result{}, nil
}
