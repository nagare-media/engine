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

package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
	"go.uber.org/zap/zapcore"

	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nagare-media/engine/internal/functions"
	"github.com/nagare-media/engine/internal/pkg/version"
)

const BaseBinaryName = "functions"

var (
	requiredNArgs = 1
	fnArg         = -1
	tddArg        = 0
)

type cli struct{}

func New() *cli {
	return &cli{}
}

func (c *cli) Execute(ctx context.Context, fn string, args []string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	setupLog := log.FromContext(ctx).WithName("setup")

	// setup CLI flags

	baseBinary := strings.HasPrefix(fn, BaseBinaryName)
	if baseBinary {
		requiredNArgs = 2
		fnArg++
		tddArg++
	}

	fs := flag.NewFlagSet("functions", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		if baseBinary {
			fmt.Fprintf(fs.Output(), "Usage: %s [options] <function> <task description document>\n", fn)
		} else {
			fmt.Fprintf(fs.Output(), "Usage: %s [options] <task description document>\n", fn)
		}
		fs.PrintDefaults()
	}

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

	// configure

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
		WithName("functions")
	ctx = log.IntoContext(ctx, l)
	log.SetLogger(l)
	klog.SetLogger(l) // see https://github.com/kubernetes-sigs/controller-runtime/issues/1420

	if fs.NArg() != requiredNArgs {
		err = fmt.Errorf("invalid number or positional arguments: %d", fs.NArg())
		setupLog.Error(err, "setup failed")
		fs.Usage()
		return err
	}

	if baseBinary {
		fn = fs.Arg(fnArg)
	}
	tddPath := fs.Arg(tddArg)

	// create and start task controller

	tskCtrl := &functions.TaskController{
		FunctionName: fn,
		TDDPath:      tddPath,
	}

	if err = tskCtrl.Start(ctx); err != nil {
		setupLog.Error(err, "problem running task controller")
		return err
	}

	return nil
}
