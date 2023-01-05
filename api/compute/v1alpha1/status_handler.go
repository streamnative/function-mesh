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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceConditionType string

const (
	// Created indicates the resource has been created
	Created ResourceConditionType = "Created"
	// Error indicates the resource had an error
	Error ResourceConditionType = "Error"
	// Pending indicates the resource hasn't been created
	Pending ResourceConditionType = "Pending"
	// Orphaned indicates that the resource is marked for deletion but hasn't
	// been deleted yet
	Orphaned ResourceConditionType = "Orphaned"
	// Ready indicates the object is fully created
	Ready ResourceConditionType = "Ready"
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
func (r *Function) SaveStatus(ctx context.Context, logger logr.Logger, c client.Client) {
	logger.Info("Updating status on FunctionStatus", "resource version", r.ResourceVersion)

	err := c.Status().Update(ctx, r)
	if err != nil {
		logger.Error(err, "failed to update status on FunctionStatus", "function", r)
	} else {
		logger.Info("Updated status on FunctionStatus", "resource version", r.ResourceVersion)
	}
}

// SetCondition adds a new condition to the Function
func (r *Function) SetCondition(component Component, condition *metav1.Condition) *Function {
	if r.Status.Conditions == nil {
		r.Status.Conditions = make(map[Component]metav1.Condition)
	}
	r.Status.Conditions[component] = *condition
	return r
}

// SaveStatus will trigger Sink object update to save the current status
// conditions
func (r *Sink) SaveStatus(ctx context.Context, logger logr.Logger, c client.Client) {
	logger.Info("Updating status on SinkStatus", "resource version", r.ResourceVersion)

	err := c.Status().Update(ctx, r)
	if err != nil {
		logger.Error(err, "failed to update status on SinkStatus", "sink", r)
	} else {
		logger.Info("Updated status on SinkStatus", "resource version", r.ResourceVersion)
	}
}

// SetCondition adds a new condition to the Sink
func (r *Sink) SetCondition(component Component, condition *metav1.Condition) *Sink {
	if r.Status.Conditions == nil {
		r.Status.Conditions = make(map[Component]metav1.Condition)
	}
	r.Status.Conditions[component] = *condition
	return r
}

// SaveStatus will trigger Source object update to save the current status
// conditions
func (r *Source) SaveStatus(ctx context.Context, logger logr.Logger, c client.Client) {
	logger.Info("Updating status on SourceStatus", "resource version", r.ResourceVersion)

	err := c.Status().Update(ctx, r)
	if err != nil {
		logger.Error(err, "failed to update status on SourceStatus", "source", r)
	} else {
		logger.Info("Updated status on SourceStatus", "resource version", r.ResourceVersion)
	}
}

// SetCondition adds a new condition to the Source
func (r *Source) SetCondition(component Component, condition *metav1.Condition) *Source {
	if r.Status.Conditions == nil {
		r.Status.Conditions = make(map[Component]metav1.Condition)
	}
	r.Status.Conditions[component] = *condition
	return r
}

// SaveStatus will trigger FunctionMesh object update to save the current status
// conditions
func (r *FunctionMesh) SaveStatus(ctx context.Context, logger logr.Logger, c client.Client) {
	logger.Info("Updating status on FunctionMeshStatus", "resource version", r.ResourceVersion)

	err := c.Status().Update(ctx, r)
	if err != nil {
		logger.Error(err, "failed to update status on FunctionMeshStatus", "functionmesh", r)
	} else {
		logger.Info("Updated status on FunctionMeshStatus", "resource version", r.ResourceVersion)
	}
}

// SetCondition adds a new condition to the FunctionMesh
func (r *FunctionMesh) SetCondition(condition *metav1.Condition) *FunctionMesh {
	r.Status.Condition = *condition
	return r
}

// SetFunctionCondition adds a new function condition to the FunctionMesh
func (r *FunctionMesh) SetFunctionCondition(name string, condition *metav1.Condition) *FunctionMesh {
	if r.Status.FunctionConditions == nil {
		r.Status.FunctionConditions = make(map[string]metav1.Condition)
	}
	r.Status.FunctionConditions[name] = *condition
	return r
}

// SetSinkCondition adds a new sink condition to the FunctionMesh
func (r *FunctionMesh) SetSinkCondition(name string, condition *metav1.Condition) *FunctionMesh {
	if r.Status.SinkConditions == nil {
		r.Status.SinkConditions = make(map[string]metav1.Condition)
	}
	r.Status.SinkConditions[name] = *condition
	return r
}

// SetSourceCondition adds a new source condition to the FunctionMesh
func (r *FunctionMesh) SetSourceCondition(name string, condition *metav1.Condition) *FunctionMesh {
	if r.Status.SourceConditions == nil {
		r.Status.SourceConditions = make(map[string]metav1.Condition)
	}
	r.Status.SourceConditions[name] = *condition
	return r
}

// CreateCondition initializes a new status condition
func CreateCondition(condType ResourceConditionType, status metav1.ConditionStatus,
	reason ResourceConditionReason, message string) *metav1.Condition {
	cond := &metav1.Condition{
		Type:               string(condType),
		Status:             status,
		Reason:             string(reason),
		Message:            message,
		LastTransitionTime: metav1.Time{Time: time.Now()},
	}
	return cond
}
