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
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/mattn/go-isatty"
	"go.uber.org/zap/zapcore"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/internal/manager/controllers"
	// +kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(enginev1.AddToScheme(scheme))
	// +kubebuilder:scaffold:scheme
}

func Execute() error {
	var configFile string
	flag.StringVar(&configFile, "config", "",
		"The controller will load its initial configuration from this file. "+
			"Omit this flag to use the default configuration values. "+
			"Command-line flags override configuration from this file.")

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
	logOpts.BindFlags(flag.CommandLine)

	flag.Parse()

	l := zap.New(zap.UseFlagOptions(&logOpts)).
		WithName("nagare-media").
		WithName("engine").
		WithName("manager")
	ctrl.SetLogger(l)
	klog.SetLogger(l) // see https://github.com/kubernetes-sigs/controller-runtime/issues/1420

	var err error
	ctrlConfig := enginev1.ControllerManagerConfiguration{}
	options := ctrl.Options{Scheme: scheme}
	if configFile != "" {
		// TODO: migrate to custom configuration logic
		options, err = options.AndFrom(ctrl.ConfigFile().AtPath(configFile).OfKind(&ctrlConfig))
		if err != nil {
			setupLog.Error(err, "unable to load the config file")
			return err
		}
	}

	ctrlConfig.Default()
	if err = ctrlConfig.Validate(); err != nil {
		setupLog.Error(err, "invalid configuration")
		return err
	}

	restCfg := ctrl.GetConfigOrDie()
	mgr, err := ctrl.NewManager(restCfg, options)
	if err != nil {
		setupLog.Error(err, "unable to start manager")
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
		os.Exit(1)
	}
	if err = (&enginev1.ClusterMediaProcessingEntity{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "ClusterMediaProcessingEntity")
		os.Exit(1)
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
		Config:          ctrlConfig.NagareMediaEngineControllerManagerConfiguration,
		APIReader:       mgr.GetAPIReader(),
		Client:          mgr.GetClient(),
		Scheme:          mgr.GetScheme(),
		LocalRESTConfig: restCfg,
		ManagerOptions:  options,
		JobEventChannel: jobEventChannel,
	}
	if err = mpeReconciler.SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "MediaProcessingEntity")
		return err
	}
	if err = (&controllers.WorkflowReconciler{
		Config:    ctrlConfig.NagareMediaEngineControllerManagerConfiguration,
		APIReader: mgr.GetAPIReader(),
		Client:    mgr.GetClient(),
		Scheme:    mgr.GetScheme(),
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Workflow")
		return err
	}
	if err = (&controllers.TaskReconciler{
		Config:                          ctrlConfig.NagareMediaEngineControllerManagerConfiguration,
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

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		return err
	}

	return nil
}
