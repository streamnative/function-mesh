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

	"github.com/streamnative/function-mesh/utils"
	"k8s.io/client-go/tools/record"

	"github.com/go-logr/logr"
	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MeshConfigReconciler reconciles a MeshConfig object
type MeshConfigReconciler struct {
	client.Client
	Log      logr.Logger
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

// +kubebuilder:rbac:groups=compute.functionmesh.io,resources=MeshConfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=compute.functionmesh.io,resources=MeshConfigs/status,verbs=get;update;patch

func (r *MeshConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("MeshConfig", req.NamespacedName)

	// your logic here
	if req.Name == utils.GlobalMeshConfig && req.Namespace == utils.GlobalMeshConfigNamespace {
		// for global mesh config
		functions := &v1alpha1.FunctionList{}
		err := r.List(ctx, functions)
		if err != nil {
			r.Log.Error(err, "failed to list all functions")
		}
		for _, function := range functions.Items {
			r.Recorder.Event(&function, "Normal", "GlobalMeshConfigUpdate", "Mesh config is updated")
		}

		sinks := &v1alpha1.SinkList{}
		err = r.List(ctx, sinks)
		if err != nil {
			r.Log.Error(err, "failed to list all sinks")
		}
		for _, sink := range sinks.Items {
			r.Recorder.Event(&sink, "Normal", "GlobalMeshConfigUpdate", "Mesh config is updated")
		}

		sources := &v1alpha1.SourceList{}
		err = r.List(ctx, sources)
		if err != nil {
			r.Log.Error(err, "failed to list all sources")
		}
		for _, source := range sources.Items {
			r.Recorder.Event(&source, "Normal", "GlobalMeshConfigUpdate", "Mesh config is updated")
		}
	} else if req.Name == utils.NamespacedMeshConfig {
		// for namespaced mesh config
		functions := &v1alpha1.FunctionList{}
		err := r.List(ctx, functions, client.InNamespace(req.Namespace))
		if err != nil {
			r.Log.Error(err, "failed to list functions")
		}
		for _, function := range functions.Items {
			r.Recorder.Event(&function, "Normal", "NamespacedMeshConfigUpdate", "Mesh config is updated")
		}

		sinks := &v1alpha1.SinkList{}
		err = r.List(ctx, sinks, client.InNamespace(req.Namespace))
		if err != nil {
			r.Log.Error(err, "failed to list sinks")
		}
		for _, sink := range sinks.Items {
			r.Recorder.Event(&sink, "Normal", "NamespacedMeshConfigUpdate", "Mesh config is updated")
		}

		sources := &v1alpha1.SourceList{}
		err = r.List(ctx, sources, client.InNamespace(req.Namespace))
		if err != nil {
			r.Log.Error(err, "failed to list sources")
		}
		for _, source := range sources.Items {
			r.Recorder.Event(&source, "Normal", "NamespacedMeshConfigUpdate", "Mesh config is updated")
		}
	}

	return ctrl.Result{}, nil
}

func (r *MeshConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	r.Recorder = mgr.GetEventRecorderFor("MeshConfig")

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.MeshConfig{}).
		Complete(r)
}
