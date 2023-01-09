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
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

func (r *Task) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-engine-nagare-media-v1alpha1-task,mutating=true,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=tasks,verbs=create;update,versions=v1alpha1,name=mtask.engine.nagare.media,admissionReviewVersions=v1

var _ webhook.Defaulter = &Task{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Task) Default() {
}

//+kubebuilder:webhook:path=/validate-engine-nagare-media-v1alpha1-task,mutating=false,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=tasks,verbs=create;update,versions=v1alpha1,name=vtask.engine.nagare.media,admissionReviewVersions=v1

var _ webhook.Validator = &Task{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Task) ValidateCreate() error {

	// check if workflow is already finished
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Task) ValidateUpdate(old runtime.Object) error {
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Task) ValidateDelete() error {
	return nil
}
