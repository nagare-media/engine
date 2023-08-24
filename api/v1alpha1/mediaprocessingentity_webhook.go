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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (mpe *MediaProcessingEntity) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(mpe).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-engine-nagare-media-v1alpha1-mediaprocessingentity,mutating=true,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=mediaprocessingentities,verbs=create;update,versions=v1alpha1,name=mmediaprocessingentity.engine.nagare.media,admissionReviewVersions=v1

var _ webhook.Defaulter = &MediaProcessingEntity{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (mpe *MediaProcessingEntity) Default() {
}

// +kubebuilder:webhook:path=/validate-engine-nagare-media-v1alpha1-mediaprocessingentity,mutating=false,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=mediaprocessingentities,verbs=create;update,versions=v1alpha1,name=vmediaprocessingentity.engine.nagare.media,admissionReviewVersions=v1

var _ webhook.Validator = &MediaProcessingEntity{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (mpe *MediaProcessingEntity) ValidateCreate() (admission.Warnings, error) {
	return mpe.validate(nil)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (mpe *MediaProcessingEntity) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	oldMPE, ok := old.(*MediaProcessingEntity)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a MediaProcessingEntity but got a %T", old))
	}
	return mpe.validate(oldMPE)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (mpe *MediaProcessingEntity) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

func (mpe *MediaProcessingEntity) validate(old *MediaProcessingEntity) (admission.Warnings, error) {
	return nil, nil
}
