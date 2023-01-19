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
)

func (cf *ClusterFunction) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(cf).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-engine-nagare-media-v1alpha1-clusterfunction,mutating=true,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=clusterfunctions,verbs=create;update,versions=v1alpha1,name=mclusterfunction.engine.nagare.media,admissionReviewVersions=v1

var _ webhook.Defaulter = &ClusterFunction{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (cf *ClusterFunction) Default() {
	(*Function)(cf).Default()
}

// +kubebuilder:webhook:path=/validate-engine-nagare-media-v1alpha1-clusterfunction,mutating=false,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=clusterfunctions,verbs=create;update,versions=v1alpha1,name=vclusterfunction.engine.nagare.media,admissionReviewVersions=v1

var _ webhook.Validator = &ClusterFunction{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (cf *ClusterFunction) ValidateCreate() error {
	return cf.validate(nil)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (cf *ClusterFunction) ValidateUpdate(old runtime.Object) error {
	oldCF, ok := old.(*ClusterFunction)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a ClusterFunction but got a %T", old))
	}
	return cf.validate(oldCF)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (cf *ClusterFunction) ValidateDelete() error {
	return nil
}

func (cf *ClusterFunction) validate(old *ClusterFunction) error {
	return (*Function)(cf).validate((*Function)(old))
}
