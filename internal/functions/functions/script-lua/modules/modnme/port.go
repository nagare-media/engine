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
	"bufio"
	"io"

	luar "github.com/nagare-media/engine/third_party/github.com/layeh/gopher-luar"
	lua "github.com/yuin/gopher-lua"

	"github.com/nagare-media/engine/internal/functions/functions/script-lua/modules"
)

func addPortType(L *lua.LState, mod *lua.LTable) {
	lPort := modules.NewType(L, Port{}).(*lua.LUserData)
	mod.RawSetString("Port", lPort)
}

type Port struct {
	// input port
	in      io.ReadCloser
	scanner *bufio.Scanner

	// output port
	// TODO: implement
	out io.WriteCloser
}

func (p *Port) ReadLine(L *luar.LState) int {
	if p.scanner == nil {
		p.scanner = bufio.NewScanner(p.in)
	}

	if p.scanner.Scan() {
		L.Push(lua.LString(p.scanner.Text()))
		return 1
	}

	L.Push(lua.LNil)
	return 1
}

func (p *Port) ReadAll(L *luar.LState) int {
	buf, err := io.ReadAll(p.in)
	if err != nil {
		modules.Error(L.LState, err)
	}

	L.Push(lua.LString(buf))
	return 1
}

func (p *Port) WriteLine(L *luar.LState) int {
	s := L.ToStringMeta(L.Get(1)).String()
	buf := make([]byte, len(s)+1)
	copy(buf, s)
	buf[len(s)] = '\n'

	_, err := p.out.Write([]byte(buf))
	if err != nil {
		return modules.Error(L.LState, err)
	}

	return 0
}

func (p *Port) Close(L *luar.LState) int {
	var err error
	if p.in != nil {
		err = p.in.Close()
	} else {
		err = p.out.Close()
	}
	if err != nil {
		return modules.Error(L.LState, err)
	}
	return 0
}
