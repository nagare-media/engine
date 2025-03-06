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

func SetupClusterMediaLocationWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&enginev1.ClusterMediaLocation{}).
		WithDefaulter(&ClusterMediaLocationCustomDefaulter{}).
		WithValidator(&ClusterMediaLocationCustomValidator{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-engine-nagare-media-v1alpha1-clustermedialocation,mutating=true,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=clustermedialocations,verbs=create;update,versions=v1alpha1,name=mclustermedialocation.engine.nagare.media,admissionReviewVersions=v1

type ClusterMediaLocationCustomDefaulter struct {
}

var _ webhook.CustomDefaulter = &ClusterMediaLocationCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (d *ClusterMediaLocationCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	o, ok := obj.(*enginev1.ClusterMediaLocation)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a ClusterMediaLocation but got a %T", obj))
	}
	return (*MediaLocationCustomDefaulter)(d).Default(ctx, (*enginev1.MediaLocation)(o))
}

// +kubebuilder:webhook:path=/validate-engine-nagare-media-v1alpha1-clustermedialocation,mutating=false,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=clustermedialocations,verbs=create;update,versions=v1alpha1,name=vclustermedialocation.engine.nagare.media,admissionReviewVersions=v1

type ClusterMediaLocationCustomValidator struct {
}

var _ webhook.CustomValidator = &ClusterMediaLocationCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *ClusterMediaLocationCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	o, ok := obj.(*enginev1.ClusterMediaLocation)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a ClusterMediaLocation but got a %T", obj))
	}
	return (*MediaLocationCustomValidator)(v).ValidateCreate(ctx, (*enginev1.MediaLocation)(o))
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *ClusterMediaLocationCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	oo, ok := oldObj.(*enginev1.ClusterMediaLocation)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a ClusterMediaLocation but got a %T", oldObj))
	}
	no, ok := newObj.(*enginev1.ClusterMediaLocation)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a ClusterMediaLocation but got a %T", newObj))
	}
	return (*MediaLocationCustomValidator)(v).ValidateUpdate(ctx, (*enginev1.MediaLocation)(oo), (*enginev1.MediaLocation)(no))
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (v *ClusterMediaLocationCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	o, ok := obj.(*enginev1.ClusterMediaLocation)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a ClusterMediaLocation but got a %T", obj))
	}
	return (*MediaLocationCustomValidator)(v).ValidateDelete(ctx, (*enginev1.MediaLocation)(o))
}
