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

	"github.com/mattn/go-isatty"
	"go.uber.org/zap/zapcore"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	ctrlcfg "sigs.k8s.io/controller-runtime/pkg/config"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/internal/pkg/version"
	"github.com/nagare-media/engine/internal/workflow-manager/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme = runtime.NewScheme()
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(enginev1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
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

	fs := flag.NewFlagSet("workflow-manager", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() {
		fmt.Fprint(fs.Output(), "Usage: workflow-manager [options]\n")
		fs.PrintDefaults()
	}

	var cfgFile string
	fs.StringVar(&cfgFile, "config", "", "Location of the workflow-manager configuration file")

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
		WithName("workflow-manager")
	ctx = log.IntoContext(ctx, l)
	ctrl.SetLogger(l)
	klog.SetLogger(l) // see https://github.com/kubernetes-sigs/controller-runtime/issues/1420

	// parse config
	cfg := &enginev1.WorkflowManagerConfig{}

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

	mgrOpts := c.getManagerOptions(cfg)

	// create components

	restCfg := ctrl.GetConfigOrDie()
	mgr, err := ctrl.NewManager(restCfg, mgrOpts)
	if err != nil {
		setupLog.Error(err, "unable to start workflow-manager")
		return err
	}

	if err = (&enginev1.Function{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Function")
		return err
	}
	if err = (&enginev1.ClusterFunction{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "ClusterFunction")
		return err
	}
	if err = (&enginev1.MediaLocation{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "MediaLocation")
		return err
	}
	if err = (&enginev1.ClusterMediaLocation{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "ClusterMediaLocation")
		return err
	}
	if err = (&enginev1.MediaProcessingEntity{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "MediaProcessingEntity")
		return err
	}
	if err = (&enginev1.ClusterMediaProcessingEntity{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "ClusterMediaProcessingEntity")
		return err
	}
	if err = (&enginev1.TaskTemplate{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "TaskTemplate")
		return err
	}
	if err = (&enginev1.ClusterTaskTemplate{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "ClusterTaskTemplate")
		return err
	}
	if err = (&enginev1.Task{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Task")
		return err
	}
	if err = (&enginev1.Workflow{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "Workflow")
		return err
	}
	jobEventChannel := make(chan event.GenericEvent)
	mpeReconciler := &controllers.MediaProcessingEntityReconciler{
		Config:          cfg,
		APIReader:       mgr.GetAPIReader(),
		Client:          mgr.GetClient(),
		Scheme:          mgr.GetScheme(),
		LocalRESTConfig: restCfg,
		ManagerOptions:  mgrOpts,
		JobEventChannel: jobEventChannel,
	}
	if err = mpeReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "MediaProcessingEntity")
		return err
	}
	if err = (&controllers.WorkflowReconciler{
		Config:    cfg,
		APIReader: mgr.GetAPIReader(),
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Workflow")
		return err
	}
	if err = (&controllers.TaskReconciler{
		Config:                          cfg,
		APIReader:                       mgr.GetAPIReader(),
		Client:                          mgr.GetClient(),
		Scheme:                          mgr.GetScheme(),
		JobEventChannel:                 jobEventChannel,
		MediaProcessingEntityReconciler: mpeReconciler,
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Task")
		return err
	}
	// +kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		return err
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		return err
	}

	// start components

	setupLog.Info("starting workflow-manager")
	if err := mgr.Start(ctx); err != nil {
		setupLog.Error(err, "problem running workflow-manager")
		return err
	}

	return nil
}

func (c *cli) getManagerOptions(cfg *enginev1.WorkflowManagerConfig) ctrl.Options {
	mgrOpts := ctrl.Options{
		Scheme: scheme,
	}

	// Cache

	mgrOpts.Cache.DefaultNamespaces = map[string]cache.Config{}
	if cfg.Cache.Namespace != nil {
		mgrOpts.Cache.DefaultNamespaces[*cfg.Cache.Namespace] = cache.Config{}
	}
	if cfg.Cache.SyncPeriod != nil {
		mgrOpts.Cache.SyncPeriod = &cfg.Cache.SyncPeriod.Duration
	}

	// LeaderElection

	if cfg.LeaderElection.LeaderElect != nil {
		mgrOpts.LeaderElection = *cfg.LeaderElection.LeaderElect
	}
	mgrOpts.LeaderElectionResourceLock = cfg.LeaderElection.ResourceLock
	mgrOpts.LeaderElectionNamespace = cfg.LeaderElection.ResourceNamespace
	mgrOpts.LeaderElectionID = cfg.LeaderElection.ResourceName
	if cfg.LeaderElection.LeaseDuration.Duration != 0 {
		mgrOpts.LeaseDuration = &cfg.LeaderElection.LeaseDuration.Duration
	}
	if cfg.LeaderElection.RenewDeadline.Duration != 0 {
		mgrOpts.RenewDeadline = &cfg.LeaderElection.RenewDeadline.Duration
	}
	if cfg.LeaderElection.RetryPeriod.Duration != 0 {
		mgrOpts.RetryPeriod = &cfg.LeaderElection.RetryPeriod.Duration
	}

	// Metrics

	mgrOpts.Metrics.BindAddress = cfg.Metrics.BindAddress

	// Health

	mgrOpts.HealthProbeBindAddress = cfg.Health.BindAddress
	mgrOpts.LivenessEndpointName = cfg.Health.LivenessEndpointName
	mgrOpts.ReadinessEndpointName = cfg.Health.ReadinessEndpointName

	// Webhook

	whOpts := webhook.Options{}
	whOpts.Host = cfg.Webhook.Host
	if cfg.Webhook.Port != nil {
		whOpts.Port = *cfg.Webhook.Port
	}
	whOpts.CertDir = cfg.Webhook.CertDir
	mgrOpts.WebhookServer = webhook.NewServer(whOpts)

	// GracefulShutdownTimeout

	if cfg.GracefulShutdownTimeout != nil {
		mgrOpts.GracefulShutdownTimeout = &cfg.GracefulShutdownTimeout.Duration
	}

	// Controller

	mgrOpts.Controller = ctrlcfg.Controller(cfg.Controller)

	return mgrOpts
}
