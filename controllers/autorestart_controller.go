// autorestart_controller.go

package controllers

import (
	"context"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
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

	// Create a label selector for matching the associated Pod
	labelSelector := labels.Set{
		"app": "fedx-proxy",
	}.AsSelector()

	// Get the list of Pods matching the label selector and the ConfigMap's namespace
	podList := &corev1.PodList{}
	err = r.List(ctx, podList, client.InNamespace(req.Namespace), client.MatchingLabelsSelector{Selector: labelSelector})
	if err != nil {
		return reconcile.Result{}, err
	}

	// Restart the associated Pods
	for _, pod := range podList.Items {
		// Add any additional logic here before restarting the Pod

		// Set the owner reference of the Pod to the ConfigMap
		err = controllerutil.SetControllerReference(configMap, &pod, r.Scheme)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Delete and recreate the Pod to trigger a restart
		err = r.Delete(ctx, &pod)
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
