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

// Package v1alpha1 contains API Schema definitions for the cloud v1alpha1 API group
// +kubebuilder:object:generate=true
// +groupName=compute.functionmesh.io
package v1alpha1

type BuiltinHPARule string

const (
	AverageUtilizationCPUPercent80 BuiltinHPARule = "AverageUtilizationCPUPercent80"
	AverageUtilizationCPUPercent50 BuiltinHPARule = "AverageUtilizationCPUPercent50"
	AverageUtilizationCPUPercent20 BuiltinHPARule = "AverageUtilizationCPUPercent20"

	AverageUtilizationMemoryPercent80 BuiltinHPARule = "AverageUtilizationMemoryPercent80"
	AverageUtilizationMemoryPercent50 BuiltinHPARule = "AverageUtilizationMemoryPercent50"
	AverageUtilizationMemoryPercent20 BuiltinHPARule = "AverageUtilizationMemoryPercent20"
)
