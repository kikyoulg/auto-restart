// autorestart_controller.go

package controllers

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
		if errors.IsNotFound(err) {
			// ConfigMap has been deleted, no action needed
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	// Check if the ConfigMap belongs to the desired namespace (fedx-1000)
	if req.Namespace != "fedx-1000" {
		// ConfigMap is not in the desired namespace, no action needed
		return reconcile.Result{}, nil
	}

	// Get the label selector from the ConfigMap's labels
	labelSelector := labels.SelectorFromSet(configMap.Labels)

	// Get the list of Pods matching the label selector in the desired namespace
	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(req.Namespace),
		client.MatchingLabelsSelector{Selector: labelSelector},
	}
	err = r.List(ctx, podList, listOpts...)
	if err != nil {
		return reconcile.Result{}, err
	}

	// Restart the associated Pods
	for _, pod := range podList.Items {
		podName := pod.Name
		podNamespace := pod.Namespace

		// Execute restart logic here for the Pod
		// ...

		fmt.Printf("Restarting Pod %s in namespace %s\n", podName, podNamespace)
		// Perform the Pod restart by deleting and recreating the Pod
		err = r.Delete(ctx, &pod, client.PropagationPolicy(metav1.DeletePropagationBackground))
		if err != nil {
			return reconcile.Result{}, err
		}

		// You can add any additional logic here after restarting the Pod
	}

	return reconcile.Result{}, nil
}

func (r *AutoRestartReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		Complete(r)
}
