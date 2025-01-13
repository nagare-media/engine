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

func addTaskType(L *lua.LState, mod *lua.LTable) {
	lTask := modules.NewType(L, Task{}).(*lua.LUserData)
	lTaskMT := lTask.Metatable.(*lua.LTable)
	lTaskMT.RawSetString("new", L.NewFunction(taskNew))
	mod.RawSetString("Task", lTask)
}

type Task struct {
	ID                  string              `luar:"id" mapstructure:"id"`
	FunctionRestriction FunctionRestriction `luar:"function_restriction" mapstructure:"function_restriction"`
}

func taskNew(L *lua.LState) int {
	tsk := &Task{}
	err := modules.TypeConstructor(L, tsk)
	if err != nil {
		return modules.Error(L, err)
	}

	L.Push(luar.New(L, tsk))
	return 1
}

func (t *Task) String() string {
	s, _ := json.MarshalIndent(t, "", "  ")
	return string(s)
}

type FunctionRestriction struct {
	ID     string            `luar:"id" mapstructure:"id"`
	Config map[string]string `luar:"config" mapstructure:"config"`
}
