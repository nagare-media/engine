/*
Copyright 2022 The nagare media authors

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
)

func (cml *ClusterMediaLocation) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(cml).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-engine-nagare-media-v1alpha1-clustermedialocation,mutating=true,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=clustermedialocations,verbs=create;update,versions=v1alpha1,name=mclustermedialocation.engine.nagare.media,admissionReviewVersions=v1

var _ webhook.Defaulter = &ClusterMediaLocation{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (cml *ClusterMediaLocation) Default() {
}

//+kubebuilder:webhook:path=/validate-engine-nagare-media-v1alpha1-clustermedialocation,mutating=false,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=clustermedialocations,verbs=create;update,versions=v1alpha1,name=vclustermedialocation.engine.nagare.media,admissionReviewVersions=v1

var _ webhook.Validator = &ClusterMediaLocation{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (cml *ClusterMediaLocation) ValidateCreate() error {
	return cml.validate(nil)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (cml *ClusterMediaLocation) ValidateUpdate(old runtime.Object) error {
	oldCML, ok := old.(*ClusterMediaLocation)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a ClusterMediaLocation but got a %T", old))
	}
	return cml.validate(oldCML)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (cml *ClusterMediaLocation) ValidateDelete() error {
	return nil
}

func (cml *ClusterMediaLocation) validate(old *ClusterMediaLocation) error {
	return nil
}
