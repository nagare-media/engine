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

package utils

import (
	"context"
	"encoding/json"
	"errors"

	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/kyaml/yaml/merge2"
	"sigs.k8s.io/kustomize/kyaml/yaml/walk"
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

func FullPatch(ctx context.Context, c client.Client, obj, oldObj client.Object) (bool, error) {
	// we create a copy as Patch / PatchStatus will replace obj.
	// Running Patch will thus remove any changes made to status fields.
	objCopy, ok := obj.DeepCopyObject().(client.Object)
	if !ok {
		return false, errors.New("patch: not a client.Object")
	}

	// apply changes to non-.status
	changed1, err1 := Patch(ctx, c, objCopy, oldObj)

	// apply changes to .status
	changed2, err2 := PatchStatus(ctx, c, obj, oldObj)

	return changed1 || changed2, kerrors.NewAggregate([]error{err1, err2})
}

func Patch(ctx context.Context, c client.Client, obj, oldObj client.Object) (bool, error) {
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
		return false, err
	}
	if !changed {
		return false, nil
	}
	return true, c.Patch(ctx, obj, patch)
}

func PatchStatus(ctx context.Context, c client.Client, obj, oldObj client.Object) (bool, error) {
	patch := client.MergeFrom(oldObj)
	changed, err := HasChangesIn(obj, patch, []string{"status"})
	if err != nil {
		return false, err
	}
	if !changed {
		return false, nil
	}
	return true, c.Status().Patch(ctx, obj, patch)
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

func StrategicMergeList[T any](dest *T, objs ...*T) error {
	var (
		err error
		res *yaml.RNode
	)

	// convert to YAML objects
	nodes := make([]*yaml.RNode, 0, len(objs))
	for _, o := range objs {
		if o == nil {
			continue
		}

		var n yaml.Node
		err = n.Encode(o)
		if err != nil {
			return err
		}

		nodes = append(nodes, yaml.NewRNode(&n))
	}

	if len(nodes) == 0 {
		return nil
	}

	if len(nodes) == 1 {
		res = nodes[0]
	} else {
		// TODO: fix merge function
		res, err = walk.Walker{
			Sources:               nodes,
			Visitor:               merge2.Merger{},
			InferAssociativeLists: true,
		}.Walk()
		if err != nil {
			return err
		}
	}

	return res.YNode().Decode(dest)
}
