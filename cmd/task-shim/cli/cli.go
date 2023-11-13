/*
Copyright 2022-2023 The nagare media authors

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

package cli

import (
	"context"
	"errors"
	"flag"
	"os"

	"github.com/mattn/go-isatty"
	"go.uber.org/zap/zapcore"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/internal/pkg/version"
	taskshim "github.com/nagare-media/engine/internal/task-shim"

	// Import task-shim actions to be included.
	_ "github.com/nagare-media/engine/internal/task-shim/actions/exec"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(enginev1.AddToScheme(scheme))
}

type cli struct{}

func New() *cli {
	return &cli{}
}

func (c *cli) Execute(ctx context.Context, args []string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	setupLog := log.FromContext(ctx).WithName("setup")

	// setup CLI flags

	fs := flag.NewFlagSet("task-shim", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	var cfgFile string
	fs.StringVar(&cfgFile, "config", "", "Location of the task-shim configuration file")

	var showUsage bool
	fs.BoolVar(&showUsage, "help", false, "Show help and exit")

	var showVersion bool
	fs.BoolVar(&showVersion, "version", false, "Show version and exit")

	logOpts := zap.Options{
		TimeEncoder: zapcore.ISO8601TimeEncoder,
	}
	logOpts.EncoderConfigOptions = []zap.EncoderConfigOption{
		func(ec *zapcore.EncoderConfig) {
			ec.EncodeLevel = zapcore.LowercaseLevelEncoder
			if logOpts.Development && (isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())) {
				ec.EncodeLevel = zapcore.LowercaseColorLevelEncoder
			}
		},
	}
	logOpts.BindFlags(fs)

	err := fs.Parse(args)
	if err != nil {
		setupLog.Error(err, "setup failed")
		return err
	}

	// configure task-shim

	if showUsage {
		fs.Usage()
		return nil
	}

	if showVersion {
		_ = version.Engine.Write(os.Stdout)
		return nil
	}

	l := zap.New(zap.UseFlagOptions(&logOpts)).
		WithName("nagare-media").
		WithName("engine").
		WithName("task-shim")
	ctx = log.IntoContext(ctx, l)
	log.SetLogger(l)

	if cfgFile == "" {
		err := errors.New("--config option missing")
		setupLog.Error(err, "setup failed")
		fs.Usage()
		return err
	}

	cfgStr, err := os.ReadFile(cfgFile)
	if err != nil {
		setupLog.Error(err, "unable to read config file")
		return err
	}

	// TODO: decoding seems to be lax; disallow unknown fields
	var cfg enginev1.TaskShimConfiguration
	codecs := serializer.NewCodecFactory(scheme)
	err = runtime.DecodeInto(codecs.UniversalDecoder(), cfgStr, &cfg)
	if err != nil {
		setupLog.Error(err, "unable to parse config file")
		return err
	}

	cfg.Default()

	err = cfg.Validate()
	if err != nil {
		setupLog.Error(err, "invalid configuration")
		return err
	}

	// create and start components

	// We work with two separate contexts:
	//   ctx     : was given to CLI and should normally only cancel if a termination signal was send by the OS
	//   httpCtx : is used for the HTTP server
	httpCtx, terminateCliFunc := context.WithCancel(context.Background())
	httpCtx = log.IntoContext(httpCtx, l)
	httpServer := taskshim.New(ctx, terminateCliFunc, &cfg)
	if err = httpServer.Start(httpCtx); err != nil {
		setupLog.Error(err, "problem running webserver")
		return err
	}

	return nil
}
