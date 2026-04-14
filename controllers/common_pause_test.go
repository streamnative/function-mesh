package controllers

import (
	"testing"
	"time"

	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	"github.com/streamnative/function-mesh/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestShouldPauseNonGenerationRollout(t *testing.T) {
	oldPauseRollout := utils.PauseRollout
	t.Cleanup(func() {
		utils.PauseRollout = oldPauseRollout
	})

	utils.PauseRollout = true
	object := &metav1.ObjectMeta{}
	if !shouldPauseNonGenerationRollout(object, false) {
		t.Fatal("expected pause rollout for non-generation reconcile")
	}

	if shouldPauseNonGenerationRollout(object, true) {
		t.Fatal("did not expect pause rollout for new generation")
	}

	now := metav1.NewTime(time.Now())
	object.DeletionTimestamp = &now
	if shouldPauseNonGenerationRollout(object, false) {
		t.Fatal("did not expect pause rollout while object is deleting")
	}
}

func TestShouldApplyPausedResourceAction(t *testing.T) {
	if !shouldApplyPausedResourceAction(v1alpha1.Create) {
		t.Fatal("expected create action to be allowed while paused")
	}
	if !shouldApplyPausedResourceAction(v1alpha1.Delete) {
		t.Fatal("expected delete action to be allowed while paused")
	}

	blockedActions := []v1alpha1.ReconcileAction{
		v1alpha1.Update,
		v1alpha1.Wait,
		v1alpha1.NoAction,
	}
	for _, action := range blockedActions {
		if shouldApplyPausedResourceAction(action) {
			t.Fatalf("did not expect %s action to be allowed while paused", action)
		}
	}
}
