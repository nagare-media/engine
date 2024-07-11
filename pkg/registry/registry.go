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

package registry

// Generic registry.
type registry[T any] struct {
	m map[string]T
}

// New returns a registry.
func New[T any]() *registry[T] {
	return &registry[T]{
		m: make(map[string]T),
	}
}

// Register obj.
func (r *registry[T]) Register(name string, t T) {
	r.m[name] = t
}

// Contains checks if obj by that name is registered.
func (r *registry[T]) Contains(name string) bool {
	_, ok := r.m[name]
	return ok
}

// Get obj with given name.
func (r *registry[T]) Get(name string) (T, bool) {
	t, ok := r.m[name]
	return t, ok
}
