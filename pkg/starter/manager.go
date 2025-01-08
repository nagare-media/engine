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

package starter

import (
	"context"
	"sync"
)

type Manager struct {
	s []Starter

	waitForAllToTerminate bool
}

func Manage(s ...Starter) *Manager {
	return &Manager{
		s: s,
	}
}

func (m *Manager) ManageStarter(s Starter) {
	m.s = append(m.s, s)
}

func (m *Manager) WaitForAllToTerminate(b ...bool) *Manager {
	v := true
	if len(b) > 0 {
		v = b[0]
	}
	m.waitForAllToTerminate = v
	return m
}

func (m *Manager) Len() int {
	return len(m.s)
}

func (m *Manager) Start(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var err error
	setErr := sync.Once{}

	wg := sync.WaitGroup{}
	for _, s := range m.s {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errStarter := s.Start(ctx)
			if errStarter != nil {
				setErr.Do(func() { err = errStarter })
			}
			if errStarter != nil || !m.waitForAllToTerminate {
				cancel()
			}
		}()
	}

	wg.Wait()
	return err
}
