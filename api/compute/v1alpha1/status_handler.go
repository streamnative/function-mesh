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
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
	// StatefulSetReady indicates the StatefulSet is fully created
	StatefulSetReady ResourceConditionType = "StatefulSetReady"
	// ServiceReady indicates the Service is fully created
	ServiceReady ResourceConditionType = "ServiceReady"
	// HPAReady indicates the HPA is fully created
	HPAReady ResourceConditionType = "HPAReady"
	// VPAReady indicates the VPA is fully created
	VPAReady ResourceConditionType = "VPAReady"
)

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
	PendingTermination       ResourceConditionReason = "PendingTermination"
	ServiceIsReady           ResourceConditionReason = "ServiceIsReady"
	MeshIsReady              ResourceConditionReason = "MeshIsReady"
	StatefulSetIsReady       ResourceConditionReason = "StatefulSetIsReady"
	HPAIsReady               ResourceConditionReason = "HPAIsReady"
	VPAIsReady               ResourceConditionReason = "VPAIsReady"
	FunctionIsReady          ResourceConditionReason = "FunctionIsReady"
	SourceIsReady            ResourceConditionReason = "SourceIsReady"
	SinkIsReady              ResourceConditionReason = "SinkIsReady"
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
	return
}

// SetCondition adds a new condition to the Function
func (r *Function) SetCondition(condType ResourceConditionType, status metav1.ConditionStatus,
	reason ResourceConditionReason, message string) {
	meta.SetStatusCondition(&r.Status.Conditions, CreateCondition(
		r.Generation, condType, status, reason, message))
}

// RemoveCondition removes a specific condition from the Function
func (r *Function) RemoveCondition(condType ResourceConditionType) {
	meta.RemoveStatusCondition(&r.Status.Conditions, string(condType))
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
	return
}

// SetCondition adds a new condition to the Sink
func (r *Sink) SetCondition(condType ResourceConditionType, status metav1.ConditionStatus,
	reason ResourceConditionReason, message string) {
	meta.SetStatusCondition(&r.Status.Conditions, CreateCondition(
		r.Generation, condType, status, reason, message))
}

// RemoveCondition removes a specific condition from the Sink
func (r *Sink) RemoveCondition(condType ResourceConditionType) {
	meta.RemoveStatusCondition(&r.Status.Conditions, string(condType))
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
	return
}

// SetCondition adds a new condition to the Source
func (r *Source) SetCondition(condType ResourceConditionType, status metav1.ConditionStatus,
	reason ResourceConditionReason, message string) {
	meta.SetStatusCondition(&r.Status.Conditions, CreateCondition(
		r.Generation, condType, status, reason, message))
}

// RemoveCondition removes a specific condition from the Source
func (r *Source) RemoveCondition(condType ResourceConditionType) {
	meta.RemoveStatusCondition(&r.Status.Conditions, string(condType))
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
	meta.SetStatusCondition(&r.Status.Conditions, CreateCondition(
		r.Generation, condType, status, reason, message))
}

// CreateCondition initializes a new status condition
func CreateCondition(generation int64, condType ResourceConditionType, status metav1.ConditionStatus,
	reason ResourceConditionReason, message string) metav1.Condition {
	cond := &metav1.Condition{
		Type:               string(condType),
		Status:             status,
		Reason:             string(reason),
		Message:            message,
		LastTransitionTime: metav1.Time{Time: time.Now()},
		ObservedGeneration: generation,
	}
	return *cond
}
