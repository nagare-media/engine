/*
Copyright 2022-2024 The nagare media authors

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

package predicate

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	kpredicate "sigs.k8s.io/controller-runtime/pkg/predicate"
)

var _ kpredicate.Predicate = HasLabels{}

// HasLabels defines a predicate that check if the object has all defined labels. The value is not checked.
type HasLabels []string

func (p HasLabels) Create(e event.CreateEvent) bool {
	return p.check(e.Object)
}

func (p HasLabels) Delete(e event.DeleteEvent) bool {
	return p.check(e.Object)
}

func (p HasLabels) Update(e event.UpdateEvent) bool {
	return p.check(e.ObjectNew) || p.check(e.ObjectOld)
}

func (p HasLabels) Generic(e event.GenericEvent) bool {
	return p.check(e.Object)
}

func (p HasLabels) check(obj client.Object) bool {
	labels := obj.GetLabels()
	for _, l := range p {
		if _, ok := labels[l]; !ok {
			return false
		}
	}
	return true
}
