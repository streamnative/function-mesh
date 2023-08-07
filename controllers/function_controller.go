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

	v1 "k8s.io/api/batch/v1"
	"k8s.io/client-go/rest"

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

// FunctionReconciler reconciles a Function object
type FunctionReconciler struct {
	client.Client
	Config            *rest.Config
	RestClient        rest.Interface
	Log               logr.Logger
	Scheme            *runtime.Scheme
	GroupVersionFlags *utils.GroupVersionFlags
}

// +kubebuilder:rbac:groups=compute.functionmesh.io,resources=functions,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=compute.functionmesh.io,resources=functions/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=compute.functionmesh.io,resources=functions/finalizers,verbs=get;update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;create;update;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=secrets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling.k8s.io,resources=verticalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;create;update;delete

func (r *FunctionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("function", req.NamespacedName)

	// your logic here
	function := &v1alpha1.Function{}
	err := r.Get(ctx, req.NamespacedName, function)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		r.Log.Error(err, "failed to get function")
		return reconcile.Result{}, err
	}

	if !spec.IsManaged(function) {
		r.Log.Info("Skipping function not managed by the controller", "Name", req.String())
		return reconcile.Result{}, nil
	}

	// initialize component status map
	if function.Status.Conditions == nil {
		function.Status.Conditions = make(map[v1alpha1.Component]v1alpha1.ResourceCondition)
	}

	err = r.ObserveFunctionStatefulSet(ctx, function)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ObserveFunctionService(ctx, function)
	if err != nil {
		return reconcile.Result{}, err
	}
	if r.GroupVersionFlags != nil && r.GroupVersionFlags.APIAutoscalingGroupVersion == utils.GroupVersionV2Beta2 {
		err = r.ObserveFunctionHPAV2Beta2(ctx, function)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if r.GroupVersionFlags != nil && r.GroupVersionFlags.APIAutoscalingGroupVersion == utils.GroupVersionV2 {
		err = r.ObserveFunctionHPA(ctx, function)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	if r.GroupVersionFlags != nil && r.GroupVersionFlags.WatchVPACRDs {
		err = r.ObserveFunctionVPA(ctx, function)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	err = r.Status().Update(ctx, function)
	if err != nil {
		r.Log.Error(err, "failed to update function status after observing resources")
		return ctrl.Result{}, err
	}

	isNewGeneration := r.checkIfFunctionGenerationsIsIncreased(function)

	err = r.ApplyFunctionStatefulSet(ctx, function, isNewGeneration)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ApplyFunctionService(ctx, function, isNewGeneration)
	if err != nil {
		return reconcile.Result{}, err
	}
	if r.GroupVersionFlags != nil && r.GroupVersionFlags.APIAutoscalingGroupVersion == utils.GroupVersionV2Beta2 {
		err = r.ApplyFunctionHPAV2Beta2(ctx, function, isNewGeneration)
		if err != nil {
			return reconcile.Result{}, err
		}
	} else if r.GroupVersionFlags != nil && r.GroupVersionFlags.APIAutoscalingGroupVersion == utils.GroupVersionV2 {
		err = r.ApplyFunctionHPA(ctx, function, isNewGeneration)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	err = r.ApplyFunctionVPA(ctx, function)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ApplyFunctionCleanUpJob(ctx, function)
	if err != nil {
		return reconcile.Result{}, err
	}

	// don't need to update status since function is deleting
	if !function.ObjectMeta.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}
	function.Status.ObservedGeneration = function.Generation
	err = r.Status().Update(ctx, function)
	if err != nil {
		r.Log.Error(err, "failed to update function status after applying resources")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *FunctionReconciler) checkIfFunctionGenerationsIsIncreased(function *v1alpha1.Function) bool {
	return function.Generation != function.Status.ObservedGeneration
}

func (r *FunctionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	manager := ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Function{}).
		Owns(&appsv1.StatefulSet{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&corev1.Service{}).
		Owns(&corev1.Secret{}).
		Owns(&v1.Job{})

	if r.GroupVersionFlags != nil && r.GroupVersionFlags.WatchVPACRDs {
		manager.Owns(&vpav1.VerticalPodAutoscaler{})
	}

	if r.GroupVersionFlags != nil && r.GroupVersionFlags.APIAutoscalingGroupVersion != "" {
		AddControllerBuilderOwn(manager, r.GroupVersionFlags.APIAutoscalingGroupVersion)
	}

	return manager.Complete(r)
}
