package v1alpha2

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	apispec "github.com/streamnative/function-mesh/pkg/spec"
)

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
