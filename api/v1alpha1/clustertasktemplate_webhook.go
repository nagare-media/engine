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

func (ctt *ClusterTaskTemplate) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(ctt).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-engine-nagare-media-v1alpha1-clustertasktemplate,mutating=true,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=clustertasktemplates,verbs=create;update,versions=v1alpha1,name=mclustertasktemplate.engine.nagare.media,admissionReviewVersions=v1

var _ webhook.Defaulter = &ClusterTaskTemplate{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (ctt *ClusterTaskTemplate) Default() {
	(*TaskTemplate)(ctt).Default()
}

// +kubebuilder:webhook:path=/validate-engine-nagare-media-v1alpha1-clustertasktemplate,mutating=false,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=clustertasktemplates,verbs=create;update,versions=v1alpha1,name=vclustertasktemplate.engine.nagare.media,admissionReviewVersions=v1

var _ webhook.Validator = &ClusterTaskTemplate{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (ctt *ClusterTaskTemplate) ValidateCreate() (admission.Warnings, error) {
	return ctt.validate(nil)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (ctt *ClusterTaskTemplate) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	oldCTT, ok := old.(*ClusterTaskTemplate)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a ClusterTaskTemplate but got a %T", old))
	}
	return ctt.validate(oldCTT)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (ctt *ClusterTaskTemplate) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

func (ctt *ClusterTaskTemplate) validate(old *ClusterTaskTemplate) (admission.Warnings, error) {
	return (*TaskTemplate)(ctt).validate((*TaskTemplate)(old))
}
