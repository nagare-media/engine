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

package modnbmp

import (
	"encoding/json"

	luar "github.com/nagare-media/engine/third_party/github.com/layeh/gopher-luar"
	lua "github.com/yuin/gopher-lua"

	"github.com/nagare-media/engine/internal/functions/functions/script-lua/modules"
)

func addConnectionType(L *lua.LState, mod *lua.LTable) {
	lConnection := modules.NewType(L, Connection{}).(*lua.LUserData)
	lConnectionMT := lConnection.Metatable.(*lua.LTable)
	lConnectionMT.RawSetString("new", L.NewFunction(connectionNew))
	mod.RawSetString("Connection", lConnection)
}

type Connection struct {
	ID   string `luar:"id" mapstructure:"id"`
	From Port   `luar:"from" mapstructure:"from"`
	To   Port   `luar:"to" mapstructure:"to"`
}

func connectionNew(L *lua.LState) int {
	con := &Connection{}
	err := modules.TypeConstructor(L, con)
	if err != nil {
		return modules.Error(L, err)
	}

	L.Push(luar.New(L, con))
	return 1
}

func (c *Connection) String() string {
	s, _ := json.MarshalIndent(c, "", "  ")
	return string(s)
}

type Port struct {
	Input  *Input  `luar:"input" mapstructure:"input"`
	Output *Output `luar:"output" mapstructure:"output"`
	Task   *Task   `luar:"task" mapstructure:"task"`
	Port   string  `luar:"port" mapstructure:"port"`
}
