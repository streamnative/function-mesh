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

package spec

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
	ErrorCreatingFunction ResourceConditionReason = "ErrorCreatingFunction"
	ErrorCreatingSource   ResourceConditionReason = "ErrorCreatingSource"
	ErrorCreatingSink     ResourceConditionReason = "ErrorCreatingSink"
	FunctionError         ResourceConditionReason = "FunctionError"
	SourceError           ResourceConditionReason = "SourceError"
	SinkError             ResourceConditionReason = "SinkError"
	PendingCreation       ResourceConditionReason = "PendingCreation"
	ServiceIsReady        ResourceConditionReason = "ServiceIsReady"
	StatefulSetIsReady    ResourceConditionReason = "StatefulSetIsReady"
	HPAIsReady            ResourceConditionReason = "HPAIsReady"
	VPAIsReady            ResourceConditionReason = "VPAIsReady"
	FunctionIsReady       ResourceConditionReason = "FunctionIsReady"
	SourceIsReady         ResourceConditionReason = "SourceIsReady"
	SinkIsReady           ResourceConditionReason = "SinkIsReady"
	MeshIsReady           ResourceConditionReason = "MeshIsReady"
)
