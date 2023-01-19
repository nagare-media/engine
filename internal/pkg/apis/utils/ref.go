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
	"context"
	"errors"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/pkg/apis/meta"
)

var (
	ErrSelectedMultiple = errors.New("selection returned multiple items")
)

func SelectLocalMediaProcessingEntityRef(ctx context.Context, c client.Client, namespace string, sel labels.Selector) (*meta.LocalObjectReference, error) {
	// local MediaProcessingEntity
	mpeList := &enginev1.MediaProcessingEntityList{}
	if err := c.List(ctx, mpeList, client.InNamespace(namespace), client.MatchingLabelsSelector{Selector: sel}); err != nil {
		return nil, err
	}

	if len(mpeList.Items) == 1 {
		return &meta.LocalObjectReference{
			APIVersion: enginev1.GroupVersion.String(),
			Kind:       mpeList.Items[0].GroupVersionKind().Kind,
			Name:       mpeList.Items[0].Name,
		}, nil
	} else if len(mpeList.Items) > 1 {
		return nil, ErrSelectedMultiple
	}

	// ClusterMediaProcessingEntity
	cmpeList := &enginev1.ClusterMediaProcessingEntityList{}
	if err := c.List(ctx, cmpeList, client.MatchingLabelsSelector{Selector: sel}); err != nil {
		return nil, err
	}

	if len(cmpeList.Items) == 1 {
		return &meta.LocalObjectReference{
			APIVersion: enginev1.GroupVersion.String(),
			Kind:       cmpeList.Items[0].GroupVersionKind().Kind,
			Name:       cmpeList.Items[0].Name,
		}, nil
	} else if len(cmpeList.Items) > 1 {
		return nil, ErrSelectedMultiple
	}

	return nil, nil
}

func ResolveRef(ctx context.Context, c client.Client, ref *meta.ObjectReference) (client.Object, error) {
	lref := &meta.LocalObjectReference{
		APIVersion: ref.APIVersion,
		Kind:       ref.Kind,
		Name:       ref.Name,
	}
	return ResolveLocalRef(ctx, c, ref.Namespace, lref)
}

func ResolveLocalRef(ctx context.Context, c client.Client, namespace string, ref *meta.LocalObjectReference) (client.Object, error) {
	gv, err := schema.ParseGroupVersion(ref.APIVersion)
	if err != nil {
		return nil, err
	}
	if gv.Empty() {
		// assume we talk about nagare media engine types
		gv = enginev1.GroupVersion
	}
	gvk := gv.WithKind(ref.Kind)

	runObj, err := c.Scheme().New(gvk)
	if err != nil {
		return nil, err
	}

	obj, ok := runObj.(client.Object)
	if !ok {
		return nil, errors.New("failed to convert runtime.Object to client.Object")
	}

	mapping, err := c.RESTMapper().RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}

	key := client.ObjectKey{Name: ref.Name}
	if mapping.Scope.Name() == apimeta.RESTScopeNameNamespace {
		key.Namespace = namespace
	}

	err = c.Get(ctx, key, obj)
	return obj, err
}

func ToRef(obj client.Object) *meta.ObjectReference {
	return &meta.ObjectReference{
		APIVersion: obj.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		Kind:       obj.GetObjectKind().GroupVersionKind().Kind,
		Name:       obj.GetName(),
		Namespace:  obj.GetNamespace(),
	}
}

func ToLocalRef(obj client.Object) *meta.LocalObjectReference {
	return &meta.LocalObjectReference{
		APIVersion: obj.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		Kind:       obj.GetObjectKind().GroupVersionKind().Kind,
		Name:       obj.GetName(),
	}
}
