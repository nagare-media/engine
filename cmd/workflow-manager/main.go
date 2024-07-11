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

package main

import (
	"os"

	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/nagare-media/engine/cmd/workflow-manager/cli"
)

func main() {
	ctx := signals.SetupSignalHandler()
	log.IntoContext(ctx, log.Log.
		WithName("nagare-media").
		WithName("engine").
		WithName("workflow-manager"))

	c := cli.New()
	if err := c.Execute(ctx, os.Args[1:]); err != nil {
		os.Exit(1)
	}
}
