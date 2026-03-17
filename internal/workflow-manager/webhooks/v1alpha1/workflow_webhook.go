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

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
)

func SetupWorkflowWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &enginev1.Workflow{}).
		WithDefaulter(&WorkflowCustomDefaulter{}).
		WithValidator(&WorkflowCustomValidator{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-engine-nagare-media-v1alpha1-workflow,mutating=true,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=workflows,verbs=create;update,versions=v1alpha1,name=mworkflow.engine.nagare.media,admissionReviewVersions=v1

type WorkflowCustomDefaulter struct{}

var _ admission.Defaulter[*enginev1.Workflow] = &WorkflowCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (d *WorkflowCustomDefaulter) Default(ctx context.Context, obj *enginev1.Workflow) error {
	// TODO: implement
	return nil
}

// +kubebuilder:webhook:path=/validate-engine-nagare-media-v1alpha1-workflow,mutating=false,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=workflows,verbs=create;update,versions=v1alpha1,name=vworkflow.engine.nagare.media,admissionReviewVersions=v1

type WorkflowCustomValidator struct{}

var _ admission.Validator[*enginev1.Workflow] = &WorkflowCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *WorkflowCustomValidator) ValidateCreate(ctx context.Context, obj *enginev1.Workflow) (admission.Warnings, error) {
	// TODO: implement
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *WorkflowCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *enginev1.Workflow) (admission.Warnings, error) {
	// TODO: implement
	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (v *WorkflowCustomValidator) ValidateDelete(ctx context.Context, obj *enginev1.Workflow) (admission.Warnings, error) {
	// TODO: implement
	return nil, nil
}
