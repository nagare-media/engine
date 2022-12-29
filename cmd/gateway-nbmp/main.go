/*
Copyright 2022 The nagare media authors

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

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/cmd/gateway-nbmp/cli"
)

var (
	scheme = runtime.NewScheme()
	logger = log.Log.WithName("engine")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(enginev1.AddToScheme(scheme))
}

func main() {
	ctx := signals.SetupSignalHandler()
	log.IntoContext(ctx, logger)

	c := cli.New()
	err := c.InjectScheme(scheme)
	if err != nil {
		os.Exit(1)
	}

	err = c.Execute(ctx, os.Args[1:])
	if err != nil {
		os.Exit(1)
	}
}
