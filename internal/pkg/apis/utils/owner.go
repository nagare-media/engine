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

package utils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Copied from https://github.com/kubernetes-sigs/cluster-api/blob/314d98e0141489e2ac7a2915c66628ad2f22d006/util/util.go

// EnsureOwnerRef makes sure the slice contains the OwnerReference.
func EnsureOwnerRef(ownerReferences []metav1.OwnerReference, ref metav1.OwnerReference) []metav1.OwnerReference {
	idx := indexOwnerRef(ownerReferences, ref)
	if idx == -1 {
		return append(ownerReferences, ref)
	}
	ownerReferences[idx] = ref
	return ownerReferences
}

// indexOwnerRef returns the index of the owner reference in the slice if found, or -1.
func indexOwnerRef(ownerReferences []metav1.OwnerReference, ref metav1.OwnerReference) int {
	for index, r := range ownerReferences {
		if referSameObject(r, ref) {
			return index
		}
	}
	return -1
}

// Returns true if a and b point to the same object.
func referSameObject(a, b metav1.OwnerReference) bool {
	// UIDs must match if set
	if a.UID != "" && b.UID != "" {
		return a.UID == b.UID
	}

	// ignore version as different version still refer to the same object
	// TODO: is is a fair assumption?

	aGV, err := schema.ParseGroupVersion(a.APIVersion)
	if err != nil {
		return false
	}

	bGV, err := schema.ParseGroupVersion(b.APIVersion)
	if err != nil {
		return false
	}

	return aGV.Group == bGV.Group && a.Kind == b.Kind && a.Name == b.Name
}
