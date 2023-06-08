package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type AutoRestartReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	ProcessedConfigs map[string]string
}

// NewAutoRestartReconciler creates an instance of AutoRestartReconciler with initialization
func NewAutoRestartReconciler(client client.Client, scheme *runtime.Scheme) *AutoRestartReconciler {
	return &AutoRestartReconciler{
		Client:           client,
		Scheme:           scheme,
		ProcessedConfigs: make(map[string]string),
	}
}

// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch
func (r *AutoRestartReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var configMap corev1.ConfigMap
	if err := r.Get(ctx, req.NamespacedName, &configMap); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	fmt.Printf("ConfigMap %s in namespace %s changed, ResourceVersion: %s\n", req.Name, req.Namespace, configMap.ObjectMeta.ResourceVersion)

	configKey := fmt.Sprintf("%s/%s", configMap.Namespace, configMap.Name)
	if val, ok := r.ProcessedConfigs[configKey]; ok && val == configMap.ObjectMeta.ResourceVersion {
		return ctrl.Result{}, nil
	}
	r.ProcessedConfigs[configKey] = configMap.ObjectMeta.ResourceVersion

	// Find Pods which have the same ConfigMap volume mount
	podList := &corev1.PodList{}
	if err := r.List(ctx, podList, client.InNamespace(req.Namespace)); err != nil {
		return ctrl.Result{}, err
	}

	for _, pod := range podList.Items {
		for _, volume := range pod.Spec.Volumes {
			if volume.ConfigMap != nil && volume.ConfigMap.Name == configMap.Name {
				fmt.Printf("Restarting Pod %s in namespace %s\n", pod.Name, pod.Namespace)
				if err := r.Delete(ctx, &pod); err != nil {
					return ctrl.Result{}, err
				}
			}
		}
	}

	return ctrl.Result{}, nil
}

func (r *AutoRestartReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.ConfigMap{}).
		Complete(NewAutoRestartReconciler(mgr.GetClient(), mgr.GetScheme()))
}
