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
	"errors"
	"reflect"
	"sort"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
)

func GetAnnotatedObject(ctx context.Context, c client.Client, obj client.Object, annotationKey, annotationVal string, opts ...client.ListOption) (client.Object, error) {
	gvk, err := apiutil.GVKForObject(runtime.Object(obj), c.Scheme())
	if err != nil {
		return nil, err
	}

	// TODO: this handles 99% of all cases, but is not strictly true
	gvk.Kind = gvk.Kind + "List"

	runObjList, err := c.Scheme().New(gvk)
	if err != nil {
		return nil, err
	}

	objList, ok := runObjList.(client.ObjectList)
	if !ok {
		return nil, errors.New("failed to convert runtime.Object to client.Object")
	}

	if err := c.List(ctx, objList, opts...); err != nil {
		return nil, err
	}

	// TODO: find a better way than reflection
	items := reflect.ValueOf(objList).Elem().FieldByName("Items")

	selectedItems := make([]client.Object, 0)
	for i := 0; i < items.Len(); i++ {
		objPtr := reflect.New(items.Index(i).Type())
		objPtr.Elem().Set(items.Index(i))
		obj := objPtr.Interface().(client.Object)
		if val := obj.GetAnnotations()[annotationKey]; val == annotationVal {
			selectedItems = append(selectedItems, obj)
		}
	}

	if len(selectedItems) == 0 {
		mapping, err := c.RESTMapper().RESTMapping(gvk.GroupKind(), gvk.Version)
		if err != nil {
			return nil, err
		}
		return nil, apierrors.NewNotFound(mapping.Resource.GroupResource(), "annotatedObj")
	}

	sort.Slice(selectedItems, func(i, j int) bool {
		if selectedItems[i].GetCreationTimestamp().UnixNano() == selectedItems[j].GetCreationTimestamp().UnixNano() {
			return selectedItems[i].GetName() < selectedItems[j].GetName()
		}
		return selectedItems[i].GetCreationTimestamp().UnixNano() > selectedItems[j].GetCreationTimestamp().UnixNano()
	})

	return selectedItems[0], nil
}
