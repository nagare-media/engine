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
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var clusterfunctionlog = logf.Log.WithName("clusterfunction-resource")

func (r *ClusterFunction) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-engine-nagare-media-v1alpha1-clusterfunction,mutating=true,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=clusterfunctions,verbs=create;update,versions=v1alpha1,name=mclusterfunction.engine.nagare.media,admissionReviewVersions=v1

var _ webhook.Defaulter = &ClusterFunction{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ClusterFunction) Default() {
	clusterfunctionlog.V(1).Info("default", "name", r.Name)
}

//+kubebuilder:webhook:path=/validate-engine-nagare-media-v1alpha1-clusterfunction,mutating=false,failurePolicy=fail,sideEffects=None,groups=engine.nagare.media,resources=clusterfunctions,verbs=create;update,versions=v1alpha1,name=vclusterfunction.engine.nagare.media,admissionReviewVersions=v1

var _ webhook.Validator = &ClusterFunction{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterFunction) ValidateCreate() error {
	clusterfunctionlog.V(1).Info("validate create", "name", r.Name)
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterFunction) ValidateUpdate(old runtime.Object) error {
	clusterfunctionlog.V(1).Info("validate update", "name", r.Name)
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ClusterFunction) ValidateDelete() error {
	clusterfunctionlog.V(1).Info("validate delete", "name", r.Name)
	return nil
}
