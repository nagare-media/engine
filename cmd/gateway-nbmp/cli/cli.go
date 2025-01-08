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
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/mattn/go-isatty"
	"go.uber.org/zap/zapcore"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	gatewaynbmp "github.com/nagare-media/engine/internal/gateway-nbmp"
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
	fs.Usage = func() {
		fmt.Fprint(fs.Output(), "Usage: gateway-nbmp [options]\n")
		fs.PrintDefaults()
	}

	var cfgFile string
	fs.StringVar(&cfgFile, "config", "", "Location of the gateway-nbmp configuration file")

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
		WithName("gateway-nbmp")
	ctx = log.IntoContext(ctx, l)
	log.SetLogger(l)
	klog.SetLogger(l) // see https://github.com/kubernetes-sigs/controller-runtime/issues/1420

	// parse config
	cfg := &enginev1.GatewayNBMPConfig{}

	if cfgFile != "" {
		cfgStr, err := os.ReadFile(cfgFile)
		if err != nil {
			setupLog.Error(err, "unable to read config file")
			return err
		}

		codecs := serializer.NewCodecFactory(scheme)
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

	// configure Kubernetes API client

	k8sCfg := ctrl.GetConfigOrDie()

	restClient, err := rest.HTTPClientFor(k8sCfg)
	if err != nil {
		setupLog.Error(err, "unable to create Kubernetes REST client")
		return err
	}

	mapper, err := apiutil.NewDynamicRESTMapper(k8sCfg, restClient)
	if err != nil {
		setupLog.Error(err, "unable to create Kubernetes REST mapper")
		return err
	}

	k8sCache, err := cache.New(k8sCfg, cache.Options{
		Scheme: scheme,
		Mapper: mapper,
		DefaultNamespaces: map[string]cache.Config{
			cfg.WorkflowService.Kubernetes.Namespace: {},
		},
	})
	if err != nil {
		setupLog.Error(err, "unable to create Kubernetes API cache")
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
		setupLog.Error(err, "unable to create Kubernetes API client")
		return err
	}
	k8sClient = client.NewNamespacedClient(k8sClient, cfg.WorkflowService.Kubernetes.Namespace)

	// create components

	gatewayNBMPCtrl := gatewaynbmp.New(cfg, k8sClient)

	// start components

	var k8sCacheErr error
	k8sCacheDone := make(chan struct{})
	go func() {
		k8sCacheErr = k8sCache.Start(ctx)
		close(k8sCacheDone)
	}()

	if ok := k8sCache.WaitForCacheSync(ctx); !ok {
		if k8sCacheErr == nil {
			k8sCacheErr = errors.New("unable to sync Kubernetes API cache")
		}
		setupLog.Error(k8sCacheErr, "unable to sync Kubernetes API cache")
		return k8sCacheErr
	}

	var gatewayNBMPCtrlErr error
	gatewayNBMPCtrlDone := make(chan struct{})
	go func() {
		gatewayNBMPCtrlErr = gatewayNBMPCtrl.Start(ctx)
		close(gatewayNBMPCtrlDone)
	}()

	// termination handling

	select {
	case <-k8sCacheDone:
		cancel()
	case <-gatewayNBMPCtrlDone:
		cancel()
	case <-ctx.Done():
	}
	<-k8sCacheDone
	<-gatewayNBMPCtrlDone

	if gatewayNBMPCtrlErr != nil {
		setupLog.Error(gatewayNBMPCtrlErr, "problem running webserver")
	}
	return gatewayNBMPCtrlErr
}
