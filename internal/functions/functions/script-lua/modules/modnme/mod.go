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

package modnme

import (
	"errors"

	"github.com/nagare-media/engine/internal/functions/functions/script-lua/modules"
	luar "github.com/nagare-media/engine/third_party/github.com/layeh/gopher-luar"
	lua "github.com/yuin/gopher-lua"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const Name = "nme"

func Open(L *lua.LState) int {
	mod := L.RegisterModule(Name, modFuncs).(*lua.LTable)
	addPortType(L, mod)
	L.Push(mod)
	return 1
}

var modFuncs = map[string]lua.LGFunction{
	"log":             nmeLog,
	"logerr":          nmeLogerr,
	"get_input_port":  nmeGetInputPort,
	"get_output_port": nmeGetOutputPort,
}

func nmeLog(L *lua.LState) int {
	l := log.FromContext(L.Context())
	msg := L.ToStringMeta(L.Get(1)).String()
	kv := []any{}
	for i := 2; i <= L.GetTop(); i++ {
		kv = append(kv, L.ToStringMeta(L.Get(i)).String())
	}
	if len(kv)%2 == 1 {
		return modules.Error(L, errors.New("odd number of key-value pairs"))
	}
	l.Info(msg, kv...)
	return 0
}

func nmeLogerr(L *lua.LState) int {
	l := log.FromContext(L.Context())
	err := L.ToStringMeta(L.Get(1)).String()
	msg := L.ToStringMeta(L.Get(2)).String()
	kv := []any{}
	for i := 3; i <= L.GetTop(); i++ {
		kv = append(kv, L.ToStringMeta(L.Get(i)).String())
	}
	if len(kv)%2 == 1 {
		return modules.Error(L, errors.New("odd number of key-value pairs"))
	}
	l.Error(errors.New(err), msg, kv...)
	return 0
}

func nmeGetInputPort(L *lua.LState) int {
	cfg := GetConfig(L)
	streamID := L.CheckString(1)

	reader := cfg.InputStreams[streamID]
	if reader == nil {
		L.Push(lua.LNil)
		return 1
	}

	s := &Port{in: reader}
	L.Push(luar.New(L, s))
	return 1
}

func nmeGetOutputPort(L *lua.LState) int {
	cfg := GetConfig(L)
	streamID := L.CheckString(1)

	writer := cfg.OutputStreams[streamID]
	if writer == nil {
		L.Push(lua.LNil)
		return 1
	}

	s := &Port{out: writer}
	L.Push(luar.New(L, s))
	return 1
}
