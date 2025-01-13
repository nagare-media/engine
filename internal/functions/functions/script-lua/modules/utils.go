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

package modules

import (
	"github.com/go-viper/mapstructure/v2"
	luar "github.com/nagare-media/engine/third_party/github.com/layeh/gopher-luar"
	lua "github.com/yuin/gopher-lua"
)

func NewType(L *lua.LState, value any) lua.LValue {
	ud := luar.NewType(L, value).(*lua.LUserData)
	// luar has a single shared metatable for all types. If we want to
	// add new methods we need to have distinct metatables per type.
	luarMT := ud.Metatable.(*lua.LTable)
	mt := L.CreateTable(0, 4)
	mt.RawSetString("__call", luarMT.RawGetString("__call"))
	mt.RawSetString("__eq", luarMT.RawGetString("__eq"))
	mt.RawSetString("__metatable", luarMT.RawGetString("__metatable"))
	mt.RawSetString("__index", mt)
	ud.Metatable = mt
	return ud
}

func TypeConstructor(L *lua.LState, t any) error {
	// arg 1 => self  (LUserData)
	// arg 2 => table (LTable)

	ltbl := L.OptTable(2, L.CreateTable(0, 0))
	tbl, err := luar.ConvertTo(L, ltbl, map[interface{}]interface{}{})
	if err != nil {
		return err
	}

	err = mapstructure.Decode(tbl, t)
	if err != nil {
		return err
	}

	return nil
}

func Error(L *lua.LState, err error) int {
	L.Error(lua.LString(err.Error()), 1)
	return 0
}
