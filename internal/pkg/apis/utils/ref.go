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
	"errors"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/pkg/apis/meta"
)

var (
	ErrSelectedMultiple = errors.New("selection returned multiple items")
)

func SelectMediaProcessingEntityRef(ctx context.Context, c client.Client, namespace string, sel labels.Selector) (*meta.ObjectReference, error) {
	// local MediaProcessingEntity
	mpeList := &enginev1.MediaProcessingEntityList{}
	if err := c.List(ctx, mpeList, client.InNamespace(namespace), client.MatchingLabelsSelector{Selector: sel}); err != nil {
		return nil, err
	}

	if len(mpeList.Items) == 1 {
		return ToRef(&mpeList.Items[0]), nil
	} else if len(mpeList.Items) > 1 {
		return nil, ErrSelectedMultiple
	}

	// ClusterMediaProcessingEntity
	cmpeList := &enginev1.ClusterMediaProcessingEntityList{}
	if err := c.List(ctx, cmpeList, client.MatchingLabelsSelector{Selector: sel}); err != nil {
		return nil, err
	}

	if len(cmpeList.Items) == 1 {
		return ToRef(&cmpeList.Items[0]), nil
	} else if len(cmpeList.Items) > 1 {
		return nil, ErrSelectedMultiple
	}

	return nil, nil
}

func SelectFunctionRef(ctx context.Context, c client.Client, namespace string, sel labels.Selector) (*meta.ObjectReference, error) {
	// local Function
	funcList := &enginev1.FunctionList{}
	if err := c.List(ctx, funcList, client.InNamespace(namespace), client.MatchingLabelsSelector{Selector: sel}); err != nil {
		return nil, err
	}

	if len(funcList.Items) == 1 {
		return ToRef(&funcList.Items[0]), nil
	} else if len(funcList.Items) > 1 {
		return nil, ErrSelectedMultiple
	}

	// ClusterFunction
	cfuncList := &enginev1.ClusterFunctionList{}
	if err := c.List(ctx, cfuncList, client.MatchingLabelsSelector{Selector: sel}); err != nil {
		return nil, err
	}

	if len(cfuncList.Items) == 1 {
		return ToRef(&cfuncList.Items[0]), nil
	} else if len(cfuncList.Items) > 1 {
		return nil, ErrSelectedMultiple
	}

	return nil, nil
}

func ToRef(obj client.Object) *meta.ObjectReference {
	return &meta.ObjectReference{
		APIVersion: obj.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		Kind:       obj.GetObjectKind().GroupVersionKind().Kind,
		Name:       obj.GetName(),
		Namespace:  obj.GetNamespace(),
	}
}

func ToExactRef(obj client.Object) *meta.ExactObjectReference {
	return &meta.ExactObjectReference{
		ObjectReference: *ToRef(obj),
		UID:             obj.GetUID(),
	}
}

func ToLocalRef(obj client.Object) *meta.LocalObjectReference {
	return &meta.LocalObjectReference{
		APIVersion: obj.GetObjectKind().GroupVersionKind().GroupVersion().String(),
		Kind:       obj.GetObjectKind().GroupVersionKind().Kind,
		Name:       obj.GetName(),
	}
}

func ResolveRef(ctx context.Context, c client.Client, ref *meta.ObjectReference) (client.Object, error) {
	lref := ref.LocalObjectReference()
	return ResolveLocalRef(ctx, c, ref.Namespace, &lref)
}

func ResolveExactRef(ctx context.Context, c client.Client, ref *meta.ExactObjectReference) (client.Object, error) {
	obj, err := ResolveRef(ctx, c, &ref.ObjectReference)
	if err != nil {
		return nil, err
	}
	if obj.GetUID() != ref.UID {
		gv, err := schema.ParseGroupVersion(ref.APIVersion)
		if err != nil {
			return nil, err
		}
		mapping, err := c.RESTMapper().RESTMapping(gv.WithKind(ref.Kind).GroupKind(), gv.Version)
		if err != nil {
			return nil, err
		}
		return nil, apierrors.NewConflict(mapping.Resource.GroupResource(), ref.Name, errors.New("reference has conflicting UID"))
	}
	return obj, nil
}

func ResolveLocalRef(ctx context.Context, c client.Client, namespace string, ref *meta.LocalObjectReference) (client.Object, error) {
	obj, err := PartiallyResolveLocalRef(c, namespace, ref)
	if err != nil {
		return nil, err
	}

	err = c.Get(ctx, client.ObjectKeyFromObject(obj), obj)
	return obj, err
}

func PartiallyResolveRef(c client.Client, ref *meta.ObjectReference) (client.Object, error) {
	lref := ref.LocalObjectReference()
	return PartiallyResolveLocalRef(c, ref.Namespace, &lref)
}

func PartiallyResolveLocalRef(c client.Client, namespace string, ref *meta.LocalObjectReference) (client.Object, error) {
	gv, err := schema.ParseGroupVersion(ref.APIVersion)
	if err != nil {
		return nil, err
	}
	if gv.Empty() {
		// assume we talk about nagare media engine types
		gv = enginev1.GroupVersion
	}
	gvk := gv.WithKind(ref.Kind)

	mapping, err := c.RESTMapper().RESTMapping(gvk.GroupKind(), gvk.Version)
	if err != nil {
		return nil, err
	}

	runObj, err := c.Scheme().New(gvk)
	if err != nil {
		return nil, err
	}

	obj, ok := runObj.(client.Object)
	if !ok {
		return nil, errors.New("failed to convert runtime.Object to client.Object")
	}

	obj.SetName(ref.Name)
	if mapping.Scope.Name() == apimeta.RESTScopeNameNamespace {
		obj.SetNamespace(namespace)
	}

	return obj, nil
}

func NormalizeRef(s *runtime.Scheme, ref *meta.ObjectReference, obj client.Object) error {
	gvks, _, err := s.ObjectKinds(obj)
	if err != nil {
		return err
	}
	ref.APIVersion = gvks[0].GroupVersion().String()
	ref.Kind = gvks[0].Kind
	return nil
}

func NormalizeExactRef(s *runtime.Scheme, ref *meta.ExactObjectReference, obj client.Object) error {
	return NormalizeRef(s, &ref.ObjectReference, obj)
}

func NormalizeLocalRef(s *runtime.Scheme, ref *meta.LocalObjectReference, obj client.Object) error {
	gvks, _, err := s.ObjectKinds(obj)
	if err != nil {
		return err
	}
	ref.APIVersion = gvks[0].GroupVersion().String()
	ref.Kind = gvks[0].Kind
	return nil
}

func NormalizeFunctionRef(s *runtime.Scheme, ref *meta.ObjectReference) error {
	var obj client.Object
	switch ref.Kind {
	case "Function":
		obj = &enginev1.Function{}
	case "ClusterFunction":
		obj = &enginev1.ClusterFunction{}
	default:
		return errors.New("Function reference does not reference a Function or ClusterFunction")
	}
	NormalizeRef(s, ref, obj)
	return nil
}

func NormalizeLocalFunctionRef(s *runtime.Scheme, ref *meta.LocalObjectReference) error {
	var obj client.Object
	switch ref.Kind {
	case "Function":
		obj = &enginev1.Function{}
	case "ClusterFunction":
		obj = &enginev1.ClusterFunction{}
	default:
		return errors.New("Function reference does not reference a Function or ClusterFunction")
	}
	NormalizeLocalRef(s, ref, obj)
	return nil
}

func NormalizeMediaProcessingEntityRef(s *runtime.Scheme, ref *meta.ObjectReference) error {
	var obj client.Object
	switch ref.Kind {
	case "MediaProcessingEntity":
		obj = &enginev1.MediaProcessingEntity{}
	case "ClusterMediaProcessingEntity":
		obj = &enginev1.ClusterMediaProcessingEntity{}
	default:
		return errors.New("MediaProcessingEntity reference does not reference a MediaProcessingEntity or ClusterMediaProcessingEntity")
	}
	NormalizeRef(s, ref, obj)
	return nil
}

func NormalizeLocalMediaProcessingEntityRef(s *runtime.Scheme, ref *meta.LocalObjectReference) error {
	var obj client.Object
	switch ref.Kind {
	case "MediaProcessingEntity":
		obj = &enginev1.MediaProcessingEntity{}
	case "ClusterMediaProcessingEntity":
		obj = &enginev1.ClusterMediaProcessingEntity{}
	default:
		return errors.New("MediaProcessingEntity reference does not reference a MediaProcessingEntity or ClusterMediaProcessingEntity")
	}
	NormalizeLocalRef(s, ref, obj)
	return nil
}

func NormalizeTaskTemplateRef(s *runtime.Scheme, ref *meta.ObjectReference) error {
	var obj client.Object
	switch ref.Kind {
	case "TaskTemplate":
		obj = &enginev1.TaskTemplate{}
	case "ClusterTaskTemplate":
		obj = &enginev1.ClusterTaskTemplate{}
	default:
		return errors.New("TaskTemplate reference does not reference a TaskTemplate or ClusterTaskTemplate")
	}
	NormalizeRef(s, ref, obj)
	return nil
}

func NormalizeLocalTaskTemplateRef(s *runtime.Scheme, ref *meta.LocalObjectReference) error {
	var obj client.Object
	switch ref.Kind {
	case "TaskTemplate":
		obj = &enginev1.TaskTemplate{}
	case "ClusterTaskTemplate":
		obj = &enginev1.ClusterTaskTemplate{}
	default:
		return errors.New("TaskTemplate reference does not reference a TaskTemplate or ClusterTaskTemplate")
	}
	NormalizeLocalRef(s, ref, obj)
	return nil
}

func LocalFunctionToObjectRef(lref *meta.LocalObjectReference, namespace string) (*meta.ObjectReference, error) {
	funcNamespace := ""
	switch lref.Kind {
	case "Function":
		funcNamespace = namespace
	case "ClusterFunction":
		// cluster scoped: does not have a namespace
	default:
		return nil, errors.New("Function reference does not reference a Function or ClusterFunction")
	}
	ref := lref.ObjectReference(funcNamespace)
	return &ref, nil
}

func LocalMediaProcessingEntityToObjectRef(lref *meta.LocalObjectReference, namespace string) (*meta.ObjectReference, error) {
	mpeNamespace := ""
	switch lref.Kind {
	case "MediaProcessingEntity":
		mpeNamespace = namespace
	case "ClusterMediaProcessingEntity":
		// cluster scoped: does not have a namespace
	default:
		return nil, errors.New("MediaProcessingEntity reference does not reference a MediaProcessingEntity or ClusterMediaProcessingEntity")
	}
	ref := lref.ObjectReference(mpeNamespace)
	return &ref, nil
}
