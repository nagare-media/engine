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

func SetupClusterTaskTemplateWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &enginev1.ClusterTaskTemplate{}).
		WithDefaulter(&ClusterTaskTemplateCustomDefaulter{}).
		WithValidator(&ClusterTaskTemplateCustomValidator{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-engine-nagare-media-v1alpha1-clustertasktemplate,mutating=true,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=clustertasktemplates,verbs=create;update,versions=v1alpha1,name=mclustertasktemplate.engine.nagare.media,admissionReviewVersions=v1

type ClusterTaskTemplateCustomDefaulter struct{}

var _ admission.Defaulter[*enginev1.ClusterTaskTemplate] = &ClusterTaskTemplateCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (d *ClusterTaskTemplateCustomDefaulter) Default(ctx context.Context, obj *enginev1.ClusterTaskTemplate) error {
	return (*TaskTemplateCustomDefaulter)(d).Default(ctx, (*enginev1.TaskTemplate)(obj))
}

// +kubebuilder:webhook:path=/validate-engine-nagare-media-v1alpha1-clustertasktemplate,mutating=false,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=clustertasktemplates,verbs=create;update,versions=v1alpha1,name=vclustertasktemplate.engine.nagare.media,admissionReviewVersions=v1

type ClusterTaskTemplateCustomValidator struct{}

var _ admission.Validator[*enginev1.ClusterTaskTemplate] = &ClusterTaskTemplateCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *ClusterTaskTemplateCustomValidator) ValidateCreate(ctx context.Context, obj *enginev1.ClusterTaskTemplate) (admission.Warnings, error) {
	return (*TaskTemplateCustomValidator)(v).ValidateCreate(ctx, (*enginev1.TaskTemplate)(obj))
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *ClusterTaskTemplateCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *enginev1.ClusterTaskTemplate) (admission.Warnings, error) {
	return (*TaskTemplateCustomValidator)(v).ValidateUpdate(ctx, (*enginev1.TaskTemplate)(oldObj), (*enginev1.TaskTemplate)(newObj))
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (v *ClusterTaskTemplateCustomValidator) ValidateDelete(ctx context.Context, obj *enginev1.ClusterTaskTemplate) (admission.Warnings, error) {
	return (*TaskTemplateCustomValidator)(v).ValidateDelete(ctx, (*enginev1.TaskTemplate)(obj))
}
