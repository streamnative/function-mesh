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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/streamnative/function-mesh/api/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/spec"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *FunctionMeshReconciler) ObserveFunctionMesh(ctx context.Context, req ctrl.Request,
	mesh *v1alpha1.FunctionMesh) error {
	// TODO update deleted function status
	if err := r.observeFunctions(ctx, mesh); err != nil {
		return err
	}

	if err := r.observeSources(ctx, mesh); err != nil {
		return err
	}

	if err := r.observeSinks(ctx, mesh); err != nil {
		return err
	}

	return nil
}

func (r *FunctionMeshReconciler) observeFunctions(ctx context.Context, mesh *v1alpha1.FunctionMesh) error {
	for _, functionSpec := range mesh.Spec.Functions {
		// present the original name to use in Status, but underlying use the complete-name
		condition, ok := mesh.Status.FunctionConditions[functionSpec.Name]
		if !ok {
			mesh.Status.FunctionConditions[functionSpec.Name] = v1alpha1.ResourceCondition{
				Condition: v1alpha1.FunctionReady,
				Status:    metav1.ConditionFalse,
				Action:    v1alpha1.Create,
			}
			continue
		}

		function := &v1alpha1.Function{}
		err := r.Get(ctx, types.NamespacedName{
			Namespace: mesh.Namespace,
			Name:      makeComponentName(mesh.Name, functionSpec.Name),
		}, function)
		if err != nil {
			if errors.IsNotFound(err) {
				r.Log.Info("function is not ready", "name", functionSpec.Name)
				continue
			}
			return err
		}

		if function.Status.Conditions[v1alpha1.StatefulSet].Status == metav1.ConditionTrue &&
			function.Status.Conditions[v1alpha1.Service].Status == metav1.ConditionTrue {
			condition.Action = v1alpha1.NoAction
			condition.Status = metav1.ConditionTrue
			mesh.Status.FunctionConditions[functionSpec.Name] = condition
		} else {
			// function created but subcomponents not ready, we need to wait
			condition.Action = v1alpha1.Wait
			mesh.Status.FunctionConditions[functionSpec.Name] = condition
		}
	}

	return nil
}

func (r *FunctionMeshReconciler) observeSources(ctx context.Context, mesh *v1alpha1.FunctionMesh) error {
	for _, sourceSpec := range mesh.Spec.Sources {
		// present the original name to use in Status, but underlying use the complete-name
		condition, ok := mesh.Status.SourceConditions[sourceSpec.Name]
		if !ok {
			mesh.Status.SourceConditions[sourceSpec.Name] = v1alpha1.ResourceCondition{
				Condition: v1alpha1.SourceReady,
				Status:    metav1.ConditionFalse,
				Action:    v1alpha1.Create,
			}
			continue
		}

		source := &v1alpha1.Source{}
		err := r.Get(ctx, types.NamespacedName{
			Namespace: mesh.Namespace,
			Name:      makeComponentName(mesh.Name, sourceSpec.Name),
		}, source)
		if err != nil {
			if errors.IsNotFound(err) {
				r.Log.Info("source is not ready", "name", sourceSpec.Name)
				continue
			}
			return err
		}

		if source.Status.Conditions[v1alpha1.StatefulSet].Status == metav1.ConditionTrue &&
			source.Status.Conditions[v1alpha1.Service].Status == metav1.ConditionTrue {
			condition.Action = v1alpha1.NoAction
			condition.Status = metav1.ConditionTrue
			mesh.Status.SourceConditions[sourceSpec.Name] = condition
		} else {
			// function created but subcomponents not ready, we need to wait
			condition.Action = v1alpha1.Wait
			mesh.Status.SourceConditions[sourceSpec.Name] = condition
		}
	}

	return nil
}

func (r *FunctionMeshReconciler) observeSinks(ctx context.Context, mesh *v1alpha1.FunctionMesh) error {
	for _, sinkSpec := range mesh.Spec.Sinks {
		// present the original name to use in Status, but underlying use the complete-name
		condition, ok := mesh.Status.SinkConditions[sinkSpec.Name]
		if !ok {
			mesh.Status.SinkConditions[sinkSpec.Name] = v1alpha1.ResourceCondition{
				Condition: v1alpha1.SinkReady,
				Status:    metav1.ConditionFalse,
				Action:    v1alpha1.Create,
			}
			continue
		}

		sink := &v1alpha1.Sink{}
		err := r.Get(ctx, types.NamespacedName{
			Namespace: mesh.Namespace,
			Name:      makeComponentName(mesh.Name, sinkSpec.Name),
		}, sink)
		if err != nil {
			if errors.IsNotFound(err) {
				r.Log.Info("sink is not ready", "name", sinkSpec.Name)
				continue
			}
			return err
		}

		if sink.Status.Conditions[v1alpha1.StatefulSet].Status == metav1.ConditionTrue &&
			sink.Status.Conditions[v1alpha1.Service].Status == metav1.ConditionTrue {
			condition.Action = v1alpha1.NoAction
			condition.Status = metav1.ConditionTrue
			mesh.Status.SinkConditions[sinkSpec.Name] = condition
		} else {
			// function created but subcomponents not ready, we need to wait
			condition.Action = v1alpha1.Wait
			mesh.Status.SinkConditions[sinkSpec.Name] = condition
		}
	}

	return nil
}

func (r *FunctionMeshReconciler) UpdateFunctionMesh(ctx context.Context, req ctrl.Request,
	mesh *v1alpha1.FunctionMesh) error {
	for _, functionSpec := range mesh.Spec.Functions {
		condition := mesh.Status.FunctionConditions[functionSpec.Name]
		if condition.Status == metav1.ConditionTrue {
			continue
		}

		function := spec.MakeFunctionComponent(makeComponentName(mesh.Name, functionSpec.Name), mesh, &functionSpec)
		if err := r.HandleAction(ctx, function, condition.Action); err != nil {
			r.Log.Error(err, "failed to handle function", "name", functionSpec.Name, "action", condition.Action)
			return err
		}
	}

	for _, sourceSpec := range mesh.Spec.Sources {
		condition := mesh.Status.SourceConditions[sourceSpec.Name]
		if condition.Status == metav1.ConditionTrue {
			continue
		}
		source := spec.MakeSourceComponent(makeComponentName(mesh.Name, sourceSpec.Name), mesh, &sourceSpec)
		if err := r.HandleAction(ctx, source, condition.Action); err != nil {
			r.Log.Error(err, "failed to handle soure", "name", sourceSpec.Name, "action", condition.Action)
			return err
		}
	}

	for _, sinkSpec := range mesh.Spec.Sinks {
		condition := mesh.Status.SinkConditions[sinkSpec.Name]
		if condition.Status == metav1.ConditionTrue {
			continue
		}
		source := spec.MakeSinkComponent(makeComponentName(mesh.Name, sinkSpec.Name), mesh, &sinkSpec)
		if err := r.HandleAction(ctx, source, condition.Action); err != nil {
			r.Log.Error(err, "failed to handle sink", "name", sinkSpec.Name, "action", condition.Action)
			return err
		}
	}

	return nil
}

func (r *FunctionMeshReconciler) HandleAction(ctx context.Context, obj client.Object,
	action v1alpha1.ReconcileAction) error {
	switch action {
	case v1alpha1.Create:
		if err := r.Create(ctx, obj); err != nil {
			return err
		}
	case v1alpha1.Update:
		if err := r.Update(ctx, obj); err != nil {
			return err
		}
	case v1alpha1.Wait:
		// do nothing
	}
	return nil
}

func makeComponentName(prefix, name string) string {
	return prefix + "-" + name
}
