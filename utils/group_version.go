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

// Package utils define some common used functions&structs
package utils

import (
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	autoscalingv2beta2 "k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
)

const (
	GroupVersionV2      = "v2"
	GroupVersionV2Beta2 = "v2beta2"
)

var (
	// GroupVersionsVPA is a list of group versions for vertical pod autoscaler
	// It should be updated when the watched crd use a new version
	GroupVersionsVPA = []string{"autoscaling.k8s.io/v1"}

	GroupVersionAutoscalingV2 = []string{autoscalingv2.SchemeGroupVersion.String()}

	GroupVersionAutoscalingV2Beta2 = []string{autoscalingv2beta2.SchemeGroupVersion.String()}
)

type GroupVersionFlags struct {
	// the controller should not watch VPA CRDs if WatchVPACRDs is false
	WatchVPACRDs bool

	// the controller should provide the HPA group version via APIAutoscalingGroupVersion
	APIAutoscalingGroupVersion string
}

// GroupVersions is a set of Kubernetes API group versions.
type GroupVersions map[string]bool

// Has returns true if the version string is in the set.
//
//	vs.Has("apps/v1")
func (v GroupVersions) Has(apiVersion string) bool {
	if _, ok := v[apiVersion]; ok {
		return true
	}

	return false
}

// HasGroupVersions returns true if the versions are both in the set
func (v GroupVersions) HasGroupVersions(versions []string) bool {
	for _, i := range versions {
		if !v.Has(i) {
			return false
		}
	}
	return true
}

// GetGroupVersions will get all group versions in the cluster
func GetGroupVersions(client discovery.ServerGroupsInterface) (GroupVersions, error) {
	var groupVersions GroupVersions

	groupList, err := client.ServerGroups()
	if err != nil {
		return groupVersions, err
	}
	groupVersions = extractGroupVersions(groupList)

	return groupVersions, nil
}

// extractGroupVersions extract the group version into a map
func extractGroupVersions(l *metav1.APIGroupList) GroupVersions {
	groupVersions := make(GroupVersions)
	for _, g := range l.Groups {
		for _, gv := range g.Versions {
			groupVersions[gv.GroupVersion] = true
		}
	}
	return groupVersions
}
