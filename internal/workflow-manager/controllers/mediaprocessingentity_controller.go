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

package controllers

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	kerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	enginev1 "github.com/nagare-media/engine/api/v1alpha1"
	"github.com/nagare-media/engine/internal/pkg/apis/utils"
	apiclient "github.com/nagare-media/engine/internal/workflow-manager/client"
	"github.com/nagare-media/engine/pkg/apis/meta"
)

const (
	MediaProcessingEntityControllerName        = "nagare-media-engine-mediaprocessingentity-controller"
	ClusterMediaProcessingEntityControllerName = "nagare-media-engine-clustermediaprocessingentity-controller"
)

// MediaProcessingEntityReconciler reconciles MediaProcessingEntities and ClusterMediaProcessingEntities objects
type MediaProcessingEntityReconciler struct {
	client.Client
	APIReader         client.Reader
	readOnlyAPIClient client.Client

	Config          *enginev1.WorkflowManagerConfig
	Scheme          *runtime.Scheme
	LocalRESTConfig *rest.Config
	ManagerOptions  manager.Options
	JobEventChannel chan<- event.GenericEvent

	mtx                sync.RWMutex
	managers           map[meta.ObjectReference]Manager
	managerCancelFuncs map[meta.ObjectReference]context.CancelFunc
	managerErrs        map[meta.ObjectReference]error

	mpeManagerErr  chan event.GenericEvent
	cmpeManagerErr chan event.GenericEvent
}

type Manager interface {
	manager.Manager

	IsRemote() bool
	IsLocal() bool
	Namespace() string
}

type mpeManager struct {
	manager.Manager
	remote    bool
	namespace string
}

func (m *mpeManager) IsRemote() bool {
	return m.remote
}

func (m *mpeManager) IsLocal() bool {
	return !m.remote
}

func (m *mpeManager) Namespace() string {
	return m.namespace
}

type Client interface {
	client.Client

	IsRemote() bool
	IsLocal() bool
	Namespace() string
}

type mpeClient struct {
	client.Client
	remote    bool
	namespace string
}

func (c *mpeClient) IsRemote() bool {
	return c.remote
}

func (c *mpeClient) IsLocal() bool {
	return !c.remote
}

func (c *mpeClient) Namespace() string {
	return c.namespace
}

// +kubebuilder:rbac:groups="",resources=secrets,verbs=get
// +kubebuilder:rbac:groups=engine.nagare.media,resources=clustermediaprocessingentities,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=engine.nagare.media,resources=clustermediaprocessingentities/finalizers,verbs=update
// +kubebuilder:rbac:groups=engine.nagare.media,resources=clustermediaprocessingentities/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=engine.nagare.media,resources=mediaprocessingentities,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=engine.nagare.media,resources=mediaprocessingentities/finalizers,verbs=update
// +kubebuilder:rbac:groups=engine.nagare.media,resources=mediaprocessingentities/status,verbs=get;update;patch

func (r *MediaProcessingEntityReconciler) reconcileMediaProcessingEntity(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := logf.FromContext(ctx)

	// fetch MediaProcessingEntity
	mpe := &enginev1.MediaProcessingEntity{}
	if err := r.Client.Get(ctx, req.NamespacedName, mpe); err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return ctrl.Result{}, nil
		}
		log.Error(err, "error fetching MediaProcessingEntity")
		return ctrl.Result{}, err
	}

	// apply patches on return
	oldMPE := mpe.DeepCopy()
	defer func() {
		r.reconcileCondition(ctx, reterr, mpe)
		_, err := utils.FullPatch(ctx, r.Client, mpe, oldMPE)
		apiErrs := kerrors.FilterOut(err, apierrors.IsNotFound)
		if apiErrs != nil {
			log.Error(apiErrs, "error patching MediaProcessingEntity")
		}
		reterr = kerrors.Flatten(kerrors.NewAggregate([]error{apiErrs, reterr}))
	}()

	// add finalizers
	if controllerutil.AddFinalizer(mpe, enginev1.MediaProcessingEntityProtectionFinalizer) {
		return ctrl.Result{}, nil
	}

	// handle delete
	if !mpe.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, mpe)
	}

	// handle normal reconciliation
	return r.reconcile(ctx, mpe)
}

func (r *MediaProcessingEntityReconciler) reconcileClusterMediaProcessingEntity(ctx context.Context, req ctrl.Request) (_ ctrl.Result, reterr error) {
	log := logf.FromContext(ctx)

	// fetch ClusterMediaProcessingEntity
	cmpe := &enginev1.ClusterMediaProcessingEntity{}
	if err := r.Client.Get(ctx, req.NamespacedName, cmpe); err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return ctrl.Result{}, nil
		}
		log.Error(err, "error fetching ClusterMediaProcessingEntity")
		return ctrl.Result{}, err
	}

	// apply patches on return
	oldMPE := cmpe.DeepCopy()
	defer func() {
		r.reconcileCondition(ctx, reterr, (*enginev1.MediaProcessingEntity)(cmpe))
		_, err := utils.FullPatch(ctx, r.Client, cmpe, oldMPE)
		apiErrs := kerrors.FilterOut(err, apierrors.IsNotFound)
		if apiErrs != nil {
			log.Error(apiErrs, "error patching ClusterMediaProcessingEntity")
		}
		reterr = kerrors.Flatten(kerrors.NewAggregate([]error{apiErrs, reterr}))
	}()

	// add finalizers
	if controllerutil.AddFinalizer(cmpe, enginev1.MediaProcessingEntityProtectionFinalizer) {
		return ctrl.Result{}, nil
	}

	// we can convert CMPE to MPE since both have the same structure

	// handle delete
	if !cmpe.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, (*enginev1.MediaProcessingEntity)(cmpe))
	}

	// handle normal reconciliation
	return r.reconcile(ctx, (*enginev1.MediaProcessingEntity)(cmpe))
}

func (r *MediaProcessingEntityReconciler) reconcileCondition(ctx context.Context, err error, mpe *enginev1.MediaProcessingEntity) {
	log := logf.FromContext(ctx)

	if err != nil {
		log.Error(err, fmt.Sprintf("error reconciling %s", mpe.Kind))
		mpe.Status.Message = err.Error()
		mpe.Status.Conditions = utils.MarkConditionTrue(mpe.Status.Conditions, enginev1.MediaProcessingEntityFailedConditionType)
		mpe.Status.Conditions = utils.MarkConditionFalse(mpe.Status.Conditions, enginev1.MediaProcessingEntityReadyConditionType)
		return
	}

	readyCondition := corev1.ConditionFalse
	_, ready := r.GetManager(utils.ToRef(mpe))
	if ready {
		readyCondition = corev1.ConditionTrue
	}

	mpe.Status.Message = ""
	mpe.Status.Conditions = utils.DeleteCondition(mpe.Status.Conditions, enginev1.MediaProcessingEntityFailedConditionType)
	mpe.Status.Conditions = utils.MarkCondition(mpe.Status.Conditions, enginev1.MediaProcessingEntityReadyConditionType, readyCondition)
}

func (r *MediaProcessingEntityReconciler) reconcileDelete(ctx context.Context, mpe *enginev1.MediaProcessingEntity) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info(fmt.Sprintf("reconcile deleted %s", mpe.Kind))

	ref := utils.ToRef(mpe)

	// check if there are jobs in running state
	// TODO: implement (if yes return with error / just requeue)

	// stop manager for MPE
	r.deleteAndStopManager(ref)

	// remove finalizer
	controllerutil.RemoveFinalizer(mpe, enginev1.MediaProcessingEntityProtectionFinalizer)

	return ctrl.Result{}, nil
}

func (r *MediaProcessingEntityReconciler) reconcile(ctx context.Context, mpe *enginev1.MediaProcessingEntity) (ctrl.Result, error) {
	log := logf.FromContext(ctx)
	log.Info(fmt.Sprintf("reconcile %s", mpe.Kind))

	ref := utils.ToRef(mpe)

	// did we already create a manager and it failed?
	if err := r.getManagerErr(ref); err != nil {
		// reset and try again
		r.deleteAndStopManager(ref)
		return ctrl.Result{}, err
	}

	// did we already create a manager?
	if _, ok := r.GetManager(ref); ok {
		return ctrl.Result{}, nil
	}

	// create manager
	var mgr Manager
	var stabilizeDuration time.Duration
	var err error

	if mpe.Spec.Local != nil {
		stabilizeDuration = 0
		mgr, err = r.newLocalManager(ctx, mpe)
	} else if mpe.Spec.Remote != nil {
		stabilizeDuration = r.Config.RemoteMediaProcessingEntityStabilizingDuration.Duration
		mgr, err = r.newRemoteManager(ctx, mpe)
	}
	if err != nil {
		return ctrl.Result{}, err
	}

	// add job controller
	if err = (&JobReconciler{
		Config:                   r.Config,
		APIReader:                mgr.GetAPIReader(),
		Client:                   mgr.GetClient(),
		Scheme:                   mgr.GetScheme(),
		EventChannel:             r.JobEventChannel,
		MediaProcessingEntityRef: ref,
	}).SetupWithManager(mgr); err != nil {
		log.Error(err, "unable to create controller", "controller", "Job")
		return ctrl.Result{}, err
	}

	// start manager
	mgrCtx, cancel := context.WithCancel(ctx)
	stopped := make(chan struct{})
	go func() {
		defer cancel()
		err := mgr.Start(mgrCtx)
		r.mtx.Lock()
		r.managerErrs[*ref] = err
		r.mtx.Unlock()
		close(stopped)

		if err != nil {
			// manager failed -> we should reconcile again and report the status
			e := event.GenericEvent{
				Object: &metav1.PartialObjectMetadata{
					ObjectMeta: metav1.ObjectMeta{
						Name:      ref.Name,
						Namespace: ref.Namespace,
					},
				},
			}
			switch ref.Kind {
			case "MediaProcessingEntity":
				r.mpeManagerErr <- e
			case "ClusterMediaProcessingEntity":
				r.cmpeManagerErr <- e
			default:
				panic("reference to unknown media processing entity kind")
			}
		}
	}()

	// wait for manager to stabilize
	select {
	case <-stopped:
		err := r.getManagerErr(ref)
		if err != nil {
			return ctrl.Result{}, err
		}
	case <-time.After(stabilizeDuration):
		// no error: assuming manager started successfully
		// TODO: there could still be a race with a manager failing just after stabilizeDuration
		r.addManager(ref, mgr, cancel)
	}

	// MediaProcessingEntity is ready
	return ctrl.Result{}, nil
}

func (r *MediaProcessingEntityReconciler) newLocalManager(_ context.Context, mpe *enginev1.MediaProcessingEntity) (Manager, error) {
	// shallow copy
	opts := r.ManagerOptions

	// set namespace
	var namespace string
	if mpe.Spec.Local.Namespace != "" {
		namespace = mpe.Spec.Local.Namespace
	} else if mpe.Namespace != "" {
		namespace = mpe.Namespace
	} else {
		// ClusterMediaProcessingEntity with no namespace
		return nil, errors.New("local ClusterMediaProcessingEntity has no namespace defined")
	}
	opts.Cache = cache.Options{
		DefaultNamespaces: map[string]cache.Config{
			namespace: {},
		},
	}

	mgr, err := manager.New(r.LocalRESTConfig, opts)
	if err != nil {
		return nil, err
	}

	return &mpeManager{
		Manager:   mgr,
		remote:    false,
		namespace: namespace,
	}, nil
}

func (r *MediaProcessingEntityReconciler) newRemoteManager(ctx context.Context, mpe *enginev1.MediaProcessingEntity) (Manager, error) {
	// shallow copy
	secretRef := mpe.Spec.Remote.Kubeconfig.SecretRef
	if err := utils.NormalizeRef(r.Scheme, &secretRef.ObjectReference, &corev1.Secret{}); err != nil {
		return nil, err
	}

	// set namespace
	if secretRef.Namespace == "" {
		if mpe.Namespace == "" {
			// ClusterMediaProcessingEntity with no namespace
			return nil, errors.New("remote ClusterMediaProcessingEntity secretRef has no namespace defined")
		}
		secretRef.Namespace = mpe.Namespace
	}

	// check namespace for MediaProcessingEntity
	if mpe.Namespace != "" && mpe.Namespace != secretRef.Namespace {
		return nil, errors.New("remote MediaProcessingEntity references secretRef in different namespace")
	}

	// resolve secret
	if err := utils.ResolveSecretRefInline(ctx, r.readOnlyAPIClient, &secretRef); err != nil {
		return nil, err
	}

	// get kubeconfig
	secretKey := enginev1.DefaultSecretKeyKubeconfig
	if secretRef.Key != "" {
		secretKey = secretRef.Key
	}
	kubeconfig := secretRef.Data[secretKey]
	if kubeconfig == nil {
		return nil, fmt.Errorf("Secret '%s/%s' has no key '%s'", secretRef.Namespace, secretRef.Name, secretKey)
	}

	// create REST config and read out namespace
	clientConfig, err := clientcmd.NewClientConfigFromBytes(kubeconfig)
	if err != nil {
		return nil, err
	}

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, err
	}

	// these options are set by controller-runtime as well (see config.GetConfigWithContext)
	if restConfig.QPS == 0.0 {
		restConfig.QPS = 20.0
		restConfig.Burst = 30.0
	}

	// set namespace
	namespace, _, err := clientConfig.Namespace()
	if err != nil {
		return nil, err
	}
	// shallow copy
	opts := r.ManagerOptions
	opts.Cache = cache.Options{
		DefaultNamespaces: map[string]cache.Config{
			namespace: {},
		},
	}

	mgr, err := manager.New(restConfig, opts)
	if err != nil {
		return nil, err
	}

	return &mpeManager{
		Manager:   mgr,
		remote:    true,
		namespace: namespace,
	}, nil
}

func (r *MediaProcessingEntityReconciler) GetClient(ref *meta.ObjectReference) (Client, bool) {
	mgr, ok := r.GetManager(ref)
	if !ok {
		return nil, false
	}
	return &mpeClient{
		Client:    mgr.GetClient(),
		remote:    mgr.IsRemote(),
		namespace: mgr.Namespace(),
	}, true
}

func (r *MediaProcessingEntityReconciler) GetManager(ref *meta.ObjectReference) (Manager, bool) {
	r.mtx.RLock()
	defer r.mtx.RUnlock()
	mgr, ok := r.managers[*ref]
	return mgr, ok
}

func (r *MediaProcessingEntityReconciler) addManager(ref *meta.ObjectReference, mgr Manager, cancel context.CancelFunc) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	r.managers[*ref] = mgr
	r.managerCancelFuncs[*ref] = cancel
	delete(r.managerErrs, *ref)
}

func (r *MediaProcessingEntityReconciler) deleteAndStopManager(ref *meta.ObjectReference) {
	r.mtx.Lock()
	defer r.mtx.Unlock()

	cancel, ok := r.managerCancelFuncs[*ref]
	if ok {
		cancel()
	}

	delete(r.managerCancelFuncs, *ref)
	delete(r.managers, *ref)
	delete(r.managerErrs, *ref)
}

func (r *MediaProcessingEntityReconciler) getManagerErr(ref *meta.ObjectReference) error {
	r.mtx.Lock()
	defer r.mtx.Unlock()
	return r.managerErrs[*ref]
}

// TODO: emit Kubernetes events
// SetupWithManager sets up the controller with the Manager.
func (r *MediaProcessingEntityReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.managers = make(map[meta.ObjectReference]Manager)
	r.managerCancelFuncs = make(map[meta.ObjectReference]context.CancelFunc)
	r.managerErrs = make(map[meta.ObjectReference]error)
	r.mpeManagerErr = make(chan event.GenericEvent)
	r.cmpeManagerErr = make(chan event.GenericEvent)

	// submanager should not use leader election
	r.ManagerOptions.LeaderElection = false

	// submanager should not bind for metrics metrics
	r.ManagerOptions.Metrics = metricsserver.Options{BindAddress: "0"}

	// submanager should not bind for metrics metrics
	r.ManagerOptions.HealthProbeBindAddress = "0"

	// submanager should not create a webhook server
	r.ManagerOptions.WebhookServer = nil

	// create read-only client
	r.readOnlyAPIClient = apiclient.NewReadOnlyClient(r.APIReader, r.Scheme, r.Client.RESTMapper())

	// MediaProcessingEntity controller
	if err := ctrl.NewControllerManagedBy(mgr).
		Named(MediaProcessingEntityControllerName).
		For(&enginev1.MediaProcessingEntity{}).
		WatchesRawSource(source.Channel(r.mpeManagerErr, &handler.EnqueueRequestForObject{})).
		Complete(reconcile.Func(r.reconcileMediaProcessingEntity)); err != nil {
		return err
	}

	// ClusterMediaProcessingEntity controller
	if err := ctrl.NewControllerManagedBy(mgr).
		Named(ClusterMediaProcessingEntityControllerName).
		For(&enginev1.ClusterMediaProcessingEntity{}).
		WatchesRawSource(source.Channel(r.cmpeManagerErr, &handler.EnqueueRequestForObject{})).
		Complete(reconcile.Func(r.reconcileClusterMediaProcessingEntity)); err != nil {
		return err
	}

	return nil
}
