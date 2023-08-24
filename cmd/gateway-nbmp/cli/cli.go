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

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"
	"k8s.io/client-go/rest"

	"github.com/mattn/go-isatty"
	"go.uber.org/zap/zapcore"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
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
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
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

	err = cfg.Validate()
	if err != nil {
		return err
	}

	// configure Kubernetes API client

	k8sCfg := ctrl.GetConfigOrDie()

	restClient, err := rest.HTTPClientFor(k8sCfg)
	if err != nil {
		return err
	}

	mapper, err := apiutil.NewDynamicRESTMapper(k8sCfg, restClient)
	if err != nil {
		return err
	}

	k8sCache, err := cache.New(k8sCfg, cache.Options{
		Scheme: scheme,
		Mapper: mapper,
		DefaultNamespaces: map[string]cache.Config{
			cfg.Services.KubernetesNamespace: {},
		},
	})
	if err != nil {
		return err
	}

	k8sClient, err := client.New(k8sCfg, client.Options{
		Scheme: scheme,
		Mapper: mapper,
		Cache: &client.CacheOptions{
			Reader: k8sCache,
		},
	})
	if err != nil {
		return err
	}
	k8sClient = client.NewNamespacedClient(k8sClient, cfg.Services.KubernetesNamespace)

	// create components

	wfsvc := svc.NewWorkflowService(&cfg, k8sClient)
	httpServer := http.NewServer(&cfg, wfsvc)

	// start components

	var k8sCacheErr error
	k8sCacheDone := make(chan struct{})
	go func() {
		k8sCacheErr = k8sCache.Start(ctx)
		close(k8sCacheDone)
	}()

	if ok := k8sCache.WaitForCacheSync(ctx); !ok {
		if k8sCacheErr == nil {
			k8sCacheErr = errors.New("problem starting Kubernetes API client cache")
		}
		setupLog.Error(k8sCacheErr, "problem starting Kubernetes API client cache")
		return k8sCacheErr
	}

	var httpServerErr error
	httpServerDone := make(chan struct{})
	go func() {
		httpServerErr = httpServer.Start(ctx)
		close(httpServerDone)
	}()

	// termination handling

	select {
	case <-k8sCacheDone:
		cancel()
	case <-httpServerDone:
		cancel()
	case <-ctx.Done():
	}
	<-k8sCacheDone
	<-httpServerDone

	if httpServerErr != nil {
		setupLog.Error(err, "problem running webserver")
	}
	return httpServerErr
}
