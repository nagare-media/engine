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

package tplfuncs

import (
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/nagare-media/engine/pkg/maps"
)

var defaultFuncMap = template.FuncMap{}

func init() {
	defaultFuncMap = maps.DeepMerge(defaultFuncMap, sprig.FuncMap())
	defaultFuncMap = maps.DeepMerge(defaultFuncMap, nagareFuncMap)
}

func DefaultFuncMap() template.FuncMap {
	return defaultFuncMap
}
