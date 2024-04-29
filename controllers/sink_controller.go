// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package controllers

import (
	"context"
	"time"

	"github.com/streamnative/function-mesh/pkg/monitoring"
	v1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/handler"

	"github.com/go-logr/logr"
	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/spec"
	"github.com/streamnative/function-mesh/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// SinkReconciler reconciles a Topic object
type SinkReconciler struct {
	client.Client
	Config            *rest.Config
	RestClient        rest.Interface
	Log               logr.Logger
	Scheme            *runtime.Scheme
	GroupVersionFlags *utils.GroupVersionFlags
}

// +kubebuilder:rbac:groups=compute.functionmesh.io,resources=sinks,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=compute.functionmesh.io,resources=sinks/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=compute.functionmesh.io,resources=sinks/finalizers,verbs=get;update
// +kubebuilder:rbac:groups=compute.functionmesh.io,resources=backendconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=pods/exec,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling.k8s.io,resources=verticalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;create;update;delete

func (r *SinkReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("sink", req.NamespacedName)

	startTime := time.Now()

	defer func() {
		monitoring.FunctionMeshControllerReconcileCount.WithLabelValues("sink", req.NamespacedName.Name,
			req.NamespacedName.Namespace).Inc()
		monitoring.FunctionMeshControllerReconcileLatency.WithLabelValues("sink", req.NamespacedName.Name,
			req.NamespacedName.Namespace).Observe(float64(time.Since(startTime).Milliseconds()))
	}()

	// your logic here
	sink := &v1alpha1.Sink{}
	err := r.Get(ctx, req.NamespacedName, sink)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		r.Log.Error(err, "failed to get sink")
		return reconcile.Result{}, err
	}

	if !spec.IsManaged(sink) {
		r.Log.Info("Skipping sink not managed by the controller", "Name", req.String())
		return reconcile.Result{}, nil
	}

	if sink.Status.Conditions == nil {
		sink.Status.Conditions = make(map[v1alpha1.Component]v1alpha1.ResourceCondition)
	}

	err = r.ObserveSinkStatefulSet(ctx, sink)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ObserveSinkService(ctx, sink)
	if err != nil {
		return reconcile.Result{}, err
	}
	if r.GroupVersionFlags != nil && r.GroupVersionFlags.APIAutoscalingGroupVersion == utils.GroupVersionV2Beta2 {
		err = r.ObserveSinkHPAV2Beta2(ctx, sink)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if r.GroupVersionFlags != nil && r.GroupVersionFlags.APIAutoscalingGroupVersion == utils.GroupVersionV2 {
		err = r.ObserveSinkHPA(ctx, sink)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	if r.GroupVersionFlags != nil && r.GroupVersionFlags.WatchVPACRDs {
		err = r.ObserveSinkVPA(ctx, sink)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	err = r.Status().Update(ctx, sink)
	if err != nil {
		r.Log.Error(err, "failed to update sink status")
		return ctrl.Result{}, err
	}

	isNewGeneration := r.checkIfSinkGenerationsIsIncreased(sink)

	err = r.ApplySinkStatefulSet(ctx, sink, isNewGeneration)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ApplySinkService(ctx, sink, isNewGeneration)
	if err != nil {
		return reconcile.Result{}, err
	}
	if r.GroupVersionFlags != nil && r.GroupVersionFlags.APIAutoscalingGroupVersion == utils.GroupVersionV2Beta2 {
		err = r.ApplySinkHPAV2Beta2(ctx, sink, isNewGeneration)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if r.GroupVersionFlags != nil && r.GroupVersionFlags.APIAutoscalingGroupVersion == utils.GroupVersionV2 {
		err = r.ApplySinkHPA(ctx, sink, isNewGeneration)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	err = r.ApplySinkVPA(ctx, sink)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ApplySinkCleanUpJob(ctx, sink)
	if err != nil {
		return reconcile.Result{}, err
	}

	// don't need to update status since sink is deleting
	if !sink.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}
	sink.Status.ObservedGeneration = sink.Generation
	err = r.Status().Update(ctx, sink)
	if err != nil {
		r.Log.Error(err, "failed to update sink status")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *SinkReconciler) checkIfSinkGenerationsIsIncreased(sink *v1alpha1.Sink) bool {
	return sink.Generation != sink.Status.ObservedGeneration
}

func (r *SinkReconciler) SetupWithManager(mgr ctrl.Manager) error {
	manager := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Sink{}).
		Owns(&appsv1.StatefulSet{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&corev1.Service{}).
		Owns(&v1.Job{})
	if r.GroupVersionFlags != nil && r.GroupVersionFlags.WatchVPACRDs {
		manager.Owns(&vpav1.VerticalPodAutoscaler{})
	}

	if r.GroupVersionFlags != nil && r.GroupVersionFlags.APIAutoscalingGroupVersion != "" {
		AddControllerBuilderOwn(manager, r.GroupVersionFlags.APIAutoscalingGroupVersion)
	}

	manager.Watches(&v1alpha1.BackendConfig{}, handler.EnqueueRequestsFromMapFunc(
		func(ctx context.Context, object client.Object) []reconcile.Request {
			if object.GetName() == utils.GlobalBackendConfig && object.GetNamespace() == utils.GlobalBackendConfigNamespace {
				ctx := context.Background()
				sinks := &v1alpha1.SinkList{}
				err := mgr.GetClient().List(ctx, sinks)
				if err != nil {
					mgr.GetLogger().Error(err, "failed to list all sinks")
				}
				var requests []reconcile.Request
				for _, sink := range sinks.Items {
					requests = append(requests, reconcile.Request{
						NamespacedName: types.NamespacedName{Namespace: sink.Namespace, Name: sink.Name},
					})
				}
				return requests
			} else if object.GetName() == utils.NamespacedBackendConfig {
				ctx := context.Background()
				sinks := &v1alpha1.SinkList{}
				err := mgr.GetClient().List(ctx, sinks, client.InNamespace(object.GetNamespace()))
				if err != nil {
					mgr.GetLogger().Error(err, "failed to list sinks in namespace: "+object.GetNamespace())
				}
				var requests []reconcile.Request
				for _, sink := range sinks.Items {
					requests = append(requests, reconcile.Request{
						NamespacedName: types.NamespacedName{Namespace: sink.Namespace, Name: sink.Name},
					})
				}
				return requests
			} else {
				return nil
			}
		}),
	)

	return manager.Complete(r)
}
