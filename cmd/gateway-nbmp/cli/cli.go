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

package cli

import (
	"context"
	"errors"
	"flag"
	"os"

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/mattn/go-isatty"
	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/internal/gateway-nbmp/http"
	"github.com/nagare-media/engine/internal/pkg/version"
)

type cli struct {
	Config enginev1.GatewayNBMPConfiguration
	scheme *runtime.Scheme
}

func New() *cli {
	return &cli{}
}

func (c *cli) InjectScheme(scheme *runtime.Scheme) error {
	c.scheme = scheme
	return nil
}

func (c *cli) Execute(ctx context.Context, args []string) error {
	setupLog := log.FromContext(ctx).WithName("setup")

	// setup CLI flags

	fs := flag.NewFlagSet("gateway-nbmp", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)

	var configFile string
	fs.StringVar(&configFile, "config", "", "Location of the gateway-nbmp configuration file")

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

	l := zap.New(zap.UseFlagOptions(&logOpts)).WithName("engine")
	ctx = log.IntoContext(ctx, l)
	log.SetLogger(l)

	if configFile == "" {
		err := errors.New("-config option missing")
		setupLog.Error(err, "setup failed")
		fs.Usage()
		return err
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	codecs := serializer.NewCodecFactory(c.scheme)
	err = runtime.DecodeInto(codecs.UniversalDecoder(), content, &c.Config)
	if err != nil {
		setupLog.Error(err, "unable to load the config file")
		return err
	}

	// start components

	ws := http.NewServer(c.Config.Webserver)
	if err := ws.Start(ctx); err != nil {
		setupLog.Error(err, "problem running webserver")
		return err
	}

	return nil
}
