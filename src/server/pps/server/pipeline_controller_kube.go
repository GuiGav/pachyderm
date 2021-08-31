package server

import (
	"context"
	"math/rand"
	"os"
	"time"

	crd "github.com/pachyderm/pachyderm/v2/src/server/pps/server/crd"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var (
	setupLog = ctrl.Log.WithName("setup")
)

type reconciler struct {
	client.Client
	scheme *runtime.Scheme
}

func (r *reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("chaospod", req.NamespacedName)
	log.V(1).Info("reconciling chaos pod")

	var chaospod crd.ChaosPod
	if err := r.Get(ctx, req.NamespacedName, &chaospod); err != nil {
		log.Error(err, "unable to get chaosctl")
		return ctrl.Result{}, err
	}

	var pod corev1.Pod
	podFound := true
	if err := r.Get(ctx, req.NamespacedName, &pod); err != nil {
		if !apierrors.IsNotFound(err) {
			log.Error(err, "unable to get pod")
			return ctrl.Result{}, err
		}
		podFound = false
	}

	if podFound {
		shouldStop := chaospod.Spec.NextStop.Time.Before(time.Now())
		if !shouldStop {
			return ctrl.Result{RequeueAfter: chaospod.Spec.NextStop.Sub(time.Now()) + 1*time.Second}, nil
		}

		if err := r.Delete(ctx, &pod); err != nil {
			log.Error(err, "unable to delete pod")
			return ctrl.Result{}, err
		}

		return ctrl.Result{Requeue: true}, nil
	}

	templ := chaospod.Spec.Template.DeepCopy()
	pod.ObjectMeta = templ.ObjectMeta
	pod.Name = req.Name
	pod.Namespace = req.Namespace
	pod.Spec = templ.Spec

	if err := ctrl.SetControllerReference(&chaospod, &pod, r.scheme); err != nil {
		log.Error(err, "unable to set pod's owner reference")
		return ctrl.Result{}, err
	}

	if err := r.Create(ctx, &pod); err != nil {
		log.Error(err, "unable to create pod")
		return ctrl.Result{}, err
	}

	chaospod.Spec.NextStop.Time = time.Now().Add(time.Duration(10*(rand.Int63n(2)+1)) * time.Second)
	chaospod.Status.LastRun = pod.CreationTimestamp
	if err := r.Update(ctx, &chaospod); err != nil {
		log.Error(err, "unable to update chaosctl status")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func StartController() {
	//ctrl.SetLogger(zap.New())

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// in a real controller, we'd create a new scheme for this
	err = crd.AddToScheme(mgr.GetScheme())
	if err != nil {
		setupLog.Error(err, "unable to add scheme")
		os.Exit(1)
	}

	err = ctrl.NewControllerManagedBy(mgr).
		For(&crd.ChaosPod{}).
		Owns(&corev1.Pod{}).
		Complete(&reconciler{
			Client: mgr.GetClient(),
			scheme: mgr.GetScheme(),
		})
	if err != nil {
		setupLog.Error(err, "unable to create controller")
		os.Exit(1)
	}

	err = ctrl.NewWebhookManagedBy(mgr).
		For(&crd.ChaosPod{}).
		Complete()
	if err != nil {
		setupLog.Error(err, "unable to create webhook")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
