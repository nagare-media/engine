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

// Package v1alpha1 contains API Schema definitions for the  v1alpha1 API group
// +kubebuilder:object:generate=true
// +groupName=engine.nagare.media
// +kubebuilder:ac:generate=true
// +kubebuilder:ac:output:package="applyconfiguration"
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	GroupName = "engine.nagare.media"
	Version   = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme adds the types in this group-version to the given scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&ClusterFunction{},
		&ClusterFunctionList{},
		&ClusterMediaLocation{},
		&ClusterMediaLocationList{},
		&ClusterMediaProcessingEntity{},
		&ClusterMediaProcessingEntityList{},
		&ClusterTaskTemplate{},
		&ClusterTaskTemplateList{},
		&GatewayNBMPConfig{},
		&TaskShimConfig{},
		&WorkflowManagerConfig{},
		&WorkflowManagerHelperConfig{},
		&WorkflowManagerHelperData{},
		&Function{},
		&FunctionList{},
		&MediaLocation{},
		&MediaLocationList{},
		&MediaProcessingEntity{},
		&MediaProcessingEntityList{},
		&Task{},
		&TaskList{},
		&TaskTemplate{},
		&TaskTemplateList{},
		&Workflow{},
		&WorkflowList{},
	)

	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

// Kind takes an unqualified kind and returns a Group qualified GroupKind
func Kind(resource string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(resource).GroupKind()
}
