/*


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

	autov1 "k8s.io/api/autoscaling/v1"

	"github.com/go-logr/logr"
	cloudv1alpha1 "github.com/streamnative/mesh-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// SinkReconciler reconciles a Sink object
type SinkReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=cloud.streamnative.io,resources=sinks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=cloud.streamnative.io,resources=sinks/status,verbs=get;update;patch

func (r *SinkReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("sink", req.NamespacedName)

	// your logic here
	sink := &cloudv1alpha1.Sink{}
	err := r.Get(ctx, req.NamespacedName, sink)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		r.Log.Error(err, "failed to get sink")
		return reconcile.Result{}, err
	}

	if sink.Status.Conditions == nil {
		sink.Status.Conditions = make(map[cloudv1alpha1.Component]cloudv1alpha1.ResourceCondition)
	}

	err = r.ObserveSinkStatefulSet(ctx, req, sink)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ObserveSinkService(ctx, req, sink)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ObserveSinkHPA(ctx, req, sink)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.Status().Update(ctx, sink)
	if err != nil {
		r.Log.Error(err, "failed to update sink status")
		return ctrl.Result{}, err
	}

	err = r.ApplySinkStatefulSet(ctx, req, sink)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ApplySinkService(ctx, req, sink)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ApplySinkHPA(ctx, req, sink)
	if err != nil {
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *SinkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&cloudv1alpha1.Sink{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&autov1.HorizontalPodAutoscaler{}).
		Complete(r)
}
