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

import "time"

// copy from https://github.com/apache/pulsar/blob/master/pulsar-function-go/conf/conf.go
// directly import Conf from pulsar-function-go/conf will break envtest

type GoFunctionConf struct {
	PulsarServiceURL string        `json:"pulsarServiceURL" yaml:"pulsarServiceURL"`
	InstanceID       int           `json:"instanceID" yaml:"instanceID"`
	FuncID           string        `json:"funcID" yaml:"funcID"`
	FuncVersion      string        `json:"funcVersion" yaml:"funcVersion"`
	MaxBufTuples     int           `json:"maxBufTuples" yaml:"maxBufTuples"`
	Port             int           `json:"port" yaml:"port"`
	ClusterName      string        `json:"clusterName" yaml:"clusterName"`
	KillAfterIdle    time.Duration `json:"killAfterIdleMs" yaml:"killAfterIdleMs"`
	// function details config
	Tenant               string `json:"tenant" yaml:"tenant"`
	NameSpace            string `json:"nameSpace" yaml:"nameSpace"`
	Name                 string `json:"name" yaml:"name"`
	LogTopic             string `json:"logTopic" yaml:"logTopic"`
	ProcessingGuarantees int32  `json:"processingGuarantees" yaml:"processingGuarantees"`
	SecretsMap           string `json:"secretsMap" yaml:"secretsMap"`
	Runtime              int32  `json:"runtime" yaml:"runtime"`
	AutoACK              bool   `json:"autoAck" yaml:"autoAck"`
	Parallelism          int32  `json:"parallelism" yaml:"parallelism"`
	//source config
	SubscriptionType    int32  `json:"subscriptionType" yaml:"subscriptionType"`
	TimeoutMs           uint64 `json:"timeoutMs" yaml:"timeoutMs"`
	SubscriptionName    string `json:"subscriptionName" yaml:"subscriptionName"`
	CleanupSubscription bool   `json:"cleanupSubscription"  yaml:"cleanupSubscription"`
	//source input specs
	SourceSpecTopic            string `json:"sourceSpecsTopic" yaml:"sourceSpecsTopic"`
	SourceSchemaType           string `json:"sourceSchemaType" yaml:"sourceSchemaType"`
	IsRegexPatternSubscription bool   `json:"isRegexPatternSubscription" yaml:"isRegexPatternSubscription"`
	ReceiverQueueSize          int32  `json:"receiverQueueSize" yaml:"receiverQueueSize"`
	//sink spec config
	SinkSpecTopic  string `json:"sinkSpecsTopic" yaml:"sinkSpecsTopic"`
	SinkSchemaType string `json:"sinkSchemaType" yaml:"sinkSchemaType"`
	//resources config
	CPU  float64 `json:"cpu" yaml:"cpu"`
	RAM  int64   `json:"ram" yaml:"ram"`
	Disk int64   `json:"disk" yaml:"disk"`
	//retryDetails config
	MaxMessageRetries           int32  `json:"maxMessageRetries" yaml:"maxMessageRetries"`
	DeadLetterTopic             string `json:"deadLetterTopic" yaml:"deadLetterTopic"`
	ExpectedHealthCheckInterval int32  `json:"expectedHealthCheckInterval" yaml:"expectedHealthCheckInterval"`
	UserConfig                  string `json:"userConfig" yaml:"userConfig"`
	//metrics config
	MetricsPort int `json:"metricsPort" yaml:"metricsPort"`
}
