/*
Copyright 2024 changqings.

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

package controller

import (
	"context"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	somecnv1 "builder-tmp/api/v1"
)

// TmpReconciler reconciles a Tmp object
type TmpReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=some.cn,resources=tmps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=some.cn,resources=tmps/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=some.cn,resources=tmps/finalizers,verbs=update

// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;update;patch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=services/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=services/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Tmp object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *TmpReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	log.Info("trigger reconciler, exec ...")

	// TODO(user): your logic here
	tmp := &somecnv1.Tmp{}
	err := r.Get(ctx, req.NamespacedName, tmp)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return ctrl.Result{}, nil
		}
		log.Error(err, "unable to fetch Tmp")
		return ctrl.Result{RequeueAfter: time.Second * 10}, err
	}

	// createOrUpdate Service
	new_servcie := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      tmp.Spec.ServieName,
			Namespace: tmp.Namespace,
		},
	}
	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, new_servcie, func() error {

		mutaService(tmp, new_servcie)
		if err := controllerutil.SetControllerReference(tmp, new_servcie, r.Scheme); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		log.Error(err, "unable to create or update Service")
		return ctrl.Result{}, err
	}

	log.Info("Service reconcile success", "name", new_servcie.Name, "result", op)
	return ctrl.Result{}, nil
}

func mutaService(tmp *somecnv1.Tmp, service *corev1.Service) {
	service.Labels = map[string]string{
		"app": tmp.Spec.ServieName,
	}
	service.Spec = corev1.ServiceSpec{
		Selector: map[string]string{
			"app": tmp.Spec.ServieName,
		},
		Type: corev1.ServiceType(tmp.Spec.ServiceType),
		Ports: []corev1.ServicePort{
			{
				Port:       int32(tmp.Spec.Port),
				TargetPort: intstr.FromInt(tmp.Spec.ContainerPort),
				Name:       "http",
			},
		},
	}

}

// SetupWithManager sets up the controller with the Manager.
func (r *TmpReconciler) SetupWithManager(mgr ctrl.Manager) error {

	// 自定义索引，用于快速查询指定条件的pod列表
	indexer := mgr.GetFieldIndexer()
	var podIndexFieldKey = "pod.contains.app" // 使用方法为indexer.GetByField(podIndexFiledKey, "true")来获取所有包含app的pod列表
	indexer.IndexField(context.Background(), &corev1.Pod{}, podIndexFieldKey,
		func(o client.Object) []string {
			if p, ok := o.(*corev1.Pod); ok {
				for _, c := range p.Spec.Containers {
					if c.Name == "app" {
						return []string{"true"}
					}
				}
			}
			return []string{}
		})

	return ctrl.NewControllerManagedBy(mgr).
		For(&somecnv1.Tmp{}).
		Owns(&corev1.Service{}).
		Watches(&corev1.Pod{}, // Watches pod, and transfer pod.annotations to tmp.NamespacedName request, means you can trigger reconcile with resources not manager by this crd-controller
			handler.EnqueueRequestsFromMapFunc(handlerMapFunc),
			builder.WithPredicates( // predicates below should be all past, then trigger reconcile
				predicate.Funcs{ // only create or update pod can trigger reconcile
					CreateFunc: func(tce event.TypedCreateEvent[client.Object]) bool {
						return true
					},
					DeleteFunc: func(tde event.TypedDeleteEvent[client.Object]) bool { return false },
					UpdateFunc: func(tue event.TypedUpdateEvent[client.Object]) bool { //
						// pod resourcesVersion will update many times at creating pod, so we only care about pod.labels
						// log.Log.Info("debug", "old_rsv", tue.ObjectOld.GetResourceVersion(), "new_rsv", tue.ObjectNew.GetResourceVersion())
						return true
					},
				},
				predicate.NewPredicateFuncs(predicatePod),   // only pod with special annotation can trigger reconcile
				predicate.ResourceVersionChangedPredicate{}, // only pod version changed can trigger reconcile
			),
			builder.OnlyMetadata,
		).
		WithOptions(controller.Options{
			CacheSyncTimeout: time.Second * 10, // cache sync timeout, default 2s
		}).
		Complete(r)
}

func predicatePod(obj client.Object) bool { // when envent is UPDATE, enqueue old and new obj
	p, ok := obj.(*corev1.Pod)
	if !ok {
		return false
	}
	_, b := p.GetAnnotations()["somecnv1.tmp"]

	return b && p.Status.Phase == corev1.PodRunning
}

// for watches resources update, this func will trigger twice
func handlerMapFunc(ctx context.Context, obj client.Object) []reconcile.Request {

	v, ok := obj.GetAnnotations()["somecnv1.tmp"]
	if ok {
		log := log.FromContext(ctx, "name", obj.GetName(), "namespace", obj.GetNamespace())
		log.Info("map func for pod")
		return []reconcile.Request{
			{
				NamespacedName: client.ObjectKey{
					Namespace: obj.GetNamespace(),
					Name:      v,
				},
			},
		}
	}
	return []reconcile.Request{}
}
