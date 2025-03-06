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

package v1alpha1

import (
	"context"
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
)

func SetupFunctionWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&enginev1.Function{}).
		WithDefaulter(&FunctionCustomDefaulter{}).
		WithValidator(&FunctionCustomValidator{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-engine-nagare-media-v1alpha1-function,mutating=true,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=functions,verbs=create;update,versions=v1alpha1,name=mfunction.engine.nagare.media,admissionReviewVersions=v1

type FunctionCustomDefaulter struct {
}

var _ webhook.CustomDefaulter = &FunctionCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (d *FunctionCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	_, ok := obj.(*enginev1.Function)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a Function but got a %T", obj))
	}
	// TODO: implement
	return nil
}

// +kubebuilder:webhook:path=/validate-engine-nagare-media-v1alpha1-function,mutating=false,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=functions,verbs=create;update,versions=v1alpha1,name=vfunction.engine.nagare.media,admissionReviewVersions=v1

type FunctionCustomValidator struct {
}

var _ webhook.CustomValidator = &FunctionCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *FunctionCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	o, ok := obj.(*enginev1.Function)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a Function but got a %T", obj))
	}
	return v.validate(ctx, o, nil)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *FunctionCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	oo, ok := oldObj.(*enginev1.Function)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a Function but got a %T", oldObj))
	}
	no, ok := newObj.(*enginev1.Function)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a Function but got a %T", newObj))
	}
	return v.validate(ctx, no, oo)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (v *FunctionCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	_, ok := obj.(*enginev1.Function)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a Function but got a %T", obj))
	}
	// TODO: implement
	return nil, nil
}

func (v *FunctionCustomValidator) validate(_ context.Context, obj, oldObj *enginev1.Function) (admission.Warnings, error) {
	var allErrs field.ErrorList
	specPath := field.NewPath("spec")

	if oldObj != nil {
		if obj.Spec.Version != oldObj.Spec.Version {
			allErrs = append(
				allErrs,
				field.Forbidden(specPath.Child("version"), "field is immutable"),
			)
		}
	}

	if len(allErrs) == 0 {
		return nil, nil
	}
	return nil, apierrors.NewInvalid(enginev1.GroupVersion.WithKind("Function").GroupKind(), obj.Name, allErrs)
}
