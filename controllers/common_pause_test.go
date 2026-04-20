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
