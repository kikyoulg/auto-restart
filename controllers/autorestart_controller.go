package controllers

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/source"

	appsv1 "fanjl/auto-restart/api/v1"
	coreV1 "k8s.io/api/core/v1"
)

// AutoRestartReconciler reconciles a AutoRestart object
type AutoRestartReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps.auto-restart,resources=autorestarts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.auto-restart,resources=autorestarts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.auto-restart,resources=autorestarts/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *AutoRestartReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Get the Autorestart resource
	var autorestart appsv1.AutoRestart
	err := r.Get(ctx, req.NamespacedName, &autorestart)
	if err != nil {
		log.Error(err, "failed to get AutoRestart")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Get all Pods in the fedx-1000 namespace
	var pods coreV1.PodList
	err = r.List(ctx, &pods, client.InNamespace("fedx-1000"))
	if err != nil {
		log.Error(err, "failed to list Pods")
		return ctrl.Result{}, err
	}

	// Check if any Pod's ConfigMap has been updated
	for _, pod := range pods.Items {
		for _, vol := range pod.Spec.Volumes {
			if vol.ConfigMap != nil {
				cm := &coreV1.ConfigMap{}
				err := r.Get(ctx, types.NamespacedName{Name: vol.ConfigMap.Name, Namespace: pod.Namespace}, cm)
				if err != nil {
					log.Error(err, "failed to get ConfigMap")
					continue
				}
				if controllerutil.ContainsFinalizer(&vol.ConfigMap.ObjectMeta, "auto-restart-finalizer") && !cm.ObjectMeta.GetDeletionTimestamp().IsZero() {
					// The ConfigMap is being deleted, do nothing
					continue
				}
				if cm.ObjectMeta.GetResourceVersion() != vol.ConfigMap.ResourceVersion {
					// The ConfigMap has been updated, restart the Pod
					log.Info("ConfigMap updated, restarting Pod", "Pod", pod.Name, "ConfigMap", cm.Name)
					err = r.Delete(ctx, &pod)
					if err != nil {
						log.Error(err, "failed to delete Pod")
						continue
					}
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AutoRestartReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.AutoRestart{}).
		Watches(&source.Kind{Type: &coreV1.ConfigMap{}}, &handler.EnqueueRequestForObject{}).
		Complete(r)
}
