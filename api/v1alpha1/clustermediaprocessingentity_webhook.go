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

package v1alpha1

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (cmpe *ClusterMediaProcessingEntity) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(cmpe).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-engine-nagare-media-v1alpha1-clustermediaprocessingentity,mutating=true,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=clustermediaprocessingentities,verbs=create;update,versions=v1alpha1,name=mclustermediaprocessingentity.engine.nagare.media,admissionReviewVersions=v1

var _ webhook.Defaulter = &ClusterMediaProcessingEntity{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (cmpe *ClusterMediaProcessingEntity) Default() {
	(*MediaProcessingEntity)(cmpe).Default()
}

// +kubebuilder:webhook:path=/validate-engine-nagare-media-v1alpha1-clustermediaprocessingentity,mutating=false,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=clustermediaprocessingentities,verbs=create;update,versions=v1alpha1,name=vclustermediaprocessingentity.engine.nagare.media,admissionReviewVersions=v1

var _ webhook.Validator = &ClusterMediaProcessingEntity{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (cmpe *ClusterMediaProcessingEntity) ValidateCreate() (admission.Warnings, error) {
	return cmpe.validate(nil)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (cmpe *ClusterMediaProcessingEntity) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	oldCMPE, ok := old.(*ClusterMediaProcessingEntity)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a ClusterMediaProcessingEntity but got a %T", old))
	}
	return cmpe.validate(oldCMPE)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (cmpe *ClusterMediaProcessingEntity) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

func (cmpe *ClusterMediaProcessingEntity) validate(old *ClusterMediaProcessingEntity) (admission.Warnings, error) {
	return (*MediaProcessingEntity)(cmpe).validate((*MediaProcessingEntity)(old))
}
