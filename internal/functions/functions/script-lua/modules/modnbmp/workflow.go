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
	"errors"

	luar "github.com/nagare-media/engine/third_party/github.com/layeh/gopher-luar"
	lua "github.com/yuin/gopher-lua"

	"github.com/nagare-media/engine/internal/functions/functions/script-lua/modules"
	nbmpconvv2 "github.com/nagare-media/engine/internal/pkg/nbmpconv/v2"
	nbmphttpv2 "github.com/nagare-media/engine/pkg/nbmp/http/v2"
	nbmputils "github.com/nagare-media/engine/pkg/nbmp/utils"
	nbmpv2 "github.com/nagare-media/models.go/iso/nbmp/v2"
)

func addWorkflowType(L *lua.LState, mod *lua.LTable) {
	lWorkflow := modules.NewType(L, Workflow{}).(*lua.LUserData)
	lWorkflowMT := lWorkflow.Metatable.(*lua.LTable)
	lWorkflowMT.RawSetString("new", L.NewFunction(workflowNew))
	lWorkflowMT.RawSetString("self", L.NewFunction(workflowSelf))
	mod.RawSetString("Workflow", lWorkflow)
}

type Workflow struct {
	Data *nbmpv2.Workflow `luar:"-" mapstructure:"-"`

	changed bool
}

func workflowNew(L *lua.LState) int {
	wf := &Workflow{
		Data: &nbmpv2.Workflow{},
	}

	err := modules.TypeConstructor(L, wf.Data)
	if err != nil {
		return modules.Error(L, err)
	}

	L.Push(luar.New(L, wf))
	return 1
}

func workflowSelf(L *lua.LState) int {
	ctx := L.Context()
	cfg := GetConfig(L)

	c := nbmphttpv2.WorkflowClient(cfg.WorkflowAPI)
	wf, err := c.Retrieve(ctx, cfg.WorkflowID)
	if err != nil {
		return modules.Error(L, err)
	}

	L.Push(luar.New(L, &Workflow{Data: wf}))
	return 1
}

func (w *Workflow) Update(L *luar.LState) int {
	if w.changed {
		ctx := L.Context()
		cfg := GetConfig(L.LState)

		c := nbmphttpv2.WorkflowClient(cfg.WorkflowAPI)
		wf, err := c.Update(ctx, w.Data)
		if err != nil {
			return modules.Error(L.LState, err)
		}

		w.Data = wf
		w.changed = false
	}
	return 0
}

func (w *Workflow) GetTask(L *luar.LState) int {
	id := L.CheckString(1)

	// search for task in connection map
	var instance string
	for _, con := range w.Data.Processing.ConnectionMap {
		if con.From.ID == id {
			instance = con.From.Instance
			break
		} else if con.To.ID == id {
			instance = con.To.Instance
			break
		}
	}
	if instance == "" {
		L.Push(lua.LNil)
		return 1
	}

	// search for function instance of task
	var fr *nbmpv2.FunctionRestriction
	if i := w.functionInstanceIndex(instance); i >= 0 {
		fr = &w.Data.Processing.FunctionRestrictions[i]
	} else {
		return modules.Error(L.LState, errors.New("illegal state: function instance does not exist"))
	}

	// convert configuration
	var (
		cfg map[string]string
		err error
	)
	if fr.Configuration != nil {
		cfg, err = nbmpconvv2.ParametersToMap(fr.Configuration.Parameters)
		if err != nil {
			return modules.Error(L.LState, err)
		}
	}

	tsk := &Task{
		ID: id,
		FunctionRestriction: FunctionRestriction{
			ID:     instance,
			Config: cfg,
		},
	}

	L.Push(luar.New(L.LState, tsk))
	return 1
}

func (w *Workflow) GetInput(L *luar.LState) int {
	streamID := L.CheckString(1)
	mp := nbmputils.FindMediaOrMetadataParameter(streamID, &w.Data.Input)
	if mp == nil {
		L.Push(lua.LNil)
		return 1
	}

	i := &Input{Data: mp}
	L.Push(luar.New(L.LState, i))
	return 1
}

func (w *Workflow) GetOutput(L *luar.LState) int {
	streamID := L.CheckString(1)
	mp := nbmputils.FindMediaOrMetadataParameter(streamID, &w.Data.Output)
	if mp == nil {
		L.Push(lua.LNil)
		return 1
	}

	o := &Output{Data: mp}
	L.Push(luar.New(L.LState, o))
	return 1
}

func (w *Workflow) AddConnection(L *luar.LState) int {
	lcon := L.CheckUserData(1)

	con, ok := lcon.Value.(*Connection)
	if !ok {
		L.ArgError(1, "not a Connection")
		return 0
	}

	// check if connection already exists
	if i := w.connectionIndex(con); i >= 0 {
		return 0
	}

	isWfInput := (con.From.Input != nil)
	isWfOutput := (con.To.Output != nil)

	// add connection
	cm := nbmpv2.ConnectionMapping{
		ConnectionID: con.ID,
	}
	if isWfInput {
		cm.From.ID = con.From.Input.Data.GetStreamID()
	} else {
		cm.From.ID = con.From.Task.ID
		cm.From.Instance = con.From.Task.ID
		cm.From.PortName = con.From.Port
	}
	if isWfOutput {
		cm.To.ID = con.To.Output.Data.GetStreamID()
	} else {
		cm.To.ID = con.To.Task.ID
		cm.To.Instance = con.To.Task.ID
		cm.To.PortName = con.To.Port
	}
	w.Data.Processing.ConnectionMap = append(w.Data.Processing.ConnectionMap, cm)
	w.changed = true

	// add function instances
	if !isWfInput {
		if i := w.functionInstanceIndex(con.From.Task.ID); i < 0 {
			fr := nbmpv2.FunctionRestriction{
				Instance: con.From.Task.ID,
				General: &nbmpv2.General{
					ID: con.From.Task.FunctionRestriction.ID,
				},
				Configuration: &nbmpv2.Configuration{
					Parameters: nbmpconvv2.MapToParameters(con.From.Task.FunctionRestriction.Config),
				},
			}
			w.Data.Processing.FunctionRestrictions = append(w.Data.Processing.FunctionRestrictions, fr)
		}
	}
	if !isWfOutput {
		if i := w.functionInstanceIndex(con.To.Task.ID); i < 0 {
			fr := nbmpv2.FunctionRestriction{
				Instance: con.To.Task.ID,
				General: &nbmpv2.General{
					ID: con.To.Task.FunctionRestriction.ID,
				},
				Configuration: &nbmpv2.Configuration{
					Parameters: nbmpconvv2.MapToParameters(con.To.Task.FunctionRestriction.Config),
				},
			}
			w.Data.Processing.FunctionRestrictions = append(w.Data.Processing.FunctionRestrictions, fr)
		}
	}

	return 0
}

func (w *Workflow) RemoveConnection(L *luar.LState) int {
	lcon := L.CheckUserData(1)

	con, ok := lcon.Value.(*Connection)
	if !ok {
		L.ArgError(1, "not a Connection")
		return 0
	}

	if i := w.connectionIndex(con); i >= 0 {
		w.Data.Processing.ConnectionMap = append(w.Data.Processing.ConnectionMap[:i], w.Data.Processing.ConnectionMap[i+1:]...)
		w.changed = true
	}

	return 0
}

func (w *Workflow) connectionIndex(con *Connection) int {
	isWfInput := (con.From.Input != nil)
	isWfOutput := (con.To.Output != nil)

	for i, cm := range w.Data.Processing.ConnectionMap {
		if
		// wf input
		(isWfInput && con.From.Input.Data.GetStreamID() != cm.From.ID) ||
			// task input
			(con.From.Task.ID != cm.From.ID) ||
			(con.From.Port != cm.From.PortName) ||
			// wf output
			(isWfOutput && con.To.Output.Data.GetStreamID() != cm.To.ID) ||
			// task output
			(con.To.Task.ID != cm.To.ID) ||
			(con.To.Port != cm.To.PortName) {
			continue
		}
		return i
	}

	return -1
}

func (w *Workflow) functionInstanceIndex(instance string) int {
	for i, fr := range w.Data.Processing.FunctionRestrictions {
		if fr.Instance == instance {
			return i
		}
	}
	return -1
}

func (w *Workflow) String() string {
	s, _ := json.MarshalIndent(w.Data, "", "  ")
	return string(s)
}

type Input struct {
	Data nbmpv2.MediaOrMetadataParameter `luar:"-" mapstructure:"-"`
}

func (i *Input) String() string {
	s, _ := json.MarshalIndent(i.Data, "", "  ")
	return string(s)
}

type Output struct {
	Data nbmpv2.MediaOrMetadataParameter `luar:"-" mapstructure:"-"`
}

func (o *Output) String() string {
	s, _ := json.MarshalIndent(o.Data, "", "  ")
	return string(s)
}
