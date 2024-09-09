/*
Copyright 2022-2024 The nagare media authors

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

package updatable

import "sync"

type updatableChannel[T any] struct {
	mtx  sync.RWMutex
	init chan struct{}
	vCh  chan VersionedValue[T]
	v    VersionedValue[T]        // protected by mtx
	subs []chan VersionedValue[T] // protected by mtx
}

var _ Updatable[interface{}] = &updatableChannel[interface{}]{}

func NewChannel[T any](bufSize ...int) Updatable[T] {
	bs := 0
	if len(bufSize) > 0 {
		bs = bufSize[0]
	}

	u := &updatableChannel[T]{
		init: make(chan struct{}),
		vCh:  make(chan VersionedValue[T], bs),
	}
	go u.start()
	return u
}

func (u *updatableChannel[T]) start() {
	initDone := sync.OnceFunc(func() { close(u.init) })

	nv, ok := <-u.vCh
	for ; ok; nv, ok = <-u.vCh {
		if u.v.Version == nv.Version {
			continue
		}

		u.mtx.Lock()
		u.v = nv
		initDone()
		for _, s := range u.subs {
			s <- u.v
		}
		u.mtx.Unlock()
	}
	// channel closed => terminate
}

func (u *updatableChannel[T]) Channel() chan<- VersionedValue[T] {
	return u.vCh
}

func (u *updatableChannel[T]) Get() VersionedValue[T] {
	<-u.init
	u.mtx.RLock()
	defer u.mtx.RUnlock()
	return u.v
}

func (u *updatableChannel[T]) Subscribe(bufSize ...int) (VersionedValue[T], <-chan VersionedValue[T]) {
	<-u.init
	u.mtx.Lock()
	defer u.mtx.Unlock()

	bs := 0
	if len(bufSize) > 0 {
		bs = bufSize[0]
	}

	s := make(chan VersionedValue[T], bs)
	u.subs = append(u.subs, s)
	return u.v, s
}

func (u *updatableChannel[T]) Close() {
	u.mtx.Lock()
	defer u.mtx.Unlock()

	close(u.vCh)
	for _, s := range u.subs {
		close(s)
	}
}
