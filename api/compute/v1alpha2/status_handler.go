package v1alpha2

import (
	"context"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apispec "github.com/streamnative/function-mesh/pkg/spec"
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
	subComponentsReady := meta.IsStatusConditionTrue(r.Status.Conditions, string(apispec.StatefulSetReady)) &&
		meta.IsStatusConditionTrue(r.Status.Conditions, string(apispec.ServiceReady))
	if r.Spec.Pod.VPA != nil {
		subComponentsReady = subComponentsReady &&
			meta.IsStatusConditionTrue(r.Status.Conditions, string(apispec.VPAReady))
	}
	if r.Spec.MaxReplicas != nil {
		subComponentsReady = subComponentsReady &&
			meta.IsStatusConditionTrue(r.Status.Conditions, string(apispec.HPAReady))
	}
	if subComponentsReady {
		r.SetCondition(apispec.Ready, metav1.ConditionTrue, apispec.FunctionIsReady, "")
		meta.RemoveStatusCondition(&r.Status.Conditions, string(apispec.Error))
		return
	}
	r.SetCondition(apispec.Ready, metav1.ConditionFalse,
		apispec.PendingCreation, "function is not ready yet...")
}

// SetCondition adds a new condition to the Function
func (r *Function) SetCondition(condType apispec.ResourceConditionType, status metav1.ConditionStatus,
	reason apispec.ResourceConditionReason, message string) {
	setCondition(&r.Status.Conditions, condType, status, reason, message, r.Generation)
}

// RemoveCondition removes a specific condition from the Function
func (r *Function) RemoveCondition(condType apispec.ResourceConditionType) {
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
	subComponentsReady := meta.IsStatusConditionTrue(r.Status.Conditions, string(apispec.StatefulSetReady)) &&
		meta.IsStatusConditionTrue(r.Status.Conditions, string(apispec.ServiceReady))
	if r.Spec.Pod.VPA != nil {
		subComponentsReady = subComponentsReady &&
			meta.IsStatusConditionTrue(r.Status.Conditions, string(apispec.VPAReady))
	}
	if r.Spec.MaxReplicas != nil {
		subComponentsReady = subComponentsReady &&
			meta.IsStatusConditionTrue(r.Status.Conditions, string(apispec.HPAReady))
	}
	if subComponentsReady {
		r.SetCondition(apispec.Ready, metav1.ConditionTrue, apispec.SourceIsReady, "")
		meta.RemoveStatusCondition(&r.Status.Conditions, string(apispec.Error))
		return
	}
	r.SetCondition(apispec.Ready, metav1.ConditionFalse,
		apispec.PendingCreation, "function is not ready yet...")
}

// SetCondition adds a new condition to the Source
func (r *Source) SetCondition(condType apispec.ResourceConditionType, status metav1.ConditionStatus,
	reason apispec.ResourceConditionReason, message string) {
	setCondition(&r.Status.Conditions, condType, status, reason, message, r.Generation)
}

// RemoveCondition removes a specific condition from the Source
func (r *Source) RemoveCondition(condType apispec.ResourceConditionType) {
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
	subComponentsReady := meta.IsStatusConditionTrue(r.Status.Conditions, string(apispec.StatefulSetReady)) &&
		meta.IsStatusConditionTrue(r.Status.Conditions, string(apispec.ServiceReady))
	if r.Spec.Pod.VPA != nil {
		subComponentsReady = subComponentsReady &&
			meta.IsStatusConditionTrue(r.Status.Conditions, string(apispec.VPAReady))
	}
	if r.Spec.MaxReplicas != nil {
		subComponentsReady = subComponentsReady &&
			meta.IsStatusConditionTrue(r.Status.Conditions, string(apispec.HPAReady))
	}
	if subComponentsReady {
		r.SetCondition(apispec.Ready, metav1.ConditionTrue, apispec.SinkIsReady, "")
		meta.RemoveStatusCondition(&r.Status.Conditions, string(apispec.Error))
		return
	}
	r.SetCondition(apispec.Ready, metav1.ConditionFalse,
		apispec.PendingCreation, "function is not ready yet...")
}

// SetCondition adds a new condition to the Sink
func (r *Sink) SetCondition(condType apispec.ResourceConditionType, status metav1.ConditionStatus,
	reason apispec.ResourceConditionReason, message string) {
	setCondition(&r.Status.Conditions, condType, status, reason, message, r.Generation)
}

// RemoveCondition removes a specific condition from the Sink
func (r *Sink) RemoveCondition(condType apispec.ResourceConditionType) {
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
func (r *FunctionMesh) SetCondition(condType apispec.ResourceConditionType, status metav1.ConditionStatus,
	reason apispec.ResourceConditionReason, message string) {
	setCondition(&r.Status.Conditions, condType, status, reason, message, r.Generation)
}

// RemoveCondition removes a specific condition from the FunctionMesh
func (r *FunctionMesh) RemoveCondition(condType apispec.ResourceConditionType) {
	meta.RemoveStatusCondition(&r.Status.Conditions, string(condType))
}

func (r *MeshComponentStatus) SetStatus(status apispec.ResourceConditionType) {
	r.Status = status
}

// CreateCondition initializes a new status condition
func CreateCondition(generation *int64, condType apispec.ResourceConditionType, status metav1.ConditionStatus,
	reason apispec.ResourceConditionReason, message string) metav1.Condition {
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

func setCondition(conditions *[]metav1.Condition, condType apispec.ResourceConditionType, status metav1.ConditionStatus,
	reason apispec.ResourceConditionReason, message string, generation int64) {
	if strings.Contains(string(condType), "Ready") && status == metav1.ConditionTrue {
		meta.SetStatusCondition(conditions, CreateCondition(
			&generation, condType, status, reason, message))
	} else {
		meta.SetStatusCondition(conditions, CreateCondition(
			nil, condType, status, reason, message))
	}
}
