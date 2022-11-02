package utils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/discovery"
)

var (
	// GroupVersionsVPA is a list of group versions for vertical pod autoscaler
	// It should be updated when the watched crd use a new version
	GroupVersionsVPA = []string{"autoscaling.k8s.io/v1"}
)

type WatchFlags struct {
	// the controller should not watch VPA CRDs if WatchVPACRDs is false
	WatchVPACRDs bool
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
