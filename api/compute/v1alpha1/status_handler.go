package v1alpha1

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type FunctionMeshConditionType string

const (
	// Created indicates the resource has been created
	Created FunctionMeshConditionType = "Created"
	// Terminated indicates the resource has been terminated
	Terminated FunctionMeshConditionType = "Terminated"
	// Error indicates the resource had an error
	Error FunctionMeshConditionType = "Error"
	// Pending indicates the resource hasn't been created
	Pending FunctionMeshConditionType = "Pending"
	// Orphaned indicates that the resource is marked for deletion but hasn't
	// been deleted yet
	Orphaned FunctionMeshConditionType = "Orphaned"
	// Unknown indicates the status is unavailable
	Unknown FunctionMeshConditionType = "Unknown"
	// Ready indicates the object is fully created
	Ready FunctionMeshConditionType = "Ready"
)

type FunctionMeshConditionReason string

const (
	ErrorCreatingStatefulSet FunctionMeshConditionReason = "ErrorCreatingStatefulSet"
	ErrorCreatingFunction    FunctionMeshConditionReason = "ErrorCreatingFunction"
	ErrorCreatingSource      FunctionMeshConditionReason = "ErrorCreatingSource"
	ErrorCreatingSink        FunctionMeshConditionReason = "ErrorCreatingSink"
	ErrorCreatingHPA         FunctionMeshConditionReason = "ErrorCreatingHPA"
	ErrorCreatingVPA         FunctionMeshConditionReason = "ErrorCreatingVPA"
	ErrorCreatingService     FunctionMeshConditionReason = "ErrorCreatingService"
	StatefulSetError         FunctionMeshConditionReason = "StatefulSetError"
	FunctionError            FunctionMeshConditionReason = "FunctionError"
	SourceError              FunctionMeshConditionReason = "SourceError"
	SinkError                FunctionMeshConditionReason = "SinkError"
	ServiceError             FunctionMeshConditionReason = "ServiceError"
	HPAError                 FunctionMeshConditionReason = "HPAError"
	VPAError                 FunctionMeshConditionReason = "VPAError"
	PendingCreation          FunctionMeshConditionReason = "PendingCreation"
	PendingTermination       FunctionMeshConditionReason = "PendingTermination"
	ServiceIsReady           FunctionMeshConditionReason = "ServiceIsReady"
	MeshIsReady              FunctionMeshConditionReason = "MeshIsReady"
	StatefulSetIsReady       FunctionMeshConditionReason = "StatefulSetIsReady"
	HPAIsReady               FunctionMeshConditionReason = "HPAIsReady"
	VPAIsReady               FunctionMeshConditionReason = "VPAIsReady"
	FunctionIsReady          FunctionMeshConditionReason = "FunctionIsReady"
	SourceIsReady            FunctionMeshConditionReason = "SourceIsReady"
	SinkIsReady              FunctionMeshConditionReason = "SinkIsReady"
)

// SaveStatus will trigger Function object update to save the current status
// conditions
func (r *Function) SaveStatus(ctx context.Context, logger logr.Logger, cl client.Client) {
	logger.Info("Updating status on FunctionStatus", "resource version", r.ResourceVersion)

	err := cl.Status().Update(ctx, r)
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
func (r *Sink) SaveStatus(ctx context.Context, logger logr.Logger, cl client.Client) {
	logger.Info("Updating status on SinkStatus", "resource version", r.ResourceVersion)

	err := cl.Status().Update(ctx, r)
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
func (r *Source) SaveStatus(ctx context.Context, logger logr.Logger, cl client.Client) {
	logger.Info("Updating status on SourceStatus", "resource version", r.ResourceVersion)

	err := cl.Status().Update(ctx, r)
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
func (r *FunctionMesh) SaveStatus(ctx context.Context, logger logr.Logger, cl client.Client) {
	logger.Info("Updating status on FunctionMeshStatus", "resource version", r.ResourceVersion)

	err := cl.Status().Update(ctx, r)
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
func CreateCondition(condType FunctionMeshConditionType, status metav1.ConditionStatus,
	reason FunctionMeshConditionReason, message string) *metav1.Condition {
	cond := &metav1.Condition{
		Type:               string(condType),
		Status:             status,
		Reason:             string(reason),
		Message:            message,
		LastTransitionTime: metav1.Time{Time: time.Now()},
	}
	return cond
}
