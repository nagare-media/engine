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
		// TODO: filter by event operation (based on Kubernetes behavior):
		//     CREATE "/run/secrets/engine.nagare.media/task/..2024_12_13_10_48_04.1309658368"
		//     CHMOD  "/run/secrets/engine.nagare.media/task/..2024_12_13_10_48_04.1309658368"
		//     CREATE "/run/secrets/engine.nagare.media/task/..data_tmp"
		//     RENAME "/run/secrets/engine.nagare.media/task/..data_tmp"
		//     CREATE "/run/secrets/engine.nagare.media/task/..data ← /run/secrets/engine.nagare.media/task/..data_tmp"
		//     REMOVE "/run/secrets/engine.nagare.media/task/..2024_12_13_10_45_49.1604045035"
		_ = fe
		_ = update() // TODO: report error?
	}
	// watcher closed => terminate
}

func (u *updatableFile[T]) Close() {
	u.w.Close()
}
