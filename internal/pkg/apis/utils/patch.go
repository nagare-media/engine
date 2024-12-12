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

package utils

import (
	"context"
	"encoding/json"
	"errors"

	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/kyaml/yaml/merge2"
	"sigs.k8s.io/yaml"
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
		Force:        ptr.To(true),
		FieldManager: manager,
	})
}

func ServerSideApplyStatus(ctx context.Context, c client.Client, obj client.Object, manager string) error {
	obj.SetManagedFields(nil)
	return c.Status().Patch(ctx, obj, client.Apply, &client.SubResourcePatchOptions{
		PatchOptions: client.PatchOptions{
			Force:        ptr.To(true),
			FieldManager: manager,
		},
	})
}

func FullPatch(ctx context.Context, c client.Client, obj, oldObj client.Object) error {
	// we create a copy as Patch / PatchStatus will replace obj.
	// Running Patch will thus remove any changes made to status fields.
	objCopy, ok := obj.DeepCopyObject().(client.Object)
	if !ok {
		return errors.New("patch: not a client.Object")
	}
	return kerrors.NewAggregate([]error{
		// apply changes to non-.status
		Patch(ctx, c, objCopy, oldObj),
		// apply changes to .status
		PatchStatus(ctx, c, obj, oldObj),
	})
}

func Patch(ctx context.Context, c client.Client, obj, oldObj client.Object) error {
	patch := client.MergeFrom(oldObj)
	changed, err := HasChangesIn(obj, patch, []string{
		// Common
		"metadata",
		"spec",
		// ConfigMap / Secret
		"immutable",
		"type",
		"data",
		"binaryData",
		"stringData",
	})
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	return c.Patch(ctx, obj, patch)
}

func PatchStatus(ctx context.Context, c client.Client, obj, oldObj client.Object) error {
	patch := client.MergeFrom(oldObj)
	changed, err := HasChangesIn(obj, patch, []string{"status"})
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	return c.Status().Patch(ctx, obj, patch)
}

func HasChangesIn(obj client.Object, patch client.Patch, keys []string) (bool, error) {
	diff, err := patch.Data(obj)
	if err != nil {
		return false, err
	}

	diffMap := make(map[string]any)
	if err = json.Unmarshal(diff, &diffMap); err != nil {
		return false, err
	}

	for _, key := range keys {
		if _, changed := diffMap[key]; changed {
			return true, nil
		}
	}

	return false, nil
}

func StrategicMerge(original, modified, patchedObj any) error {
	var yamlOriginal, yamlModified []byte
	var yamlPatched string
	var err error

	if yamlOriginal, err = yaml.Marshal(original); err != nil {
		return err
	}

	if yamlModified, err = yaml.Marshal(modified); err != nil {
		return err
	}

	if yamlPatched, err = merge2.MergeStrings(string(yamlModified), string(yamlOriginal), true, kyaml.MergeOptions{}); err != nil {
		return err
	}

	if err = yaml.Unmarshal([]byte(yamlPatched), patchedObj); err != nil {
		return err
	}

	return nil
}
