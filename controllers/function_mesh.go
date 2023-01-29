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
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/spec"
)

func (r *FunctionMeshReconciler) ObserveFunctionMesh(ctx context.Context, mesh *v1alpha1.FunctionMesh) error {
	defer mesh.SaveStatus(ctx, r.Log, r.Client)

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
	unreadyFunctions := []string{}

	defer func() {
		if len(unreadyFunctions) > 0 {
			mesh.Status.FunctionConditions.Ready = false
		}
	}()

	if len(mesh.Status.FunctionConditions.Status) > 0 {
		for name, _ := range mesh.Status.FunctionConditions.Status {
			orphanedFunctions[name] = true
		}
	}

	for _, functionSpec := range mesh.Spec.Functions {
		delete(orphanedFunctions, functionSpec.Name)

		// present the original name to use in Status, but underlying use the complete-name
		function := &v1alpha1.Function{}
		err := r.Get(ctx, types.NamespacedName{
			Namespace: mesh.Namespace,
			Name:      makeComponentName(mesh.Name, functionSpec.Name),
		}, function)
		if err != nil {
			if errors.IsNotFound(err) {
				r.Log.Info("function is not ready", "name", functionSpec.Name)
				unreadyFunctions = append(unreadyFunctions, functionSpec.Name)
				continue
			}
			mesh.SetComponentCondition(v1alpha1.FunctionComponent, functionSpec.Name, v1alpha1.Error)
			unreadyFunctions = append(unreadyFunctions, functionSpec.Name)
			return err
		}

		if meta.IsStatusConditionTrue(function.Status.Conditions, string(v1alpha1.Ready)) {
			mesh.SetComponentCondition(v1alpha1.FunctionComponent, functionSpec.Name, v1alpha1.Ready)
			continue
		}
		unreadyFunctions = append(unreadyFunctions, functionSpec.Name)
	}

	for name, isOrphaned := range orphanedFunctions {
		if isOrphaned {
			mesh.SetComponentCondition(v1alpha1.FunctionComponent, name, v1alpha1.Orphaned)
		}
	}
	return nil
}

func (r *FunctionMeshReconciler) observeSources(ctx context.Context, mesh *v1alpha1.FunctionMesh) error {
	orphanedSources := map[string]bool{}
	unreadySources := []string{}

	defer func() {
		if len(unreadySources) > 0 {
			mesh.Status.SourceConditions.Ready = false
		}
	}()

	if len(mesh.Status.SourceConditions.Status) > 0 {
		for name, _ := range mesh.Status.SourceConditions.Status {
			orphanedSources[name] = true
		}
	}

	for _, sourceSpec := range mesh.Spec.Sources {
		delete(orphanedSources, sourceSpec.Name)

		// present the original name to use in Status, but underlying use the complete-name
		source := &v1alpha1.Source{}
		err := r.Get(ctx, types.NamespacedName{
			Namespace: mesh.Namespace,
			Name:      makeComponentName(mesh.Name, sourceSpec.Name),
		}, source)
		if err != nil {
			if errors.IsNotFound(err) {
				r.Log.Info("source is not ready", "name", sourceSpec.Name)
				unreadySources = append(unreadySources, sourceSpec.Name)
				continue
			}
			mesh.SetComponentCondition(v1alpha1.SourceComponent, source.Name, v1alpha1.Error)
			unreadySources = append(unreadySources, sourceSpec.Name)
			return err
		}

		if meta.IsStatusConditionTrue(source.Status.Conditions, string(v1alpha1.Ready)) {
			mesh.SetComponentCondition(v1alpha1.SourceComponent, source.Name, v1alpha1.Ready)
			continue
		}
		unreadySources = append(unreadySources, sourceSpec.Name)
	}

	for name, isOrphaned := range orphanedSources {
		if isOrphaned {
			mesh.SetComponentCondition(v1alpha1.SourceComponent, name, v1alpha1.Orphaned)
		}
	}
	return nil
}

func (r *FunctionMeshReconciler) observeSinks(ctx context.Context, mesh *v1alpha1.FunctionMesh) error {
	orphanedSinks := map[string]bool{}
	unreadySinks := []string{}

	defer func() {
		if len(unreadySinks) > 0 {
			mesh.Status.SinkConditions.Ready = false
		}
	}()

	if len(mesh.Status.SinkConditions.Status) > 0 {
		for name, _ := range mesh.Status.SinkConditions.Status {
			orphanedSinks[name] = true
		}
	}

	for _, sinkSpec := range mesh.Spec.Sinks {
		delete(orphanedSinks, sinkSpec.Name)

		// present the original name to use in Status, but underlying use the complete-name
		sink := &v1alpha1.Sink{}
		err := r.Get(ctx, types.NamespacedName{
			Namespace: mesh.Namespace,
			Name:      makeComponentName(mesh.Name, sinkSpec.Name),
		}, sink)
		if err != nil {
			if errors.IsNotFound(err) {
				r.Log.Info("sink is not ready", "name", sinkSpec.Name)
				unreadySinks = append(unreadySinks, sinkSpec.Name)
				continue
			}
			mesh.SetComponentCondition(v1alpha1.SinkComponent, sink.Name, v1alpha1.Error)
			unreadySinks = append(unreadySinks, sinkSpec.Name)
			return err
		}

		if meta.IsStatusConditionTrue(sink.Status.Conditions, string(v1alpha1.Ready)) {
			mesh.SetComponentCondition(v1alpha1.SinkComponent, sink.Name, v1alpha1.Ready)
			continue
		}
		unreadySinks = append(unreadySinks, sinkSpec.Name)
	}

	for name, isOrphaned := range orphanedSinks {
		if isOrphaned {
			mesh.SetComponentCondition(v1alpha1.SinkComponent, name, v1alpha1.Orphaned)
		}
	}
	return nil
}

func (r *FunctionMeshReconciler) observeMeshes(mesh *v1alpha1.FunctionMesh) {
	if mesh.Status.FunctionConditions != nil && !mesh.Status.FunctionConditions.Ready {
		mesh.SetCondition(v1alpha1.Wait, metav1.ConditionTrue, v1alpha1.PendingCreation,
			"wait for sub functions to be ready")
		return
	}
	if mesh.Status.SinkConditions != nil && !mesh.Status.SinkConditions.Ready {
		mesh.SetCondition(v1alpha1.Wait, metav1.ConditionTrue, v1alpha1.PendingCreation,
			"wait for sub sinks to be ready")
		return
	}
	if mesh.Status.SourceConditions != nil && !mesh.Status.SourceConditions.Ready {
		mesh.SetCondition(v1alpha1.Wait, metav1.ConditionTrue, v1alpha1.PendingCreation,
			"wait for sub source to be ready")
		return
	}
	mesh.SetCondition(v1alpha1.Ready, metav1.ConditionTrue, v1alpha1.MeshIsReady, "")
}

func (r *FunctionMeshReconciler) UpdateFunctionMesh(ctx context.Context, mesh *v1alpha1.FunctionMesh) error {
	defer mesh.SaveStatus(ctx, r.Log, r.Client)

	if !r.checkIfMeshNeedUpdate(mesh) {
		return nil
	}

	for _, functionSpec := range mesh.Spec.Functions {
		condition := mesh.Status.FunctionConditions.Status[functionSpec.Name]
		if functionSpec.MaxReplicas != nil && condition == v1alpha1.Ready {
			continue
		}
		desiredFunction := spec.MakeFunctionComponent(makeComponentName(mesh.Name, functionSpec.Name), mesh, &functionSpec)
		desiredFunctionSpec := desiredFunction.Spec
		if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredFunction, func() error {
			// function mutate logic
			desiredFunction.Spec = desiredFunctionSpec
			return nil
		}); err != nil {
			r.Log.Error(err, "error creating or updating function",
				"namespace", mesh.Namespace, "name", mesh.Name,
				"function name", desiredFunction.Name)
			mesh.SetComponentCondition(v1alpha1.FunctionComponent, functionSpec.Name, v1alpha1.Error)
			mesh.SetCondition(v1alpha1.Error, metav1.ConditionTrue, v1alpha1.ErrorCreatingFunction,
				fmt.Sprintf("error creating or updating function: %v", err))
			return err
		}
		mesh.SetComponentCondition(v1alpha1.FunctionComponent, functionSpec.Name, v1alpha1.Wait)
	}

	for _, sourceSpec := range mesh.Spec.Sources {
		condition := mesh.Status.SourceConditions.Status[sourceSpec.Name]
		if sourceSpec.MaxReplicas != nil && condition == v1alpha1.Ready {
			continue
		}
		desiredSource := spec.MakeSourceComponent(makeComponentName(mesh.Name, sourceSpec.Name), mesh, &sourceSpec)
		desiredSourceSpec := desiredSource.Spec
		if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredSource, func() error {
			// source mutate logic
			desiredSource.Spec = desiredSourceSpec
			return nil
		}); err != nil {
			r.Log.Error(err, "error creating or updating source",
				"namespace", mesh.Namespace, "name", mesh.Name,
				"source name", desiredSource.Name)
			mesh.SetComponentCondition(v1alpha1.SourceComponent, sourceSpec.Name, v1alpha1.Error)
			mesh.SetCondition(v1alpha1.Error, metav1.ConditionTrue, v1alpha1.ErrorCreatingSource,
				fmt.Sprintf("error creating or updating source: %v", err))
			return err
		}
		mesh.SetComponentCondition(v1alpha1.SourceComponent, sourceSpec.Name, v1alpha1.Wait)
	}

	for _, sinkSpec := range mesh.Spec.Sinks {
		condition := mesh.Status.SinkConditions.Status[sinkSpec.Name]
		if sinkSpec.MaxReplicas != nil && condition == v1alpha1.Ready {
			continue
		}
		desiredSink := spec.MakeSinkComponent(makeComponentName(mesh.Name, sinkSpec.Name), mesh, &sinkSpec)
		desiredSinkSpec := desiredSink.Spec
		if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredSink, func() error {
			// sink mutate logic
			desiredSink.Spec = desiredSinkSpec
			return nil
		}); err != nil {
			r.Log.Error(err, "error creating or updating sink",
				"namespace", mesh.Namespace, "name", mesh.Name,
				"sink name", desiredSink.Name)
			mesh.SetComponentCondition(v1alpha1.SinkComponent, sinkSpec.Name, v1alpha1.Error)
			mesh.SetCondition(v1alpha1.Error, metav1.ConditionTrue, v1alpha1.ErrorCreatingSink,
				fmt.Sprintf("error creating or updating sink: %v", err))
			return err
		}
		mesh.SetComponentCondition(v1alpha1.SinkComponent, sinkSpec.Name, v1alpha1.Wait)
	}

	// handle logic for cleaning up orphaned subcomponents
	if len(mesh.Spec.Functions) != len(mesh.Status.FunctionConditions.Status) {
		for name, cond := range mesh.Status.FunctionConditions.Status {
			if cond == v1alpha1.Orphaned {
				// clean up the orphaned functions
				function := &v1alpha1.Function{}
				function.Namespace = mesh.Namespace
				function.Name = makeComponentName(mesh.Name, name)
				if err := r.Delete(ctx, function); err != nil && !errors.IsNotFound(err) {
					r.Log.Error(err, "error deleting orphaned function for mesh",
						"namespace", mesh.Namespace, "name", mesh.Name,
						"function name", function.Name)
					mesh.SetCondition(v1alpha1.Error, metav1.ConditionTrue, v1alpha1.FunctionError,
						fmt.Sprintf("error deleting orphaned function for mesh: %v", err))
					return err
				}
				delete(mesh.Status.FunctionConditions.Status, name)
			}
		}
	}

	if len(mesh.Spec.Sources) != len(mesh.Status.SourceConditions.Status) {
		for name, cond := range mesh.Status.SourceConditions.Status {
			if cond == v1alpha1.Orphaned {
				// clean up the orphaned sources
				source := &v1alpha1.Source{}
				source.Namespace = mesh.Namespace
				source.Name = makeComponentName(mesh.Name, name)
				if err := r.Delete(ctx, source); err != nil && !errors.IsNotFound(err) {
					r.Log.Error(err, "error deleting orphaned source for mesh",
						"namespace", mesh.Namespace, "name", mesh.Name,
						"source name", source.Name)
					mesh.SetCondition(v1alpha1.Error, metav1.ConditionTrue, v1alpha1.SourceError,
						fmt.Sprintf("error deleting orphaned source for mesh: %v", err))
					return err
				}
				delete(mesh.Status.SourceConditions.Status, name)
			}
		}
	}

	if len(mesh.Spec.Sinks) != len(mesh.Status.SinkConditions.Status) {
		for name, cond := range mesh.Status.SinkConditions.Status {
			if cond == v1alpha1.Orphaned {
				// clean up the orphaned sinks
				sink := &v1alpha1.Sink{}
				sink.Namespace = mesh.Namespace
				sink.Name = makeComponentName(mesh.Name, name)
				if err := r.Delete(ctx, sink); err != nil && !errors.IsNotFound(err) {
					r.Log.Error(err, "error deleting orphaned sink for mesh",
						"namespace", mesh.Namespace, "name", mesh.Name,
						"sink name", sink.Name)
					mesh.SetCondition(v1alpha1.Error, metav1.ConditionTrue, v1alpha1.SinkError,
						fmt.Sprintf("error deleting orphaned sink for mesh: %v", err))
					return err
				}
				delete(mesh.Status.SinkConditions.Status, name)
			}
		}
	}
	return nil
}

func makeComponentName(prefix, name string) string {
	return prefix + "-" + name
}

func (r *FunctionMeshReconciler) initializeMesh(mesh *v1alpha1.FunctionMesh) {
	// initialize function conditions
	if len(mesh.Spec.Functions) > 0 {
		mesh.Status.FunctionConditions = initializeComponentConditions(mesh.Status.FunctionConditions)
		for _, function := range mesh.Spec.Functions {
			if _, exist := mesh.Status.FunctionConditions.Status[function.Name]; exist {
				r.Log.Error(fmt.Errorf("the function %s already exists", function.Name),
					"the function will be overridden by the current specification.",
					"namespace", mesh.Namespace, "mesh name", mesh.Name,
					"function name", makeComponentName(mesh.Name, function.Name))
			}
			mesh.Status.FunctionConditions.Status[function.Name] = v1alpha1.Wait
		}
	} else {
		mesh.Status.FunctionConditions = nil
	}

	// initialize sink conditions
	if len(mesh.Spec.Sinks) > 0 {
		mesh.Status.SinkConditions = initializeComponentConditions(mesh.Status.SinkConditions)
		for _, sink := range mesh.Spec.Sinks {
			if _, exist := mesh.Status.SinkConditions.Status[sink.Name]; exist {
				r.Log.Error(fmt.Errorf("the sink %s already exists", sink.Name),
					"the sink will be overridden by the current specification.",
					"namespace", mesh.Namespace, "mesh name", mesh.Name,
					"sink name", makeComponentName(mesh.Name, sink.Name))
			}
			mesh.Status.SinkConditions.Status[sink.Name] = v1alpha1.Wait
		}
	} else {
		mesh.Status.SinkConditions = nil
	}

	// initialize source conditions
	if len(mesh.Spec.Sources) > 0 {
		mesh.Status.SourceConditions = initializeComponentConditions(mesh.Status.SourceConditions)
		for _, source := range mesh.Spec.Sources {
			if _, exist := mesh.Status.SourceConditions.Status[source.Name]; exist {
				r.Log.Error(fmt.Errorf("the source %s already exists", source.Name),
					"the source will be overridden by the current specification.",
					"namespace", mesh.Namespace, "mesh name", mesh.Name,
					"source name", makeComponentName(mesh.Name, source.Name))
			}
			mesh.Status.SourceConditions.Status[source.Name] = v1alpha1.Wait
		}
	} else {
		mesh.Status.SourceConditions = nil
	}
}

func initializeComponentConditions(componentCondition *v1alpha1.ComponentCondition) *v1alpha1.ComponentCondition {
	if componentCondition == nil {
		componentCondition = &v1alpha1.ComponentCondition{}
		componentCondition.Ready = false
	}
	if componentCondition.Status == nil {
		componentCondition.Status = map[string]v1alpha1.ResourceConditionType{}
	}
	return componentCondition
}

func (r *FunctionMeshReconciler) checkIfMeshNeedUpdate(mesh *v1alpha1.FunctionMesh) bool {
	if cond := meta.FindStatusCondition(mesh.Status.Conditions, string(v1alpha1.Ready)); cond != nil {
		if cond.ObservedGeneration != mesh.Generation {
			return true
		}
		if cond.Status == metav1.ConditionTrue {
			return false
		}
	}
	return true
}
