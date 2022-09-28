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
	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// FunctionMeshReconciler reconciles a FunctionMesh object
type FunctionMeshReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=compute.functionmesh.io,resources=functionmeshes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=compute.functionmesh.io,resources=functionmeshes/status,verbs=get;update;patch

func (r *FunctionMeshReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	_ = r.Log.WithValues("functionMesh", req.NamespacedName)

	// your logic here
	mesh := &v1alpha1.FunctionMesh{}
	err := r.Get(ctx, req.NamespacedName, mesh)
	if err != nil {
		if errors.IsNotFound(err) {
			// mesh must be deleted
			return ctrl.Result{}, nil
		}
		r.Log.Error(err, "failed to get mesh")
		return reconcile.Result{Requeue: true}, err
	}

	if !spec.IsManaged(mesh) {
		r.Log.Info("Skipping function mesh not managed by the controller", "Name", req.String())
		return reconcile.Result{}, nil
	}

	// initialize component status map
	if mesh.Status.FunctionConditions == nil {
		mesh.Status.FunctionConditions = make(map[string]v1alpha1.ResourceCondition)
	}
	if mesh.Status.SourceConditions == nil {
		mesh.Status.SourceConditions = make(map[string]v1alpha1.ResourceCondition)
	}
	if mesh.Status.SinkConditions == nil {
		mesh.Status.SinkConditions = make(map[string]v1alpha1.ResourceCondition)
	}

	// TODO validate function spec correctness such as no duplicated func name etc

	// make observations
	err = r.ObserveFunctionMesh(ctx, req, mesh)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.Status().Update(ctx, mesh)
	if err != nil {
		r.Log.Error(err, "failed to update mesh status")
		return ctrl.Result{}, err
	}

	isNewGeneration := r.checkIfFunctionMeshGenerationsIsIncreased(mesh)

	// apply changes
	err = r.UpdateFunctionMesh(ctx, req, mesh, isNewGeneration)
	if err != nil {
		return reconcile.Result{}, err
	}

	mesh.Status.ObservedGeneration = mesh.Generation
	err = r.Status().Update(ctx, mesh)
	if err != nil {
		r.Log.Error(err, "failed to update functionmesh status")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *FunctionMeshReconciler) checkIfFunctionMeshGenerationsIsIncreased(mesh *v1alpha1.FunctionMesh) bool {
	return mesh.Generation != mesh.Status.ObservedGeneration
}

func (r *FunctionMeshReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.FunctionMesh{}).
		Owns(&v1alpha1.Function{}).
		Owns(&v1alpha1.Source{}).
		Owns(&v1alpha1.Sink{}).
		Complete(r)
}
