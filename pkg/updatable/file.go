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

import (
	"fmt"

	"github.com/fsnotify/fsnotify"

	"crypto/sha256"
	"os"
	"path/filepath"
)

type updatableFile[T any] struct {
	Updatable[T]

	w   *fsnotify.Watcher
	f   string
	tFn TransformFunc[[]byte, T]
}

var _ Updatable[interface{}] = &updatableFile[interface{}]{}

func NewFile(f string, bufSize ...int) (Updatable[[]byte], error) {
	return NewFileWithTransform(f, IdentityTransform, bufSize...)
}

func NewFileWithTransform[T any](f string, tFn TransformFunc[[]byte, T], bufSize ...int) (Updatable[T], error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	f, err = filepath.Abs(f)
	if err != nil {
		return nil, err
	}

	dir := filepath.Dir(f)
	err = w.Add(dir)
	if err != nil {
		return nil, err
	}

	u := &updatableFile[T]{
		Updatable: NewChannel[T](bufSize...),

		w:   w,
		f:   f,
		tFn: tFn,
	}

	go u.start()
	return u, nil
}

func (u *updatableFile[T]) start() {
	ch := u.Updatable.((*updatableChannel[T])).Channel()

	update := func() error {
		c, err := os.ReadFile(u.f)
		if err != nil {
			return err
		}

		ver := fmt.Sprintf("%x", sha256.Sum256(c))
		val, _ := u.tFn(c) // TODO: report error?

		ch <- VersionedValue[T]{
			Version: ver,
			Value:   val,
		}
		return nil
	}

	_ = update() // TODO: report error?

	fe, ok := <-u.w.Events
	for ; ok; fe, ok = <-u.w.Events {
		if !(fe.Op.Has(fsnotify.Write) || fe.Op.Has(fsnotify.Create)) {
			continue
		}
		// fe.Name should be an absolut path as we watch only absolut paths
		if u.f != fe.Name {
			continue
		}
		_ = update() // TODO: report error?
	}
	// watcher closed => terminate
}

func (u *updatableFile[T]) Close() {
	u.w.Close()
}
