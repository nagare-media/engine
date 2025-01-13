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

package modjson

import (
	"encoding/json"

	lua "github.com/yuin/gopher-lua"

	"github.com/nagare-media/engine/internal/functions/functions/script-lua/modules"
)

const Name = "json"

func Open(L *lua.LState) int {
	mod := L.RegisterModule(Name, modFuncs).(*lua.LTable)
	L.Push(mod)
	return 1
}

var modFuncs = map[string]lua.LGFunction{
	"parse": jsonParse,
}

func jsonParse(L *lua.LState) int {
	str := L.CheckString(1)

	var j any
	if err := json.Unmarshal([]byte(str), &j); err != nil {
		return modules.Error(L, err)
	}

	L.Push(lValue(L, j))
	return 1
}

func lValue(L *lua.LState, val any) lua.LValue {
	switch conv := val.(type) {
	case bool:
		return lua.LBool(conv)

	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		n, _ := conv.(float64)
		return lua.LNumber(float64(n))

	case string:
		return lua.LString(conv)

	case json.Number:
		return lua.LString(conv)

	case []any:
		arr := L.CreateTable(len(conv), 0)
		for _, v := range conv {
			arr.Append(lValue(L, v))
		}
		return arr

	case map[string]any:
		tbl := L.CreateTable(0, len(conv))
		for k, v := range conv {
			tbl.RawSetString(k, lValue(L, v))
		}
		return tbl
	}

	return lua.LNil
}
