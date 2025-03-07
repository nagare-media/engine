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

func SetupClusterFunctionWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&enginev1.ClusterFunction{}).
		WithDefaulter(&ClusterFunctionCustomDefaulter{}).
		WithValidator(&ClusterFunctionCustomValidator{}).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-engine-nagare-media-v1alpha1-clusterfunction,mutating=true,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=clusterfunctions,verbs=create;update,versions=v1alpha1,name=mclusterfunction.engine.nagare.media,admissionReviewVersions=v1

type ClusterFunctionCustomDefaulter struct {
}

var _ webhook.CustomDefaulter = &ClusterFunctionCustomDefaulter{}

// Default implements webhook.CustomDefaulter so a webhook will be registered for the type
func (d *ClusterFunctionCustomDefaulter) Default(ctx context.Context, obj runtime.Object) error {
	o, ok := obj.(*enginev1.ClusterFunction)
	if !ok {
		return apierrors.NewBadRequest(fmt.Sprintf("expected a ClusterFunction but got a %T", obj))
	}
	return (*FunctionCustomDefaulter)(d).Default(ctx, (*enginev1.Function)(o))
}

// +kubebuilder:webhook:path=/validate-engine-nagare-media-v1alpha1-clusterfunction,mutating=false,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=clusterfunctions,verbs=create;update,versions=v1alpha1,name=vclusterfunction.engine.nagare.media,admissionReviewVersions=v1

type ClusterFunctionCustomValidator struct {
}

var _ webhook.CustomValidator = &ClusterFunctionCustomValidator{}

// ValidateCreate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *ClusterFunctionCustomValidator) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	o, ok := obj.(*enginev1.ClusterFunction)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a ClusterFunction but got a %T", obj))
	}
	return (*FunctionCustomValidator)(v).ValidateCreate(ctx, (*enginev1.Function)(o))
}

// ValidateUpdate implements webhook.CustomValidator so a webhook will be registered for the type
func (v *ClusterFunctionCustomValidator) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	oo, ok := oldObj.(*enginev1.ClusterFunction)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a ClusterFunction but got a %T", oldObj))
	}
	no, ok := newObj.(*enginev1.ClusterFunction)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a ClusterFunction but got a %T", newObj))
	}
	return (*FunctionCustomValidator)(v).ValidateUpdate(ctx, (*enginev1.Function)(oo), (*enginev1.Function)(no))
}

// ValidateDelete implements webhook.CustomValidator so a webhook will be registered for the type
func (v *ClusterFunctionCustomValidator) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	o, ok := obj.(*enginev1.ClusterFunction)
	if !ok {
		return nil, apierrors.NewBadRequest(fmt.Sprintf("expected a ClusterFunction but got a %T", obj))
	}
	return (*FunctionCustomValidator)(v).ValidateDelete(ctx, (*enginev1.Function)(o))
}
