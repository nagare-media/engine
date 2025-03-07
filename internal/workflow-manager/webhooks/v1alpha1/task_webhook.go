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

func SetupTaskWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&enginev1.Task{}).
		WithDefaulter(&TaskCustomDefaulter{}).
		WithValidator(&TaskCustomValidator{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-engine-nagare-media-v1alpha1-task,mutating=true,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=tasks,verbs=create;update,versions=v1alpha1,name=mtask.engine.nagare.media,admissionReviewVersions=v1

type TaskCustomDefaulter struct {
}

var _ webhook.CustomDefaulter = &TaskCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (d *TaskCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	_, ok := obj.(*enginev1.Task)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a Task but got a %T", obj))
	}
	// TODO: implement
	return nil
}

// +kubebuilder:webhook:path=/validate-engine-nagare-media-v1alpha1-task,mutating=false,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=tasks,verbs=create;update,versions=v1alpha1,name=vtask.engine.nagare.media,admissionReviewVersions=v1

type TaskCustomValidator struct {
}

var _ webhook.CustomValidator = &TaskCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *TaskCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	_, ok := obj.(*enginev1.Task)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a Task but got a %T", obj))
	}
	// TODO: implement
	return nil, nil
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *TaskCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	_, ok := oldObj.(*enginev1.Task)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a Task but got a %T", oldObj))
	}
	_, ok = newObj.(*enginev1.Task)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a Task but got a %T", newObj))
	}
	// TODO: implement
	return nil, nil
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (v *TaskCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	_, ok := obj.(*enginev1.Task)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a Task but got a %T", obj))
	}
	// TODO: implement
	return nil, nil
}
