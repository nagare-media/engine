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

func SetupClusterMediaProcessingEntityWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &enginev1.ClusterMediaProcessingEntity{}).
		WithDefaulter(&ClusterMediaProcessingEntityCustomDefaulter{}).
		WithValidator(&ClusterMediaProcessingEntityCustomValidator{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-engine-nagare-media-v1alpha1-clustermediaprocessingentity,mutating=true,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=clustermediaprocessingentities,verbs=create;update,versions=v1alpha1,name=mclustermediaprocessingentity.engine.nagare.media,admissionReviewVersions=v1

type ClusterMediaProcessingEntityCustomDefaulter struct{}

var _ admission.Defaulter[*enginev1.ClusterMediaProcessingEntity] = &ClusterMediaProcessingEntityCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (d *ClusterMediaProcessingEntityCustomDefaulter) Default(ctx context.Context, obj *enginev1.ClusterMediaProcessingEntity) error {
	return (*MediaProcessingEntityCustomDefaulter)(d).Default(ctx, (*enginev1.MediaProcessingEntity)(obj))
}

// +kubebuilder:webhook:path=/validate-engine-nagare-media-v1alpha1-clustermediaprocessingentity,mutating=false,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=clustermediaprocessingentities,verbs=create;update,versions=v1alpha1,name=vclustermediaprocessingentity.engine.nagare.media,admissionReviewVersions=v1

type ClusterMediaProcessingEntityCustomValidator struct{}

var _ admission.Validator[*enginev1.ClusterMediaProcessingEntity] = &ClusterMediaProcessingEntityCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *ClusterMediaProcessingEntityCustomValidator) ValidateCreate(ctx context.Context, obj *enginev1.ClusterMediaProcessingEntity) (admission.Warnings, error) {
	return (*MediaProcessingEntityCustomValidator)(v).ValidateCreate(ctx, (*enginev1.MediaProcessingEntity)(obj))
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *ClusterMediaProcessingEntityCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *enginev1.ClusterMediaProcessingEntity) (admission.Warnings, error) {
	return (*MediaProcessingEntityCustomValidator)(v).ValidateUpdate(ctx, (*enginev1.MediaProcessingEntity)(oldObj), (*enginev1.MediaProcessingEntity)(newObj))
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (v *ClusterMediaProcessingEntityCustomValidator) ValidateDelete(ctx context.Context, obj *enginev1.ClusterMediaProcessingEntity) (admission.Warnings, error) {
	return (*MediaProcessingEntityCustomValidator)(v).ValidateDelete(ctx, (*enginev1.MediaProcessingEntity)(obj))
}
