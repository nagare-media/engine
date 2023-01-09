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

func (f *Function) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(f).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-engine-nagare-media-v1alpha1-function,mutating=true,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=functions,verbs=create;update,versions=v1alpha1,name=mfunction.engine.nagare.media,admissionReviewVersions=v1

var _ webhook.Defaulter = &Function{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (f *Function) Default() {
}

//+kubebuilder:webhook:path=/validate-engine-nagare-media-v1alpha1-function,mutating=false,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=functions,verbs=create;update,versions=v1alpha1,name=vfunction.engine.nagare.media,admissionReviewVersions=v1

var _ webhook.Validator = &Function{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (f *Function) ValidateCreate() error {
	return f.validate(nil)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (f *Function) ValidateUpdate(old runtime.Object) error {
	oldF, ok := old.(*Function)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a Function but got a %T", old))
	}
	return f.validate(oldF)
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (f *Function) ValidateDelete() error {
	return nil
}

func (f *Function) validate(old *Function) error {
	return nil
}
