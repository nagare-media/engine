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

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/klog/v2"

	"github.com/mattn/go-isatty"
	"go.uber.org/zap/zapcore"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/internal/gateway-nbmp/http"
	"github.com/nagare-media/engine/internal/gateway-nbmp/svc"
	"github.com/nagare-media/engine/internal/pkg/version"
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(enginev1.AddToScheme(scheme))
}

type cli struct{}

func New() *cli {
	return &cli{}
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

	// configure gateway-nbmp

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
	klog.SetLogger(l) // see https://github.com/kubernetes-sigs/controller-runtime/issues/1420

	if configFile == "" {
		err := errors.New("--config option missing")
		setupLog.Error(err, "setup failed")
		fs.Usage()
		return err
	}

	content, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	var cfg enginev1.GatewayNBMPConfiguration
	codecs := serializer.NewCodecFactory(scheme)
	err = runtime.DecodeInto(codecs.UniversalDecoder(), content, &cfg)
	if err != nil {
		setupLog.Error(err, "unable to load the config file")
		return err
	}

	cfg.Default()

	// configure Kubernetes API client

	k8sCfg := ctrl.GetConfigOrDie()

	mapper, err := apiutil.NewDynamicRESTMapper(k8sCfg)
	if err != nil {
		return err
	}

	k8sCache, err := cache.New(k8sCfg, cache.Options{Scheme: scheme, Mapper: mapper})
	if err != nil {
		return err
	}

	k8sClient, err := client.New(k8sCfg, client.Options{Scheme: scheme, Mapper: mapper})
	if err != nil {
		return err
	}
	k8sClient = client.NewNamespacedClient(k8sClient, cfg.Services.KubernetesNamespace)
	k8sClient, err = client.NewDelegatingClient(client.NewDelegatingClientInput{
		CacheReader: k8sCache,
		Client:      k8sClient,
	})
	if err != nil {
		return err
	}

	// start components

	wfsvc := svc.NewWorkflowService(&cfg, k8sClient)

	ws := http.NewServer(&cfg, wfsvc)
	if err := ws.Start(ctx); err != nil {
		setupLog.Error(err, "problem running webserver")
		return err
	}

	return nil
}
