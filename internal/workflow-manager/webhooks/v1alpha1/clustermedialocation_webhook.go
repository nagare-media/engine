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

func SetupClusterMediaLocationWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr, &enginev1.ClusterMediaLocation{}).
		WithDefaulter(&ClusterMediaLocationCustomDefaulter{}).
		WithValidator(&ClusterMediaLocationCustomValidator{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-engine-nagare-media-v1alpha1-clustermedialocation,mutating=true,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=clustermedialocations,verbs=create;update,versions=v1alpha1,name=mclustermedialocation.engine.nagare.media,admissionReviewVersions=v1

type ClusterMediaLocationCustomDefaulter struct{}

var _ admission.Defaulter[*enginev1.ClusterMediaLocation] = &ClusterMediaLocationCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (d *ClusterMediaLocationCustomDefaulter) Default(ctx context.Context, obj *enginev1.ClusterMediaLocation) error {
	return (*MediaLocationCustomDefaulter)(d).Default(ctx, (*enginev1.MediaLocation)(obj))
}

// +kubebuilder:webhook:path=/validate-engine-nagare-media-v1alpha1-clustermedialocation,mutating=false,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=clustermedialocations,verbs=create;update,versions=v1alpha1,name=vclustermedialocation.engine.nagare.media,admissionReviewVersions=v1

type ClusterMediaLocationCustomValidator struct{}

var _ admission.Validator[*enginev1.ClusterMediaLocation] = &ClusterMediaLocationCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *ClusterMediaLocationCustomValidator) ValidateCreate(ctx context.Context, obj *enginev1.ClusterMediaLocation) (admission.Warnings, error) {
	return (*MediaLocationCustomValidator)(v).ValidateCreate(ctx, (*enginev1.MediaLocation)(obj))
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *ClusterMediaLocationCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj *enginev1.ClusterMediaLocation) (admission.Warnings, error) {
	return (*MediaLocationCustomValidator)(v).ValidateUpdate(ctx, (*enginev1.MediaLocation)(oldObj), (*enginev1.MediaLocation)(newObj))
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (v *ClusterMediaLocationCustomValidator) ValidateDelete(ctx context.Context, obj *enginev1.ClusterMediaLocation) (admission.Warnings, error) {
	return (*MediaLocationCustomValidator)(v).ValidateDelete(ctx, (*enginev1.MediaLocation)(obj))
}
