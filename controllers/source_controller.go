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

	"github.com/streamnative/function-mesh/controllers/spec"

	"github.com/go-logr/logr"
	computev1alpha1 "github.com/streamnative/function-mesh/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	autov2beta2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// SourceReconciler reconciles a Source object
type SourceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=compute.functionmesh.io,resources=sources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=compute.functionmesh.io,resources=sources/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete

func (r *SourceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("source", req.NamespacedName)

	// your logic here
	source := &computev1alpha1.Source{}
	err := r.Get(ctx, req.NamespacedName, source)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		r.Log.Error(err, "failed to get source")
		return reconcile.Result{}, err
	}
	var configHash string
	if source.Spec.SourceConfig != nil {
		configHash, err = spec.ComputeConfigHash(source.Spec.SourceConfig.Data)
	}
	if err != nil {
		r.Log.Error(err, "fail to compute source config hash")
		return reconcile.Result{}, err
	}

	if source.Status.Conditions == nil {
		source.Status.Conditions = make(map[computev1alpha1.Component]computev1alpha1.ResourceCondition)
	}

	err = r.ObserveSourceStatefulSet(ctx, req, source, configHash)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ObserveSourceService(ctx, req, source)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ObserveSourceHPA(ctx, req, source)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.Status().Update(ctx, source)
	if err != nil {
		r.Log.Error(err, "failed to update source status")
		return ctrl.Result{}, err
	}

	err = r.ApplySourceStatefulSet(ctx, req, source, configHash)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ApplySourceService(ctx, req, source)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = r.ApplySourceHPA(ctx, req, source)
	if err != nil {
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *SourceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&computev1alpha1.Source{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Owns(&autov2beta2.HorizontalPodAutoscaler{}).
		Complete(r)
}
