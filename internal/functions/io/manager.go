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

package io

import (
	"context"

	"github.com/nagare-media/engine/pkg/starter"
)

type Manager interface {
	starter.Starter

	ManageCriticalStarter(starter.Starter)
	ManageStarter(starter.Starter)
	ManageStreamProcessor(StreamProcessor) error
	ManagePort(Port) error
}

type manager struct {
	cancelMgr *starter.Manager
	waitMgr   *starter.Manager
	srv       map[uint16]Server
}

var _ Manager = &manager{}

func NewManager() Manager {
	return &manager{
		srv:       make(map[uint16]Server),
		cancelMgr: starter.Manage(),
		waitMgr:   starter.Manage().WaitForAllToTerminate(),
	}
}

func (m *manager) ManageCriticalStarter(s starter.Starter) {
	m.cancelMgr.ManageStarter(s)
}

func (m *manager) ManageStarter(s starter.Starter) {
	m.waitMgr.ManageStarter(s)
}

func (m *manager) ManageStreamProcessor(sp StreamProcessor) error {
	m.ManageStarter(sp)
	return nil
}

func (m *manager) ManagePort(p Port) (err error) {
	if IsServerPort(p) {
		srv, ok := m.srv[p.PortNumber()]
		if !ok {
			srv, err = NewServerFor(p)
			if err != nil {
				return
			}
			m.srv[p.PortNumber()] = srv
				m.ManageCriticalStarter(srv)
		}

		if err = p.MountTo(srv); err != nil {
			return
		}
	}

	m.ManageStarter(p)
	return
}

func (m *manager) Start(ctx context.Context) error {
	s := make([]starter.Starter, 0, 2)
	if m.cancelMgr.Len() > 0 {
		s = append(s, m.cancelMgr)
	}
	if m.waitMgr.Len() > 0 {
		s = append(s, m.waitMgr)
	}
	return starter.Manage(s...).Start(ctx)
}
