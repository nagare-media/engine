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

package utils

import (
	"context"

	kerrors "k8s.io/apimachinery/pkg/util/errors"

	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func FullServerSideApply(ctx context.Context, c client.Client, obj client.Object, manager string) error {
	return kerrors.NewAggregate([]error{
		// apply changes to .spec
		ServerSideApply(ctx, c, obj, manager),
		// apply changes to .status
		ServerSideApplyStatus(ctx, c, obj, manager),
	})
}

func ServerSideApply(ctx context.Context, c client.Client, obj client.Object, manager string) error {
	obj.SetManagedFields(nil)
	return c.Patch(ctx, obj, client.Apply, &client.PatchOptions{
		Force:        pointer.Bool(true),
		FieldManager: manager,
	})
}

func ServerSideApplyStatus(ctx context.Context, c client.Client, obj client.Object, manager string) error {
	obj.SetManagedFields(nil)
	return c.Status().Patch(ctx, obj, client.Apply, &client.SubResourcePatchOptions{
		PatchOptions: client.PatchOptions{
			Force:        pointer.Bool(true),
			FieldManager: manager,
		},
	})
}

func FullPatch(ctx context.Context, c client.Client, obj, oldObj client.Object) error {
	return kerrors.NewAggregate([]error{
		// apply changes to non-.status
		Patch(ctx, c, obj, oldObj),
		// apply changes to .status
		PatchStatus(ctx, c, obj, oldObj),
	})
}

func Patch(ctx context.Context, c client.Client, obj, oldObj client.Object) error {
	return c.Patch(ctx, obj.DeepCopyObject().(client.Object), client.MergeFrom(oldObj))
}

func PatchStatus(ctx context.Context, c client.Client, obj, oldObj client.Object) error {
	return c.Status().Patch(ctx, obj.DeepCopyObject().(client.Object), client.MergeFrom(oldObj))
}
