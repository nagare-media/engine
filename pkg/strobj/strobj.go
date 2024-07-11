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

// +kubebuilder:object:generate=true
package strobj

import (
	"encoding/json"
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

type StringOrObject struct {
	Type      Type
	StrVal    string
	ObjectVal apiextensionsv1.JSON
}

type Type int

const (
	String Type = iota // The StringOrObject holds a string.
	Object             // The StringOrObject holds an object.
)

// FromString creates an StringOrObject object with a string value.
func FromString(val string) StringOrObject {
	return StringOrObject{Type: String, StrVal: val}
}

// FromObject creates an StringOrObject object with an object value.
func FromObject(val apiextensionsv1.JSON) StringOrObject {
	return StringOrObject{Type: Object, ObjectVal: val}
}

// UnmarshalJSON implements the json.Unmarshaller interface.
func (strobj *StringOrObject) UnmarshalJSON(value []byte) error {
	if value[0] == '"' {
		strobj.Type = String
		return json.Unmarshal(value, &strobj.StrVal)
	}
	strobj.Type = Object
	return json.Unmarshal(value, &strobj.ObjectVal)
}

// String returns the string value, or the raw JSON object value.
func (strobj *StringOrObject) String() string {
	if strobj == nil {
		return "null"
	}
	s, _ := strobj.MarshalJSON()
	return string(s)
}

// MarshalJSON implements the json.Marshaller interface.
func (strobj StringOrObject) MarshalJSON() ([]byte, error) {
	switch strobj.Type {
	case String:
		return json.Marshal(strobj.StrVal)
	case Object:
		return strobj.ObjectVal.Raw, nil
	default:
		return []byte{}, fmt.Errorf("strobj: impossible StringOrObject.Type")
	}
}
