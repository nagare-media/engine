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

package maps

// Copied from https://github.com/helm/helm/blob/44ec1c1575d405675272ee2d8131130a5decee7c/pkg/cli/values/options.go#L100-L117
func DeepMerge(a, b map[string]any) map[string]any {
	out := make(map[string]any, len(a))
	for k, v := range a {
		out[k] = v
	}
	for k, v := range b {
		if v, ok := v.(map[string]any); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]any); ok {
					out[k] = DeepMerge(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}

func Merge(maps ...map[string]string) map[string]string {
	if len(maps) == 0 {
		return nil
	}
	out := make(map[string]string, len(maps[0]))
	for _, m := range maps {
		for k, v := range m {
			out[k] = v
		}
	}
	return out
}
