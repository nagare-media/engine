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
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

func ExtractStringParameterValue(p nbmpv2.Parameter) (string, bool) {
	if len(p.Values) != 1 {
		return "", false
	}

	if pVal, ok := p.Values[0].(*nbmpv2.StringParameterValue); ok {
		if len(pVal.Restrictions) != 1 {
			return "", false
		}
		return pVal.Restrictions[0], true
	}

	return "", false
}

// GetStringParameterValue returns the value of given key and a bool indicating if the value could be retrieved
// successfully. This assumes there is exactly one string value with exactly one restriction.
func GetStringParameterValue(parameters []nbmpv2.Parameter, key string) (string, bool) {
	for _, p := range parameters {
		if p.Name != key {
			continue
		}

		return ExtractStringParameterValue(p)
	}

	return "", false
}

func SetStringParameterValue(parameters []nbmpv2.Parameter, key, value string) []nbmpv2.Parameter {
	pnew := nbmpv2.Parameter{
		Name:     key,
		Datatype: nbmpv2.StringDatatype,
		Values: []nbmpv2.ParameterValue{&nbmpv2.StringParameterValue{
			Restrictions: []string{value},
		}},
	}

	for i, p := range parameters {
		if p.Name == key {
			parameters[i] = pnew
			return parameters
		}
	}

	return append(parameters, pnew)
}
