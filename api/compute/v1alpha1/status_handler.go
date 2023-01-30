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

package v1alpha1

import (
	"context"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ResourceConditionType indicates the available resource condition type
type ResourceConditionType string

const (
	// Error indicates the resource had an error
	Error ResourceConditionType = "Error"
	// Wait indicates the resource hasn't been created
	Wait ResourceConditionType = "Wait"
	// Orphaned indicates that the resource is marked for deletion but hasn't
	// been deleted yet
	Orphaned ResourceConditionType = "Orphaned"
	// Ready indicates the object is fully created
	Ready ResourceConditionType = "Ready"
	// FunctionReady indicates the Function is fully created
	FunctionReady ResourceConditionType = "FunctionReady"
	// SinkReady indicates the Sink is fully created
	SinkReady ResourceConditionType = "SinkReady"
	// SourceReady indicates the Source is fully created
	SourceReady ResourceConditionType = "SourceReady"
	// StatefulSetReady indicates the StatefulSet is fully created
	StatefulSetReady ResourceConditionType = "StatefulSetReady"
	// ServiceReady indicates the Service is fully created
	ServiceReady ResourceConditionType = "ServiceReady"
	// HPAReady indicates the HPA is fully created
	HPAReady ResourceConditionType = "HPAReady"
	// VPAReady indicates the VPA is fully created
	VPAReady ResourceConditionType = "VPAReady"
)

// ResourceConditionReason indicates the reason why the resource is in its current condition
type ResourceConditionReason string

const (
	ErrorCreatingStatefulSet ResourceConditionReason = "ErrorCreatingStatefulSet"
	ErrorCreatingFunction    ResourceConditionReason = "ErrorCreatingFunction"
	ErrorCreatingSource      ResourceConditionReason = "ErrorCreatingSource"
	ErrorCreatingSink        ResourceConditionReason = "ErrorCreatingSink"
	ErrorCreatingHPA         ResourceConditionReason = "ErrorCreatingHPA"
	ErrorCreatingVPA         ResourceConditionReason = "ErrorCreatingVPA"
	ErrorCreatingService     ResourceConditionReason = "ErrorCreatingService"
	StatefulSetError         ResourceConditionReason = "StatefulSetError"
	FunctionError            ResourceConditionReason = "FunctionError"
	SourceError              ResourceConditionReason = "SourceError"
	SinkError                ResourceConditionReason = "SinkError"
	ServiceError             ResourceConditionReason = "ServiceError"
	HPAError                 ResourceConditionReason = "HPAError"
	VPAError                 ResourceConditionReason = "VPAError"
	PendingCreation          ResourceConditionReason = "PendingCreation"
	ServiceIsReady           ResourceConditionReason = "ServiceIsReady"
	StatefulSetIsReady       ResourceConditionReason = "StatefulSetIsReady"
	HPAIsReady               ResourceConditionReason = "HPAIsReady"
	VPAIsReady               ResourceConditionReason = "VPAIsReady"
	FunctionIsReady          ResourceConditionReason = "FunctionIsReady"
	SourceIsReady            ResourceConditionReason = "SourceIsReady"
	SinkIsReady              ResourceConditionReason = "SinkIsReady"
	MeshIsReady              ResourceConditionReason = "MeshIsReady"
)

// SaveStatus will trigger Function object update to save the current status
// conditions
func (r *Function) SaveStatus(ctx context.Context, logger logr.Logger, c client.StatusClient) {
	logger.Info("Updating status on FunctionStatus", "resource version", r.ResourceVersion)

	r.computeFunctionStatus()
	err := c.Status().Update(ctx, r)
	if err != nil {
		logger.Error(err, "failed to update status on FunctionStatus", "function", r)
	} else {
		logger.Info("Updated status on FunctionStatus", "resource version", r.ResourceVersion)
	}
}

func (r *Function) computeFunctionStatus() {
	subComponentsReady := meta.IsStatusConditionTrue(r.Status.Conditions, string(StatefulSetReady)) &&
		meta.IsStatusConditionTrue(r.Status.Conditions, string(ServiceReady))
	if r.Spec.Pod.VPA != nil {
		subComponentsReady = subComponentsReady &&
			meta.IsStatusConditionTrue(r.Status.Conditions, string(VPAReady))
	}
	if r.Spec.MaxReplicas != nil {
		subComponentsReady = subComponentsReady &&
			meta.IsStatusConditionTrue(r.Status.Conditions, string(HPAReady))
	}
	if subComponentsReady {
		r.SetCondition(Ready, metav1.ConditionTrue, FunctionIsReady, "")
		meta.RemoveStatusCondition(&r.Status.Conditions, string(Error))
		return
	}
	r.SetCondition(Ready, metav1.ConditionFalse,
		PendingCreation, "function is not ready yet...")
}

// SetCondition adds a new condition to the Function
func (r *Function) SetCondition(condType ResourceConditionType, status metav1.ConditionStatus,
	reason ResourceConditionReason, message string) {
	setCondition(&r.Status.Conditions, condType, status, reason, message, r.Generation)
}

// RemoveCondition removes a specific condition from the Function
func (r *Function) RemoveCondition(condType ResourceConditionType) {
	meta.RemoveStatusCondition(&r.Status.Conditions, string(condType))
}

// SetComponentHash adds a new component hash to the Function
func (r *Function) SetComponentHash(component Component, specHash string) {
	if r.Status.ComponentHash == nil {
		r.Status.ComponentHash = map[Component]string{}
	}
	r.Status.ComponentHash[component] = specHash
}

// GetComponentHash returns a specific component hash or nil
func (r *Function) GetComponentHash(component Component) *string {
	if r.Status.ComponentHash != nil {
		if specHash, exist := r.Status.ComponentHash[component]; exist {
			return &specHash
		}
	}
	return nil
}

// RemoveComponentHash removes a specific component hash from the Function
func (r *Function) RemoveComponentHash(component Component) {
	if r.Status.ComponentHash != nil {
		delete(r.Status.ComponentHash, component)
	}
}

// SaveStatus will trigger Sink object update to save the current status
// conditions
func (r *Sink) SaveStatus(ctx context.Context, logger logr.Logger, c client.StatusClient) {
	logger.Info("Updating status on SinkStatus", "resource version", r.ResourceVersion)

	r.computeSinkStatus()
	err := c.Status().Update(ctx, r)
	if err != nil {
		logger.Error(err, "failed to update status on SinkStatus", "sink", r)
	} else {
		logger.Info("Updated status on SinkStatus", "resource version", r.ResourceVersion)
	}
}

func (r *Sink) computeSinkStatus() {
	subComponentsReady := meta.IsStatusConditionTrue(r.Status.Conditions, string(StatefulSetReady)) &&
		meta.IsStatusConditionTrue(r.Status.Conditions, string(ServiceReady))
	if r.Spec.Pod.VPA != nil {
		subComponentsReady = subComponentsReady &&
			meta.IsStatusConditionTrue(r.Status.Conditions, string(VPAReady))
	}
	if r.Spec.MaxReplicas != nil {
		subComponentsReady = subComponentsReady &&
			meta.IsStatusConditionTrue(r.Status.Conditions, string(HPAReady))
	}
	if subComponentsReady {
		r.SetCondition(Ready, metav1.ConditionTrue, SinkIsReady, "")
		meta.RemoveStatusCondition(&r.Status.Conditions, string(Error))
		return
	}
	r.SetCondition(Ready, metav1.ConditionFalse,
		PendingCreation, "function is not ready yet...")
}

// SetCondition adds a new condition to the Sink
func (r *Sink) SetCondition(condType ResourceConditionType, status metav1.ConditionStatus,
	reason ResourceConditionReason, message string) {
	setCondition(&r.Status.Conditions, condType, status, reason, message, r.Generation)
}

// RemoveCondition removes a specific condition from the Sink
func (r *Sink) RemoveCondition(condType ResourceConditionType) {
	meta.RemoveStatusCondition(&r.Status.Conditions, string(condType))
}

// SetComponentHash adds a new component hash to the Sink
func (r *Sink) SetComponentHash(component Component, specHash string) {
	if r.Status.ComponentHash == nil {
		r.Status.ComponentHash = map[Component]string{}
	}
	r.Status.ComponentHash[component] = specHash
}

// GetComponentHash returns a specific component hash or nil
func (r *Sink) GetComponentHash(component Component) *string {
	if r.Status.ComponentHash != nil {
		if specHash, exist := r.Status.ComponentHash[component]; exist {
			return &specHash
		}
	}
	return nil
}

// RemoveComponentHash removes a specific component hash from the Sink
func (r *Sink) RemoveComponentHash(component Component) {
	if r.Status.ComponentHash != nil {
		delete(r.Status.ComponentHash, component)
	}
}

// SaveStatus will trigger Source object update to save the current status
// conditions
func (r *Source) SaveStatus(ctx context.Context, logger logr.Logger, c client.StatusClient) {
	logger.Info("Updating status on SourceStatus", "resource version", r.ResourceVersion)

	r.computeSourceStatus()
	err := c.Status().Update(ctx, r)
	if err != nil {
		logger.Error(err, "failed to update status on SourceStatus", "source", r)
	} else {
		logger.Info("Updated status on SourceStatus", "resource version", r.ResourceVersion)
	}
}

func (r *Source) computeSourceStatus() {
	subComponentsReady := meta.IsStatusConditionTrue(r.Status.Conditions, string(StatefulSetReady)) &&
		meta.IsStatusConditionTrue(r.Status.Conditions, string(ServiceReady))
	if r.Spec.Pod.VPA != nil {
		subComponentsReady = subComponentsReady &&
			meta.IsStatusConditionTrue(r.Status.Conditions, string(VPAReady))
	}
	if r.Spec.MaxReplicas != nil {
		subComponentsReady = subComponentsReady &&
			meta.IsStatusConditionTrue(r.Status.Conditions, string(HPAReady))
	}
	if subComponentsReady {
		r.SetCondition(Ready, metav1.ConditionTrue, SourceIsReady, "")
		meta.RemoveStatusCondition(&r.Status.Conditions, string(Error))
		return
	}
	r.SetCondition(Ready, metav1.ConditionFalse,
		PendingCreation, "function is not ready yet...")
}

// SetCondition adds a new condition to the Source
func (r *Source) SetCondition(condType ResourceConditionType, status metav1.ConditionStatus,
	reason ResourceConditionReason, message string) {
	setCondition(&r.Status.Conditions, condType, status, reason, message, r.Generation)
}

// RemoveCondition removes a specific condition from the Source
func (r *Source) RemoveCondition(condType ResourceConditionType) {
	meta.RemoveStatusCondition(&r.Status.Conditions, string(condType))
}

// SetComponentHash adds a new component hash to the Source
func (r *Source) SetComponentHash(component Component, specHash string) {
	if r.Status.ComponentHash == nil {
		r.Status.ComponentHash = map[Component]string{}
	}
	r.Status.ComponentHash[component] = specHash
}

// GetComponentHash returns a specific component hash or nil
func (r *Source) GetComponentHash(component Component) *string {
	if r.Status.ComponentHash != nil {
		if specHash, exist := r.Status.ComponentHash[component]; exist {
			return &specHash
		}
	}
	return nil
}

// RemoveComponentHash removes a specific component hash from the Source
func (r *Source) RemoveComponentHash(component Component) {
	if r.Status.ComponentHash != nil {
		delete(r.Status.ComponentHash, component)
	}
}

// SaveStatus will trigger FunctionMesh object update to save the current status
// conditions
func (r *FunctionMesh) SaveStatus(ctx context.Context, logger logr.Logger, c client.StatusClient) {
	logger.Info("Updating status on FunctionMeshStatus", "resource version", r.ResourceVersion)

	err := c.Status().Update(ctx, r)
	if err != nil {
		logger.Error(err, "failed to update status on FunctionMeshStatus", "functionmesh", r)
	} else {
		logger.Info("Updated status on FunctionMeshStatus", "resource version", r.ResourceVersion)
	}
}

// SetCondition adds a new condition to the FunctionMesh
func (r *FunctionMesh) SetCondition(condType ResourceConditionType, status metav1.ConditionStatus,
	reason ResourceConditionReason, message string) {
	setCondition(&r.Status.Conditions, condType, status, reason, message, r.Generation)
}

// RemoveCondition removes a specific condition from the FunctionMesh
func (r *FunctionMesh) RemoveCondition(condType ResourceConditionType) {
	meta.RemoveStatusCondition(&r.Status.Conditions, string(condType))
}

func (r *ComponentCondition) SetStatus(status ResourceConditionType) {
	r.Status = status
}

func (r *ComponentCondition) SetHash(specHash string) {
	r.Hash = &specHash
}

// CreateCondition initializes a new status condition
func CreateCondition(generation *int64, condType ResourceConditionType, status metav1.ConditionStatus,
	reason ResourceConditionReason, message string) metav1.Condition {
	cond := &metav1.Condition{
		Type:               string(condType),
		Status:             status,
		Reason:             string(reason),
		Message:            message,
		LastTransitionTime: metav1.Time{Time: time.Now()},
	}
	if generation != nil {
		cond.ObservedGeneration = *generation
	}
	return *cond
}

func setCondition(conditions *[]metav1.Condition, condType ResourceConditionType, status metav1.ConditionStatus,
	reason ResourceConditionReason, message string, generation int64) {
	if strings.Contains(string(condType), "Ready") && status == metav1.ConditionTrue {
		meta.SetStatusCondition(conditions, CreateCondition(
			&generation, condType, status, reason, message))
	} else {
		meta.SetStatusCondition(conditions, CreateCondition(
			nil, condType, status, reason, message))
	}
}
