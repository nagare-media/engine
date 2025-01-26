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

package function

import (
	"context"
	"fmt"
	goio "io"
	"reflect"

	"github.com/stoewer/go-strcase"
	lua "github.com/yuin/gopher-lua"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/nagare-media/engine/internal/functions"
	"github.com/nagare-media/engine/internal/functions/functions/script-lua/modules/modjson"
	"github.com/nagare-media/engine/internal/functions/functions/script-lua/modules/modnbmp"
	"github.com/nagare-media/engine/internal/functions/functions/script-lua/modules/modnme"
	"github.com/nagare-media/engine/internal/functions/io"
	"github.com/nagare-media/engine/pkg/nbmp"
	nbmputils "github.com/nagare-media/engine/pkg/nbmp/utils"
	"github.com/nagare-media/engine/pkg/starter"
	luar "github.com/nagare-media/engine/third_party/github.com/layeh/gopher-luar"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

// Function description
const (
	Name = "script-lua"

	ScriptParameterKey      = Name + ".engine.nagare.media/script"
	WorkflowAPIParameterKey = Name + ".engine.nagare.media/workflow-api"
)

// function executes Lua script.
type function struct {
	tsk *nbmpv2.Task

	script      string
	workflowAPI string
	workflowID  string

	inputStreams  map[string]goio.ReadCloser
	outputStreams map[string]goio.WriteCloser
}

var _ nbmp.Function = &function{}

// Exec script-lua function.
func (f *function) Exec(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	l := log.FromContext(ctx).WithName(Name)
	ctx = log.IntoContext(ctx, l)

	mgr := io.NewManager()

	// input ports
	f.inputStreams = make(map[string]goio.ReadCloser, len(f.tsk.General.InputPorts))
	for _, p := range f.tsk.General.InputPorts {
		port, err := io.NewInputPortForInputs(p, &f.tsk.Input)
		if err != nil {
			return err
		}

		pr, pw := goio.Pipe()
		f.inputStreams[port.Name()] = pr
		port.Connect(io.NewConnection(pw))

		if err := mgr.ManagePort(port, false); err != nil {
			return err
		}
	}

	// output ports
	f.outputStreams = make(map[string]goio.WriteCloser, len(f.tsk.General.OutputPorts))
	for _, p := range f.tsk.General.OutputPorts {
		port, err := io.NewOutputPortForOutputs(p, &f.tsk.Output)
		if err != nil {
			return err
		}

		f.outputStreams[port.Name()] = port

		if err := mgr.ManagePort(port, false); err != nil {
			return err
		}
	}

	// lua script
	mgr.ManageCriticalStarter(starter.Func(f.execLua))

	return mgr.Start(ctx)
}

func (f *function) execLua(ctx context.Context) error {
	l := log.FromContext(ctx)

	L := lua.NewState(lua.Options{
		SkipOpenLibs: true,
	})
	defer L.Close()
	L.SetContext(log.IntoContext(ctx, l.WithName("lua")))

	// configure Lua
	if err := f.configureLua(L); err != nil {
		l.Error(err, "Lua configuration failed")
		return err
	}

	// load modules
	if err := f.openModules(L); err != nil {
		l.Error(err, "opening modules failed")
		return err
	}

	// configure modules
	if err := f.configureLuaModules(L); err != nil {
		l.Error(err, "Lua configuration failed")
		return err
	}

	if err := L.DoString(f.script); err != nil {
		l.Error(err, "script execution failed")
		return err
	}

	return nil
}

func (f *function) configureLua(L *lua.LState) error {
	luarCfg := luar.GetConfig(L)

	luarCfg.FieldNames = func(s reflect.Type, f reflect.StructField) []string {
		const tagName = "luar"
		tag := f.Tag.Get(tagName)
		if tag == "-" {
			return nil
		}
		if tag != "" {
			return []string{tag}
		}
		return []string{strcase.SnakeCase(f.Name)}
	}

	luarCfg.MethodNames = func(t reflect.Type, m reflect.Method) []string {
		return []string{strcase.SnakeCase(m.Name)}
	}

	return nil
}

func (f *function) configureLuaModules(L *lua.LState) error {
	// mod_nbmp
	modnbmp.SetConfig(L, &modnbmp.Config{
		WorkflowAPI: f.workflowAPI,
		WorkflowID:  f.workflowID,
	})

	// mod_nme
	modnme.SetConfig(L, &modnme.Config{
		InputStreams:  f.inputStreams,
		OutputStreams: f.outputStreams,
	})

	return nil
}

func (f *function) openModules(L *lua.LState) error {
	type luaModule struct {
		name string
		fn   lua.LGFunction
	}

	luaModules := []luaModule{
		// stdlib
		// "io" not loaded
		// "os" not loaded
		{lua.LoadLibName, lua.OpenPackage}, // must be first
		{lua.BaseLibName, lua.OpenBase},
		// TODO: remove "dofile" function
		// TODO: remove "loadfile" function
		{lua.TabLibName, lua.OpenTable},
		{lua.StringLibName, lua.OpenString},
		{lua.MathLibName, lua.OpenMath},
		{lua.DebugLibName, lua.OpenDebug},
		{lua.ChannelLibName, lua.OpenChannel},
		{lua.CoroutineLibName, lua.OpenCoroutine},

		// libnme
		{modnbmp.Name, modnbmp.Open},
		{modnme.Name, modnme.Open},
		{modjson.Name, modjson.Open},
	}

	for _, mod := range luaModules {
		err := L.CallByParam(
			lua.P{
				Fn:      L.NewFunction(mod.fn),
				NRet:    0,
				Protect: true,
			},
			lua.LString(mod.name),
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// BuildTask from script-lua function.
func BuildTask(ctx context.Context, t *nbmpv2.Task) (nbmp.Function, error) {
	f := &function{
		tsk: t,
	}

	// config: script
	script, ok := nbmputils.GetStringParameterValue(t.Configuration.Parameters, ScriptParameterKey)
	if !ok {
		return nil, fmt.Errorf("missing %s parameter", ScriptParameterKey)
	}
	f.script = script

	// config: workflowAPI
	workflowAPI, ok := nbmputils.GetStringParameterValue(t.Configuration.Parameters, WorkflowAPIParameterKey)
	if !ok {
		return nil, fmt.Errorf("missing %s parameter", WorkflowAPIParameterKey)
	}
	f.workflowAPI = workflowAPI

	// config: workflowID
	workflowID, ok := nbmputils.GetStringParameterValue(t.Configuration.Parameters, nbmp.EngineWorkflowIDParameterKey)
	if !ok {
		return nil, fmt.Errorf("missing %s parameter", nbmp.EngineWorkflowIDParameterKey)
	}
	f.workflowID = workflowID

	return f, nil
}

func init() {
	functions.TaskBuilders.Register(Name, BuildTask)
}
