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

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	autov2beta2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	computeapi "github.com/streamnative/function-mesh/api/compute/v1alpha2"
	apispec "github.com/streamnative/function-mesh/controllers/spec"
	"github.com/streamnative/function-mesh/utils"
)

// SourceReconciler reconciles a Source object
type SourceReconciler struct {
	client.Client
	Log        logr.Logger
	Scheme     *runtime.Scheme
	WatchFlags *utils.WatchFlags
}

// +kubebuilder:rbac:groups=compute.functionmesh.io,resources=sources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=compute.functionmesh.io,resources=sources/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling.k8s.io,resources=verticalpodautoscalers,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;create;update;delete

func (r *SourceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("source", req.NamespacedName)

	// your logic here
	source := &computeapi.Source{}
	err := r.Get(ctx, req.NamespacedName, source)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		r.Log.Error(err, "failed to get source")
		return reconcile.Result{}, err
	}

	if !apispec.IsManaged(source) {
		r.Log.Info("Skipping source not managed by the controller", "Name", req.String())
		return reconcile.Result{}, nil
	}

	defer source.SaveStatus(ctx, r.Log, r.Client)
	if result, err := r.observe(ctx, source); err != nil {
		return result, err
	}
	if result, err := r.reconcile(ctx, source); err != nil {
		return result, err
	}

	return ctrl.Result{}, nil
}

func (r *SourceReconciler) observe(ctx context.Context, source *computeapi.Source) (ctrl.Result, error) {
	if err := r.ObserveSourceStatefulSet(ctx, source); err != nil {
		return reconcile.Result{}, err
	}
	if err := r.ObserveSourceService(ctx, source); err != nil {
		return reconcile.Result{}, err
	}
	if err := r.ObserveSourceHPA(ctx, source); err != nil {
		return reconcile.Result{}, err
	}
	if r.WatchFlags != nil && r.WatchFlags.WatchVPACRDs {
		if err := r.ObserveSourceVPA(ctx, source); err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *SourceReconciler) reconcile(ctx context.Context, source *computeapi.Source) (ctrl.Result, error) {
	if err := r.ApplySourceStatefulSet(ctx, source); err != nil {
		return reconcile.Result{}, err
	}
	if err := r.ApplySourceService(ctx, source); err != nil {
		return reconcile.Result{}, err
	}
	if err := r.ApplySourceHPA(ctx, source); err != nil {
		return reconcile.Result{}, err
	}
	if r.WatchFlags != nil && r.WatchFlags.WatchVPACRDs {
		if err := r.ApplySourceVPA(ctx, source); err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *SourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	manager := ctrl.NewControllerManagedBy(mgr).
		For(&computeapi.Source{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&autov2beta2.HorizontalPodAutoscaler{})
	if r.WatchFlags != nil && r.WatchFlags.WatchVPACRDs {
		manager.Owns(&vpav1.VerticalPodAutoscaler{})
	}
	return manager.Complete(r)
}
