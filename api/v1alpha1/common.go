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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Messaging struct {
	Pulsar *PulsarMessaging `json:"pulsar,omitempty"`
}

type PulsarMessaging struct {
	// The config map need to contain the following fields
	// webServiceURL
	// brokerServiceURL
	PulsarConfig string `json:"pulsarConfig,omitempty"`
	AuthConfig   string `json:"authConfig,omitempty"`
}

// TODO: put them into the pulsar cofig or separate it?
//type AuthenticationConfig struct {
//	clientAuthenticationPlugin     string
//	clientAuthenticationParameters string
//	tlsTrustCertsFilePath          string
//	useTls                         bool
//	tlsAllowInsecureConnection     bool
//	tlsHostnameVerificationEnable  bool
//}

type Runtime struct {
	Java   *JavaRuntime   `json:"java,omitempty"`
	Python *PythonRuntime `json:"python,omitempty"`
	Golang *GoRuntime     `json:"golang,omitempty"`
}

type JavaRuntime struct {
	Jar         string `json:"jar,omitempty"`
	JarLocation string `json:"jarLocation,omitempty"`
}

type PythonRuntime struct {
	Py string `json:"py,omitempty"`
}

type GoRuntime struct {
	Go string `json:"go,omitempty"`
}

type SecretRef struct {
	Path string `json:"path,omitempty"`
	Key  string `json:"key,omitempty"`
}

type Component string

const (
	StatefulSet Component = "StatefulSet"
	Service     Component = "Service"
	HPA         Component = "HorizontalPodAutoscaler"
)

// The `Status` of a given `Condition` and the `Action` needed to reach the `Status`
type ResourceCondition struct {
	Condition ResourceConditionType  `json:"condition,omitempty"`
	Status    metav1.ConditionStatus `json:"status,omitempty"`
	Action    ReconcileAction        `json:"action,omitempty"`
}

type ResourceConditionType string

const (
	FunctionReady ResourceConditionType = "FunctionReady"
	SourceReady   ResourceConditionType = "SourceReady"
	SinkReady     ResourceConditionType = "SinkReady"

	StatefulSetReady ResourceConditionType = "StatefulSetReady"
	ServiceReady     ResourceConditionType = "ServiceReady"
	HPAReady         ResourceConditionType = "HPAReady"
)

type ReconcileAction string

const (
	Create   ReconcileAction = "Create"
	Delete   ReconcileAction = "Delete"
	Update   ReconcileAction = "Update"
	Wait     ReconcileAction = "Wait"
	NoAction ReconcileAction = "NoAction"
)

const (
	AtleastOnce     string = "atleast_once"
	AtmostOnce      string = "atmost_once"
	EffectivelyOnce string = "effectively_once"

	DefaultTenant  string = "public"
	DefaultCluster string = "kubernetes"
)

func validResource(resources corev1.ResourceList) bool {
	// cpu & memory > 0 and storage >= 0
	return resources.Cpu().Sign() == 1 &&
		resources.Memory().Sign() == 1 &&
		resources.Storage().Sign() >= 0
}
