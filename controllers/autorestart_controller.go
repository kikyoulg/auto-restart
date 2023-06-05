// autorestart_controller.go

package controllers

import (
	"context"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// AutoRestartReconciler reconciles a AutoRestart object
type AutoRestartReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// ...

// Reconcile is the main reconciliation loop of the AutoRestart controller
func (r *AutoRestartReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("autorestart", req.NamespacedName)

	// Fetch the ConfigMap
	configMap := &corev1.ConfigMap{}
	err := r.Client.Get(ctx, req.NamespacedName, configMap)
	if err != nil {
		if errors.IsNotFound(err) {
			// ConfigMap is deleted, do cleanup here if necessary
			return reconcile.Result{}, nil
		}
		// Error occurred, requeue the request
		return reconcile.Result{}, err
	}

	// Check if the ConfigMap belongs to the specified namespace
	if configMap.Namespace != "fedx-1000" {
		// ConfigMap is not in the target namespace, ignore it
		return reconcile.Result{}, nil
	}

	// ConfigMap has been updated, restart the corresponding Pod(s)
	podList := &corev1.PodList{}
	err = r.Client.List(ctx, podList, &client.ListOptions{Namespace: "fedx-1000"})
	if err != nil {
		return reconcile.Result{}, err
	}

	for _, pod := range podList.Items {
		// Check if the Pod references the ConfigMap
		for _, volume := range pod.Spec.Volumes {
			if volume.ConfigMap != nil && volume.ConfigMap.Name == configMap.Name {
				// Restart the Pod by deleting it
				err = r.Client.Delete(ctx, &pod)
				if err != nil {
					return reconcile.Result{}, err
				}
				log.Info("Pod restarted", "pod", pod.Name)
				break
			}
		}
	}

	return reconcile.Result{}, nil
}

// ...

// SetupWithManager sets up the controller with the Manager
func (r *AutoRestartReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
