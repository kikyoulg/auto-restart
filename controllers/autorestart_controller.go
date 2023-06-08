// autorestart_controller.go

package controllers

import (
	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type AutoRestartReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *AutoRestartReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	//新建configMap对象，类型为指向 corev1.ConfigMap 结构体的指针,可以通过 corev1.ConfigMap{} 的字段来修改 ConfigMap 对象的键值对数据
	configMap := &corev1.ConfigMap{}

	/*
		r.Get 方法的作用是从 Kubernetes API Server 中获取指定的资源对象，并将其存储到指定的变量中
	*/
	err := r.Get(ctx, req.NamespacedName, configMap)

	//检查cm是否存在

	//if err != nil {
	//	if errors.IsNotFound(err) {
	//		// ConfigMap has been deleted, no action needed
	//		return reconcile.Result{}, nil
	//	}
	//	return reconcile.Result{}, err
	//}

	if req.Namespace != "fedx-1000" {
		return reconcile.Result{}, nil
	}

	labelSelector := labels.SelectorFromSet(configMap.Labels)

	//声明corev1.PodList 类型的指针变量 podList
	podList := &corev1.PodList{}

	//定义了一个 client.ListOption 类型的切片，该切片的名称为 listOpts
	listOpts := []client.ListOption{
		/*
			client.InNamespace(req.Namespace)：只获取指定命名空间下的 Pod 资源对象
			client.MatchingLabelsSelector{Selector: labelSelector}：只获取标签匹配指定选择器的 Pod 资源对象
		*/
		client.InNamespace(req.Namespace),
		client.MatchingLabelsSelector{Selector: labelSelector},
	}
	//使用了 r.List 方法从 Kubernetes API Server 中获取指定选项的 Pod 资源对象列表，并将其存储到 podList 变量中
	err = r.List(ctx, podList, listOpts...)
	if err != nil {
		return reconcile.Result{}, err
	}

	for _, pod := range podList.Items {
		podName := pod.Name
		podNamespace := pod.Namespace

		fmt.Printf("Restarting Pod %s in namespace %s\n", podName, podNamespace)
		//r.Delete 方法是 Kubernetes 客户端库中 client.Client 接口的一个方法，用于删除指定的 Kubernetes 资源对象
		err = r.Delete(ctx, &pod, client.PropagationPolicy(metav1.DeletePropagationBackground))
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
