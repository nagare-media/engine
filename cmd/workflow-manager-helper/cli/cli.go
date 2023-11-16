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
	"fmt"
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
	workflowmanagerhelper "github.com/nagare-media/engine/internal/workflow-manager-helper"
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

	fs := flag.NewFlagSet("workflow-manager-helper", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Fprint(fs.Output(), "Usage: workflow-manager-helper [options] <data>\n")
		fs.PrintDefaults()
	}

	var cfgFile string
	fs.StringVar(&cfgFile, "config", "", "Location of the workflow-manager-helper configuration file")

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

	// configure workflow-manager-helper

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
		WithName("workflow-manager-helper")
	ctx = log.IntoContext(ctx, l)
	log.SetLogger(l)

	// parse data
	if fs.NArg() != 1 {
		err = fmt.Errorf("invalid number or positional arguments: %d", fs.NArg())
		setupLog.Error(err, "setup failed")
		fs.Usage()
		return err
	}

	dataFile := fs.Arg(0)
	dataStr, err := os.ReadFile(dataFile)
	if err != nil {
		setupLog.Error(err, "unable to read data file")
		return err
	}

	// TODO: decoding seems to be lax; disallow unknown fields
	data := &enginev1.WorkflowManagerHelperData{}
	codecs := serializer.NewCodecFactory(scheme)
	err = runtime.DecodeInto(codecs.UniversalDecoder(), dataStr, data)
	if err != nil {
		setupLog.Error(err, "unable to parse data file")
		return err
	}

	// parse config
	if cfgFile == "" {
		// TODO: make this optional
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
	cfg := &enginev1.WorkflowManagerHelperConfiguration{}
	err = runtime.DecodeInto(codecs.UniversalDecoder(), cfgStr, cfg)
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

	// create components

	reportsCtrl := workflowmanagerhelper.NewReportsController(cfg, data)
	taskCtrl := workflowmanagerhelper.NewTaskController(cfg, data)

	// start components

	var (
		reportsCtrlErr, taskCtrlErr error
		reportsCtrlDone             = make(chan struct{})
		taskCtrlDone                = make(chan struct{})
	)

	go func() { reportsCtrlErr = reportsCtrl.Start(ctx); close(reportsCtrlDone) }()
	go func() { taskCtrlErr = taskCtrl.Start(ctx); close(taskCtrlDone) }()

	// termination handling

	select {
	case <-reportsCtrlDone:
		cancel()
	case <-taskCtrlDone:
		cancel()
	case <-ctx.Done():
	}
	<-taskCtrlDone
	<-reportsCtrlDone

	if reportsCtrlErr != nil {
		err = reportsCtrlErr
		setupLog.Error(err, "problem running reports controller")
	}

	if taskCtrlErr != nil {
		err = taskCtrlErr
		setupLog.Error(err, "problem running task controller")
	}

	return err
}
