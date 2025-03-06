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
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
)

func SetupTaskTemplateWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&enginev1.TaskTemplate{}).
		WithDefaulter(&TaskTemplateCustomDefaulter{}).
		WithValidator(&TaskTemplateCustomValidator{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-engine-nagare-media-v1alpha1-tasktemplate,mutating=true,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=tasktemplates,verbs=create;update,versions=v1alpha1,name=mtasktemplate.engine.nagare.media,admissionReviewVersions=v1

type TaskTemplateCustomDefaulter struct {
}

var _ webhook.CustomDefaulter = &TaskTemplateCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (d *TaskTemplateCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	_, ok := obj.(*enginev1.TaskTemplate)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a TaskTemplate but got a %T", obj))
	}
	// TODO: implement
	return nil
}

// +kubebuilder:webhook:path=/validate-engine-nagare-media-v1alpha1-tasktemplate,mutating=false,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=tasktemplates,verbs=create;update,versions=v1alpha1,name=vtasktemplate.engine.nagare.media,admissionReviewVersions=v1

type TaskTemplateCustomValidator struct {
}

var _ webhook.CustomValidator = &TaskTemplateCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *TaskTemplateCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	_, ok := obj.(*enginev1.TaskTemplate)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a TaskTemplate but got a %T", obj))
	}
	// TODO: implement
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *TaskTemplateCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	_, ok := oldObj.(*enginev1.TaskTemplate)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a TaskTemplate but got a %T", oldObj))
	}
	_, ok = newObj.(*enginev1.TaskTemplate)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a TaskTemplate but got a %T", newObj))
	}
	// TODO: implement
	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (v *TaskTemplateCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	_, ok := obj.(*enginev1.TaskTemplate)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a TaskTemplate but got a %T", obj))
	}
	// TODO: implement
	return nil, nil
}
