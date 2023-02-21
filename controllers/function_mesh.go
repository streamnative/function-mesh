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
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	computeapi "github.com/streamnative/function-mesh/api/compute/v1alpha2"
	"github.com/streamnative/function-mesh/controllers/spec"
)

func (r *FunctionMeshReconciler) ObserveFunctionMesh(
	ctx context.Context, mesh *computeapi.FunctionMesh, helper ReconciliationHelper) error {
	if err := r.observeFunctions(ctx, mesh, helper); err != nil {
		return err
	}
	if err := r.observeSources(ctx, mesh, helper); err != nil {
		return err
	}
	if err := r.observeSinks(ctx, mesh, helper); err != nil {
		return err
	}
	return nil
}

func (r *FunctionMeshReconciler) observeFunctions(
	ctx context.Context, mesh *computeapi.FunctionMesh, helper ReconciliationHelper) error {

	if len(mesh.Spec.Functions) == 0 {
		helper.GetState().FunctionState = "ready"
		return nil
	}

	desiredFunctions := map[string]bool{}
	for _, function := range mesh.Spec.Functions {
		desiredFunctions[makeComponentName(mesh.Name, function.Name)] = true
	}

	list := &computeapi.FunctionList{}
	if err := r.List(ctx, list, client.InNamespace(mesh.Namespace), client.MatchingLabels{
		"compute.functionmesh.io/part-of": mesh.Name,
	}); err != nil {
		return fmt.Errorf("list functions [%w]", err)
	}
	fmt.Println("11111111111111111111")
	fmt.Println(list)
	unreadyFunctions := []string{}
	for i := 0; i < len(list.Items); i++ {
		instance := &list.Items[i]
		if meta.IsStatusConditionFalse(instance.Status.Conditions, ResourceReady) {
			unreadyFunctions = append(unreadyFunctions, instance.Name)
		}
		if _, exist := desiredFunctions[instance.GetName()]; !exist {
			helper.GetState().SetOrphanedFunction(instance)
		}
	}

	helper.GetState().FunctionState = "unready"
	if len(mesh.Spec.Functions) <= len(list.Items) {
		helper.GetState().FunctionState = "created"
	}
	if !r.isGenerationChanged(mesh) && len(unreadyFunctions) == 0 {
		helper.GetState().FunctionState = "ready"
	}
	return nil
}

func (r *FunctionMeshReconciler) observeSources(
	ctx context.Context, mesh *computeapi.FunctionMesh, helper ReconciliationHelper) error {

	if len(mesh.Spec.Sources) == 0 {
		helper.GetState().SourceState = "ready"
		return nil
	}

	desiredSource := map[string]bool{}
	for _, source := range mesh.Spec.Sources {
		desiredSource[makeComponentName(mesh.Name, source.Name)] = true
	}

	list := &computeapi.SourceList{}
	if err := r.List(ctx, list, client.InNamespace(mesh.Namespace), client.MatchingLabels{
		"compute.functionmesh.io/part-of": mesh.Name,
	}); err != nil {
		return fmt.Errorf("list sources [%w]", err)
	}
	unreadySources := []string{}
	for i := 0; i < len(list.Items); i++ {
		instance := &list.Items[i]
		if meta.IsStatusConditionFalse(instance.Status.Conditions, ResourceReady) {
			unreadySources = append(unreadySources, instance.Name)
		}
		if _, exist := desiredSource[instance.GetName()]; !exist {
			helper.GetState().SetOrphanedSource(instance)
		}
	}

	helper.GetState().SourceState = "unready"
	if len(mesh.Spec.Sources) <= len(list.Items) {
		helper.GetState().SourceState = "created"
	}
	if !r.isGenerationChanged(mesh) && len(unreadySources) == 0 {
		helper.GetState().SourceState = "ready"
	}
	return nil
}

func (r *FunctionMeshReconciler) observeSinks(
	ctx context.Context, mesh *computeapi.FunctionMesh, helper ReconciliationHelper) error {

	if len(mesh.Spec.Sinks) == 0 {
		helper.GetState().SinkState = "ready"
		return nil
	}

	desiredSink := map[string]bool{}
	for _, sink := range mesh.Spec.Sinks {
		desiredSink[makeComponentName(mesh.Name, sink.Name)] = true
	}

	list := &computeapi.SinkList{}
	if err := r.List(ctx, list, client.InNamespace(mesh.Namespace), client.MatchingLabels{
		"compute.functionmesh.io/part-of": mesh.Name,
	}); err != nil {
		return fmt.Errorf("list sinks [%w]", err)
	}
	unreadySinks := []string{}
	for i := 0; i < len(list.Items); i++ {
		instance := &list.Items[i]
		if meta.IsStatusConditionFalse(instance.Status.Conditions, ResourceReady) {
			unreadySinks = append(unreadySinks, instance.Name)
		}
		if _, exist := desiredSink[instance.GetName()]; !exist {
			helper.GetState().SetOrphanedSink(instance)
		}
	}

	helper.GetState().SinkState = "unready"
	if len(mesh.Spec.Sinks) <= len(list.Items) {
		helper.GetState().SinkState = "created"
	}
	if !r.isGenerationChanged(mesh) && len(unreadySinks) == 0 {
		helper.GetState().SinkState = "ready"
	}
	return nil
}

func (r *FunctionMeshReconciler) ReconcileFunctionMesh(
	ctx context.Context, mesh *computeapi.FunctionMesh, helper ReconciliationHelper) error {
	if err := r.reconcileFunctions(ctx, mesh, helper); err != nil {
		return err
	}
	if err := r.reconcileSources(ctx, mesh, helper); err != nil {
		return err
	}
	if err := r.reconcileSinks(ctx, mesh, helper); err != nil {
		return err
	}
	return nil
}

func (r *FunctionMeshReconciler) reconcileFunctions(
	ctx context.Context, mesh *computeapi.FunctionMesh, helper ReconciliationHelper) error {

	if len(helper.GetState().OrphanedFunctionInstances) > 0 {
		for i := 0; i < len(helper.GetState().OrphanedFunctionInstances); i++ {
			instance := &helper.GetState().OrphanedFunctionInstances[i]
			if err := r.Delete(ctx, instance); err != nil {
				if errors.IsNotFound(err) {
					continue
				}
				return fmt.Errorf("error deleting function [%w]", err)
			}
		}
	}

	if helper.GetState().FunctionState == "ready" {
		return nil
	}

	for _, function := range mesh.Spec.Functions {
		desiredFunction := spec.MakeFunctionComponent(makeComponentName(mesh.Name, function.Name), mesh, &function)
		desiredFunctionSpec := desiredFunction.Spec
		if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredFunction, func() error {
			// function mutate logic
			desiredFunction.Spec = desiredFunctionSpec
			return nil
		}); err != nil {
			return fmt.Errorf("error creating or updating function [%w]", err)
		}
	}
	helper.GetState().FunctionState = "created"
	return nil
}

func (r *FunctionMeshReconciler) reconcileSources(
	ctx context.Context, mesh *computeapi.FunctionMesh, helper ReconciliationHelper) error {

	if len(helper.GetState().OrphanedSourceInstances) > 0 {
		for i := 0; i < len(helper.GetState().OrphanedSourceInstances); i++ {
			instance := &helper.GetState().OrphanedSourceInstances[i]
			if err := r.Delete(ctx, instance); err != nil {
				if errors.IsNotFound(err) {
					continue
				}
				return fmt.Errorf("error deleting source [%w]", err)
			}
		}
	}

	if helper.GetState().SourceState == "ready" {
		return nil
	}

	for _, source := range mesh.Spec.Sources {
		desiredSource := spec.MakeSourceComponent(makeComponentName(mesh.Name, source.Name), mesh, &source)
		desiredSourceSpec := desiredSource.Spec
		if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredSource, func() error {
			// source mutate logic
			desiredSource.Spec = desiredSourceSpec
			return nil
		}); err != nil {
			return fmt.Errorf("error creating or updating source [%w]", err)
		}
	}
	helper.GetState().SourceState = "created"
	return nil
}

func (r *FunctionMeshReconciler) reconcileSinks(
	ctx context.Context, mesh *computeapi.FunctionMesh, helper ReconciliationHelper) error {

	if len(helper.GetState().OrphanedSinkInstances) > 0 {
		for i := 0; i < len(helper.GetState().OrphanedSinkInstances); i++ {
			instance := &helper.GetState().OrphanedSinkInstances[i]
			if err := r.Delete(ctx, instance); err != nil {
				if errors.IsNotFound(err) {
					continue
				}
				return fmt.Errorf("error deleting sink [%w]", err)
			}
		}
	}

	if helper.GetState().SinkState == "ready" {
		return nil
	}

	for _, sink := range mesh.Spec.Sinks {
		desiredSink := spec.MakeSinkComponent(makeComponentName(mesh.Name, sink.Name), mesh, &sink)
		desiredSinkSpec := desiredSink.Spec
		if _, err := ctrl.CreateOrUpdate(ctx, r.Client, desiredSink, func() error {
			// sink mutate logic
			desiredSink.Spec = desiredSinkSpec
			return nil
		}); err != nil {
			return fmt.Errorf("error creating or updating sink [%w]", err)
		}
	}
	helper.GetState().SinkState = "created"
	return nil
}

func makeComponentName(prefix, name string) string {
	return prefix + "-" + name
}

func (r *FunctionMeshReconciler) isGenerationChanged(mesh *computeapi.FunctionMesh) bool {
	// if the generation has not changed, we do not need to update the component
	return mesh.Generation != mesh.Status.ObservedGeneration
}
