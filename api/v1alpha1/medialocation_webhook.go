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

package v1alpha1

import (
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func (ml *MediaLocation) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(ml).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-engine-nagare-media-v1alpha1-medialocation,mutating=true,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=medialocations,verbs=create;update,versions=v1alpha1,name=mmedialocation.engine.nagare.media,admissionReviewVersions=v1

var _ webhook.Defaulter = &MediaLocation{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (ml *MediaLocation) Default() {
}

// +kubebuilder:webhook:path=/validate-engine-nagare-media-v1alpha1-medialocation,mutating=false,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=medialocations,verbs=create;update,versions=v1alpha1,name=vmedialocation.engine.nagare.media,admissionReviewVersions=v1

var _ webhook.Validator = &MediaLocation{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (ml *MediaLocation) ValidateCreate() error {
	return ml.validate(nil)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (ml *MediaLocation) ValidateUpdate(old runtime.Object) error {
	oldML, ok := old.(*MediaLocation)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a MediaLocation but got a %T", old))
	}
	return ml.validate(oldML)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (ml *MediaLocation) ValidateDelete() error {
	return nil
}

func (ml *MediaLocation) validate(old *MediaLocation) error {
	var allErrs field.ErrorList
	specPath := field.NewPath("spec")

	if ml.Spec.RIST != nil {
		if ml.Spec.RIST.BufferSize != nil {
			if ml.Spec.RIST.BufferSize.Duration < 0 {
				allErrs = append(
					allErrs,
					field.Invalid(
						specPath.Child("rist", "bufferSize"),
						ml.Spec.RIST.BufferSize,
						"must not be negative",
					),
				)
			}
			if ml.Spec.RIST.BufferSize.Duration > 30*time.Second {
				allErrs = append(
					allErrs,
					field.Invalid(
						specPath.Child("rist", "bufferSize"),
						ml.Spec.RIST.BufferSize,
						"must not be larger than 30s",
					),
				)
			}
		}
	}

	if len(allErrs) == 0 {
		return nil
	}
	return apierrors.NewInvalid(GroupVersion.WithKind("MediaLocation").GroupKind(), ml.Name, allErrs)
}
