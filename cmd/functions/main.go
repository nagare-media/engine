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

package main

import (
	"context"
	"os"
	"path/filepath"

	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/nagare-media/engine/cmd/functions/cli"

	// Import nagare media functions to be included in this multi-binary.
	_ "github.com/nagare-media/engine/internal/functions/functions/data-buffer-fs"
	_ "github.com/nagare-media/engine/internal/functions/functions/data-copy"
	_ "github.com/nagare-media/engine/internal/functions/functions/data-discard"
	_ "github.com/nagare-media/engine/internal/functions/functions/generic-noop"
	_ "github.com/nagare-media/engine/internal/functions/functions/generic-sleep"
	_ "github.com/nagare-media/engine/internal/functions/functions/media-encode"
	_ "github.com/nagare-media/engine/internal/functions/functions/media-generate-testpattern"
	_ "github.com/nagare-media/engine/internal/functions/functions/media-merge"
	_ "github.com/nagare-media/engine/internal/functions/functions/media-metadata-technical"
	_ "github.com/nagare-media/engine/internal/functions/functions/media-package-hls"
	_ "github.com/nagare-media/engine/internal/functions/functions/script-lua"

	// Import io implementations to be included in this multi-binary.
	_ "github.com/nagare-media/engine/internal/functions/io/protocols/buffered"
	_ "github.com/nagare-media/engine/internal/functions/io/protocols/http"
)

func main() {
	sigCtx := signals.SetupSignalHandler()
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-sigCtx.Done()
		cancel()
	}()

	log.IntoContext(ctx, log.Log.
		WithName("nagare-media").
		WithName("engine").
		WithName("functions"))

	fn := filepath.Base(os.Args[0])
	c := cli.New()
	if err := c.Execute(ctx, fn, os.Args[1:]); err != nil || sigCtx.Err() != nil {
		os.Exit(1)
	}
}
