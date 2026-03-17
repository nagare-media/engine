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
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
)

func SetupMediaLocationWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &enginev1.MediaLocation{}).
		WithDefaulter(&MediaLocationCustomDefaulter{}).
		WithValidator(&MediaLocationCustomValidator{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-engine-nagare-media-v1alpha1-medialocation,mutating=true,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=medialocations,verbs=create;update,versions=v1alpha1,name=mmedialocation.engine.nagare.media,admissionReviewVersions=v1

type MediaLocationCustomDefaulter struct{}

var _ admission.Defaulter[*enginev1.MediaLocation] = &MediaLocationCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (d *MediaLocationCustomDefaulter) Default(ctx context.Context, obj *enginev1.MediaLocation) error {
	// TODO: implement
	return nil
}

// +kubebuilder:webhook:path=/validate-engine-nagare-media-v1alpha1-medialocation,mutating=false,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=medialocations,verbs=create;update,versions=v1alpha1,name=vmedialocation.engine.nagare.media,admissionReviewVersions=v1

type MediaLocationCustomValidator struct{}

var _ admission.Validator[*enginev1.MediaLocation] = &MediaLocationCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *MediaLocationCustomValidator) ValidateCreate(ctx context.Context, obj *enginev1.MediaLocation) (admission.Warnings, error) {
	return v.validate(ctx, obj, nil)
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *MediaLocationCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *enginev1.MediaLocation) (admission.Warnings, error) {
	return v.validate(ctx, newObj, oldObj)
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (v *MediaLocationCustomValidator) ValidateDelete(ctx context.Context, obj *enginev1.MediaLocation) (admission.Warnings, error) {
	// TODO: implement
	return nil, nil
}

func (v *MediaLocationCustomValidator) validate(_ context.Context, obj, _ *enginev1.MediaLocation) (admission.Warnings, error) {
	var allErrs field.ErrorList
	specPath := field.NewPath("spec")

	if obj.Spec.RIST != nil {
		if obj.Spec.RIST.BufferSize != nil {
			if obj.Spec.RIST.BufferSize.Duration < 0 {
				allErrs = append(
					allErrs,
					field.Invalid(
						specPath.Child("rist", "bufferSize"),
						obj.Spec.RIST.BufferSize,
						"must not be negative",
					),
				)
			}
			if obj.Spec.RIST.BufferSize.Duration > 30*time.Second {
				allErrs = append(
					allErrs,
					field.Invalid(
						specPath.Child("rist", "bufferSize"),
						obj.Spec.RIST.BufferSize,
						"must not be larger than 30s",
					),
				)
			}
		}
	}

	if len(allErrs) == 0 {
		return nil, nil
	}
	return nil, apierrors.NewInvalid(enginev1.GroupVersion.WithKind("MediaLocation").GroupKind(), obj.Name, allErrs)
}
