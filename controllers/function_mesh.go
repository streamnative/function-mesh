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
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/spec"
)

func (r *FunctionMeshReconciler) ObserveFunctionMesh(ctx context.Context, req ctrl.Request,
	mesh *v1alpha1.FunctionMesh) error {
	if err := r.observeFunctions(ctx, mesh); err != nil {
		return err
	}

	if err := r.observeSources(ctx, mesh); err != nil {
		return err
	}

	if err := r.observeSinks(ctx, mesh); err != nil {
		return err
	}

	// Observation only
	r.observeMeshes(mesh)

	return nil
}

func (r *FunctionMeshReconciler) observeFunctions(ctx context.Context, mesh *v1alpha1.FunctionMesh) error {
	orphanedFunctions := map[string]bool{}

	if len(mesh.Status.FunctionConditions) > 0 {
		for functionName := range mesh.Status.FunctionConditions {
			orphanedFunctions[functionName] = true
		}
	}

	for _, functionSpec := range mesh.Spec.Functions {
		delete(orphanedFunctions, functionSpec.Name)

		function := &v1alpha1.Function{}
		err := r.Get(ctx, types.NamespacedName{
			Namespace: mesh.Namespace,
			Name:      makeComponentName(mesh.Name, functionSpec.Name),
		}, function)
		if err != nil {
			if errors.IsNotFound(err) {
				r.Log.Info("function is not ready", "namespace", functionSpec.Namespace, "name", functionSpec.Name)
				mesh.SetFunctionCondition(functionSpec.Name, v1alpha1.CreateCondition(
					v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
					"function is not ready yet..."))
				continue
			}
			mesh.SetFunctionCondition(functionSpec.Name, v1alpha1.CreateCondition(
				v1alpha1.Error, metav1.ConditionFalse, v1alpha1.FunctionError,
				fmt.Sprintf("failed to fetch function: %v", err)))
			return err
		}

		if checkComponentsReady(function.Status.Conditions, functionSpec.MaxReplicas != nil, functionSpec.Pod.VPA != nil) {
			mesh.SetFunctionCondition(functionSpec.Name, v1alpha1.CreateCondition(
				v1alpha1.Ready, metav1.ConditionTrue, v1alpha1.FunctionIsReady, ""))
		} else {
			// function created but subcomponents not ready, we need to wait
			mesh.SetFunctionCondition(functionSpec.Name, v1alpha1.CreateCondition(
				v1alpha1.Created, metav1.ConditionFalse, v1alpha1.PendingCreation,
				"function is created, but needs to wait to be ready..."))
		}
	}

	for functionName, isOrphaned := range orphanedFunctions {
		if isOrphaned {
			mesh.SetFunctionCondition(functionName, v1alpha1.CreateCondition(
				v1alpha1.Orphaned, metav1.ConditionFalse, v1alpha1.PendingTermination,
				fmt.Sprintf("function needs to be terminated: %s", functionName)))
		}
	}

	return nil
}

func (r *FunctionMeshReconciler) observeSources(ctx context.Context, mesh *v1alpha1.FunctionMesh) error {
	orphanedSources := map[string]bool{}

	if len(mesh.Status.SourceConditions) > 0 {
		for sourceName := range mesh.Status.SourceConditions {
			orphanedSources[sourceName] = true
		}
	}

	for _, sourceSpec := range mesh.Spec.Sources {
		delete(orphanedSources, sourceSpec.Name)

		source := &v1alpha1.Source{}
		err := r.Get(ctx, types.NamespacedName{
			Namespace: mesh.Namespace,
			Name:      makeComponentName(mesh.Name, sourceSpec.Name),
		}, source)
		if err != nil {
			if errors.IsNotFound(err) {
				r.Log.Info("source is not ready", "namespace", sourceSpec.Namespace, "name", sourceSpec.Name)
				mesh.SetSourceCondition(sourceSpec.Name, v1alpha1.CreateCondition(
					v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
					"source is not ready yet..."))
				continue
			}
			mesh.SetFunctionCondition(sourceSpec.Name, v1alpha1.CreateCondition(
				v1alpha1.Error, metav1.ConditionFalse, v1alpha1.SourceError,
				fmt.Sprintf("failed to fetch source: %v", err)))
			return err
		}

		if checkComponentsReady(source.Status.Conditions, sourceSpec.MaxReplicas != nil, sourceSpec.Pod.VPA != nil) {
			mesh.SetSourceCondition(sourceSpec.Name, v1alpha1.CreateCondition(
				v1alpha1.Ready, metav1.ConditionTrue, v1alpha1.SourceIsReady, ""))
		} else {
			// source created but subcomponents not ready, we need to wait
			mesh.SetSourceCondition(sourceSpec.Name, v1alpha1.CreateCondition(
				v1alpha1.Created, metav1.ConditionFalse, v1alpha1.PendingCreation,
				"source is created, but needs to wait to be ready..."))
		}
	}

	for sourceName, isOrphaned := range orphanedSources {
		if isOrphaned {
			mesh.SetSourceCondition(sourceName, v1alpha1.CreateCondition(
				v1alpha1.Orphaned, metav1.ConditionFalse, v1alpha1.PendingTermination,
				fmt.Sprintf("source needs to be terminated: %s", sourceName)))
		}
	}
	return nil
}

func (r *FunctionMeshReconciler) observeSinks(ctx context.Context, mesh *v1alpha1.FunctionMesh) error {
	orphanedSinks := map[string]bool{}

	if len(mesh.Status.SinkConditions) > 0 {
		for sinkName := range mesh.Status.SinkConditions {
			orphanedSinks[sinkName] = true
		}
	}

	for _, sinkSpec := range mesh.Spec.Sinks {
		delete(orphanedSinks, sinkSpec.Name)

		sink := &v1alpha1.Sink{}
		err := r.Get(ctx, types.NamespacedName{
			Namespace: mesh.Namespace,
			Name:      makeComponentName(mesh.Name, sinkSpec.Name),
		}, sink)
		if err != nil {
			if errors.IsNotFound(err) {
				r.Log.Info("sink is not ready", "namespace", sinkSpec.Namespace, "name", sinkSpec.Name)
				mesh.SetSinkCondition(sinkSpec.Name, v1alpha1.CreateCondition(
					v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
					"sink is not ready yet..."))
				continue
			}
			mesh.SetSinkCondition(sinkSpec.Name, v1alpha1.CreateCondition(
				v1alpha1.Error, metav1.ConditionFalse, v1alpha1.SinkError,
				fmt.Sprintf("failed to fetch sink: %v", err)))
			return err
		}

		if checkComponentsReady(sink.Status.Conditions, sinkSpec.MaxReplicas != nil, sinkSpec.Pod.VPA != nil) {
			mesh.SetSinkCondition(sinkSpec.Name, v1alpha1.CreateCondition(
				v1alpha1.Ready, metav1.ConditionTrue, v1alpha1.SinkIsReady, ""))
		} else {
			// sink created but subcomponents not ready, we need to wait
			mesh.SetSinkCondition(sinkSpec.Name, v1alpha1.CreateCondition(
				v1alpha1.Created, metav1.ConditionFalse, v1alpha1.PendingCreation,
				"sink is created, but needs to wait to be ready..."))
		}
	}

	for sinkName, isOrphaned := range orphanedSinks {
		if isOrphaned {
			mesh.SetSinkCondition(sinkName, v1alpha1.CreateCondition(
				v1alpha1.Orphaned, metav1.ConditionFalse, v1alpha1.PendingTermination,
				fmt.Sprintf("sink needs to be terminated: %s", sinkName)))
		}
	}
	return nil
}

func (r *FunctionMeshReconciler) observeMeshes(mesh *v1alpha1.FunctionMesh) {
	for _, cond := range mesh.Status.FunctionConditions {
		if cond.Type == string(v1alpha1.Ready) && cond.Status == metav1.ConditionTrue {
			continue
		}
		mesh.SetCondition(v1alpha1.CreateCondition(
			v1alpha1.Ready, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"function is not ready yet..."))
		return
	}

	for _, cond := range mesh.Status.SinkConditions {
		if cond.Type == string(v1alpha1.Ready) && cond.Status == metav1.ConditionTrue {
			continue
		}
		mesh.SetCondition(v1alpha1.CreateCondition(
			v1alpha1.Ready, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"function is not ready yet..."))
		return
	}

	for _, cond := range mesh.Status.SourceConditions {
		if cond.Type == string(v1alpha1.Ready) && cond.Status == metav1.ConditionTrue {
			continue
		}
		mesh.SetCondition(v1alpha1.CreateCondition(
			v1alpha1.Ready, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"function is not ready yet..."))
		return
	}

	mesh.SetCondition(v1alpha1.CreateCondition(
		v1alpha1.Ready, metav1.ConditionTrue, v1alpha1.MeshIsReady, ""))
}

func (r *FunctionMeshReconciler) UpdateFunctionMesh(ctx context.Context, req ctrl.Request,
	mesh *v1alpha1.FunctionMesh) error {
	newGeneration := mesh.Generation != mesh.Status.ObservedGeneration

	for _, functionSpec := range mesh.Spec.Functions {
		condition := mesh.Status.FunctionConditions[functionSpec.Name]
		if !newGeneration &&
			condition.Status == metav1.ConditionTrue &&
			condition.Type == string(v1alpha1.Ready) {
			continue
		}
		function := spec.MakeFunctionComponent(makeComponentName(mesh.Name, functionSpec.Name), mesh, &functionSpec)
		if err := r.CreateOrUpdateFunction(ctx, function, function.Spec); err != nil {
			r.Log.Error(err, "failed to handle function", "namespace", functionSpec.Namespace, "name", functionSpec.Name)
			mesh.SetFunctionCondition(functionSpec.Name, v1alpha1.CreateCondition(
				v1alpha1.Error, metav1.ConditionFalse, v1alpha1.ErrorCreatingFunction,
				fmt.Sprintf("error create or update function for mesh: %v", err)))
			return err
		}
		mesh.SetFunctionCondition(functionSpec.Name, v1alpha1.CreateCondition(
			v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"creating or updating function for mesh..."))
	}

	for _, sourceSpec := range mesh.Spec.Sources {
		condition := mesh.Status.SourceConditions[sourceSpec.Name]
		if !newGeneration &&
			condition.Status == metav1.ConditionTrue &&
			condition.Type == string(v1alpha1.Ready) {
			continue
		}
		source := spec.MakeSourceComponent(makeComponentName(mesh.Name, sourceSpec.Name), mesh, &sourceSpec)
		if err := r.CreateOrUpdateSource(ctx, source, source.Spec); err != nil {
			r.Log.Error(err, "failed to handle soure", "namespace", sourceSpec.Namespace, "name", sourceSpec.Name)
			mesh.SetSourceCondition(sourceSpec.Name, v1alpha1.CreateCondition(
				v1alpha1.Error, metav1.ConditionFalse, v1alpha1.ErrorCreatingSource,
				fmt.Sprintf("error create or update source for mesh: %v", err)))
			return err
		}
		mesh.SetSourceCondition(sourceSpec.Name, v1alpha1.CreateCondition(
			v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"creating or updating source for mesh..."))
	}

	for _, sinkSpec := range mesh.Spec.Sinks {
		condition := mesh.Status.SinkConditions[sinkSpec.Name]
		if !newGeneration &&
			condition.Status == metav1.ConditionTrue &&
			condition.Type == string(v1alpha1.Ready) {
			continue
		}
		sink := spec.MakeSinkComponent(makeComponentName(mesh.Name, sinkSpec.Name), mesh, &sinkSpec)
		if err := r.CreateOrUpdateSink(ctx, sink, sink.Spec); err != nil {
			r.Log.Error(err, "failed to handle sink", "namespace", sinkSpec.Namespace, "name", sinkSpec.Name)
			mesh.SetSinkCondition(sinkSpec.Name, v1alpha1.CreateCondition(
				v1alpha1.Error, metav1.ConditionFalse, v1alpha1.ErrorCreatingSink,
				fmt.Sprintf("error create or update sink for mesh: %v", err)))
			return err
		}
		mesh.SetSinkCondition(sinkSpec.Name, v1alpha1.CreateCondition(
			v1alpha1.Pending, metav1.ConditionFalse, v1alpha1.PendingCreation,
			"creating or updating sink for mesh..."))
	}

	// handle logic for cleaning up orphaned subcomponents
	if len(mesh.Spec.Functions) != len(mesh.Status.FunctionConditions) {
		for functionName, condition := range mesh.Status.FunctionConditions {
			if condition.Type == string(v1alpha1.Orphaned) {
				function := &v1alpha1.Function{}
				function.Namespace = mesh.Namespace
				function.Name = makeComponentName(mesh.Name, functionName)
				if err := r.Delete(ctx, function); err != nil && !errors.IsNotFound(err) {
					mesh.SetFunctionCondition(functionName, v1alpha1.CreateCondition(
						v1alpha1.Orphaned, metav1.ConditionFalse, v1alpha1.FunctionError,
						fmt.Sprintf("failed to delete function: %v", err)))
					return err
				}
				delete(mesh.Status.FunctionConditions, functionName)
			}
		}
	}

	if len(mesh.Spec.Sources) != len(mesh.Status.SourceConditions) {
		for sourceName, condition := range mesh.Status.SourceConditions {
			if condition.Type == string(v1alpha1.Orphaned) {
				source := &v1alpha1.Source{}
				source.Namespace = mesh.Namespace
				source.Name = makeComponentName(mesh.Name, sourceName)
				if err := r.Delete(ctx, source); err != nil && !errors.IsNotFound(err) {
					mesh.SetSourceCondition(sourceName, v1alpha1.CreateCondition(
						v1alpha1.Orphaned, metav1.ConditionFalse, v1alpha1.SourceError,
						fmt.Sprintf("failed to delete source: %v", err)))
					return err
				}
				delete(mesh.Status.SourceConditions, sourceName)
			}
		}
	}

	if len(mesh.Spec.Sinks) != len(mesh.Status.SinkConditions) {
		for sinkName, condition := range mesh.Status.SinkConditions {
			if condition.Type == string(v1alpha1.Orphaned) {
				sink := &v1alpha1.Sink{}
				sink.Namespace = mesh.Namespace
				sink.Name = makeComponentName(mesh.Name, sinkName)
				if err := r.Delete(ctx, sink); err != nil && !errors.IsNotFound(err) {
					mesh.SetSinkCondition(sinkName, v1alpha1.CreateCondition(
						v1alpha1.Orphaned, metav1.ConditionFalse, v1alpha1.SinkError,
						fmt.Sprintf("failed to delete sink: %v", err)))
					return err
				}
				delete(mesh.Status.SinkConditions, sinkName)
			}
		}
	}

	return nil
}

func (r *FunctionMeshReconciler) CreateOrUpdateFunction(ctx context.Context, function *v1alpha1.Function, functionSpec v1alpha1.FunctionSpec) error {
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, function, func() error {
		// function mutate logic
		function.Spec = functionSpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error create or update function", "namespace", function.Namespace, "name", function.Name)
		return err
	}
	return nil
}

func (r *FunctionMeshReconciler) CreateOrUpdateSink(ctx context.Context, sink *v1alpha1.Sink, sinkSpec v1alpha1.SinkSpec) error {
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, sink, func() error {
		// sink mutate logic
		sink.Spec = sinkSpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error create or update sink", "namespace", sink.Namespace, "name", sink.Name)
		return err
	}
	return nil
}

func (r *FunctionMeshReconciler) CreateOrUpdateSource(ctx context.Context, source *v1alpha1.Source, sourceSpec v1alpha1.SourceSpec) error {
	if _, err := ctrl.CreateOrUpdate(ctx, r.Client, source, func() error {
		// source mutate logic
		source.Spec = sourceSpec
		return nil
	}); err != nil {
		r.Log.Error(err, "error create or update source", "namespace", source.Namespace, "name", source.Name)
		return err
	}
	return nil
}

func makeComponentName(prefix, name string) string {
	return prefix + "-" + name
}

func checkComponentsReady(conditions map[v1alpha1.Component]metav1.Condition, isHPAEnabled bool, isVPAEnabled bool) bool {
	ready := conditions[v1alpha1.StatefulSet].Type == string(v1alpha1.Ready) &&
		conditions[v1alpha1.StatefulSet].Status == metav1.ConditionTrue &&
		conditions[v1alpha1.Service].Type == string(v1alpha1.Ready) &&
		conditions[v1alpha1.Service].Status == metav1.ConditionTrue
	if isHPAEnabled {
		hpaReady := conditions[v1alpha1.HPA].Type == string(v1alpha1.Ready) &&
			conditions[v1alpha1.HPA].Status == metav1.ConditionTrue
		ready = ready && hpaReady
	}
	if isVPAEnabled {
		vpaReady := conditions[v1alpha1.VPA].Type == string(v1alpha1.Ready) &&
			conditions[v1alpha1.VPA].Status == metav1.ConditionTrue
		ready = ready && vpaReady
	}
	return ready
}
