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

package cli

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"go.uber.org/zap/zapcore"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/klog/v2"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/internal/pkg/version"
	workflowmanagerhelper "github.com/nagare-media/engine/internal/workflow-manager-helper"
	"github.com/nagare-media/engine/pkg/updatable"
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
		WithName("workflow-manager-helper")
	ctx = log.IntoContext(ctx, l)
	log.SetLogger(l)
	klog.SetLogger(l) // see https://github.com/kubernetes-sigs/controller-runtime/issues/1420

	// parse data
	if fs.NArg() != 1 {
		err = fmt.Errorf("invalid number or positional arguments: %d", fs.NArg())
		setupLog.Error(err, "setup failed")
		fs.Usage()
		return err
	}

	codecs := serializer.NewCodecFactory(scheme)

	dataFile := fs.Arg(0)
	data, err := updatable.NewFileWithTransform(dataFile, func(s []byte) (*enginev1.WorkflowManagerHelperData, error) {
		data := &enginev1.WorkflowManagerHelperData{}
		err = runtime.DecodeInto(codecs.UniversalDecoder(), s, data)
		if err != nil {
			setupLog.Error(err, "unable to parse data file")
			return nil, err
		}
		return data, nil
	})
	if err != nil {
		setupLog.Error(err, "unable to read data file")
		return err
	}
	defer data.Close()

	// parse config
	cfg := &enginev1.WorkflowManagerHelperConfig{}

	if cfgFile != "" {
		cfgStr, err := os.ReadFile(cfgFile)
		if err != nil {
			setupLog.Error(err, "unable to read config file")
			return err
		}

		err = runtime.DecodeInto(codecs.UniversalDecoder(), cfgStr, cfg)
		if err != nil {
			setupLog.Error(err, "unable to parse config file")
			return err
		}
	}

	cfg.Default()
	if err = cfg.Validate(); err != nil {
		setupLog.Error(err, "invalid configuration")
		return err
	}

	// create components

	taskCtrl := workflowmanagerhelper.NewTaskController(cfg, data)
	reportsCtrl := workflowmanagerhelper.NewReportsController(cfg, data)

	// start components

	var (
		taskCtrlErr, reportsCtrlErr error
		taskCtrlDone                = make(chan struct{})
		reportsCtrlDone             = make(chan struct{})

		// create a new context for the reports controller as it should only be terminated after the task controller has terminated
		reportsCtx, reportsCtxCancel = context.WithCancel(context.WithoutCancel(ctx))
	)

	go func() { taskCtrlErr = taskCtrl.Start(ctx); close(taskCtrlDone); reportsCtxCancel() }()
	go func() { reportsCtrlErr = reportsCtrl.Start(reportsCtx); close(reportsCtrlDone) }()

	// termination handling

	select {
	case <-taskCtrlDone:
		// we already canceled reportsCtx in the Go routine
	case <-reportsCtrlDone:
		cancel()
	case <-ctx.Done():
	}
	<-taskCtrlDone
	<-reportsCtrlDone

	if taskCtrlErr != nil {
		err = taskCtrlErr
		setupLog.Error(err, "problem running task controller")
	}

	if reportsCtrlErr != nil {
		err = reportsCtrlErr
		setupLog.Error(err, "problem running reports controller")
	}

	return err
}
