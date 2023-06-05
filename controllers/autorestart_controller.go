package controllers

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// AutoRestartReconciler reconciles a ConfigMap object
type AutoRestartReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;update
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;update

func (r *AutoRestartReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	configMap := &corev1.ConfigMap{}
	err := r.Get(ctx, req.NamespacedName, configMap)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Only restart the associated Pod if it belongs to the "fedx-1000" namespace
	if req.Namespace == "fedx-1000" {
		// Get the corresponding Pod
		pod := &corev1.Pod{}
		err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: configMap.Name}, pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Update the Pod's restart timestamp annotation to trigger a restart
		if pod.Annotations == nil {
			pod.Annotations = make(map[string]string)
		}
		pod.Annotations["auto-restart-timestamp"] = "restart" // Update the timestamp to trigger a restart

		err = r.Update(ctx, pod)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *AutoRestartReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		Complete(r)
}
