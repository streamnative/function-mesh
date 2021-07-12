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

package main

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/streamnative/function-mesh/api/v1alpha1"
	cmdutils "github.com/streamnative/pulsarctl/pkg/cmdutils"
	"github.com/streamnative/pulsarctl/pkg/pulsar/common"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func SHA1(s string) string {
	o := sha1.New()
	o.Write([]byte(s))
	return hex.EncodeToString(o.Sum(nil))
}

func main() {
	admin := cmdutils.NewPulsarClient()
	functionAdmin := cmdutils.NewPulsarClientWithAPIVersion(common.V3)
	tenants, err := admin.Tenants().List()
	if err != nil {
		fmt.Printf("List tenant failed from service %s\n", cmdutils.PulsarCtlConfig.WebServiceURL)
		os.Exit(1)
	}
	for _, tenant := range tenants {
		namespaces, err := admin.Namespaces().GetNamespaces(tenant)
		if err != nil {
			fmt.Printf("List namespace failed from tenant %s service %s, error %v",
				tenant, cmdutils.PulsarCtlConfig.WebServiceURL, err)
			os.Exit(1)
		}
		for _, namespace := range namespaces {
			tenantNamespace := strings.Split(namespace, "/")
			functions, err := functionAdmin.Functions().GetFunctions(tenantNamespace[0], tenantNamespace[1])
			if err != nil {
				fmt.Printf("List functions failed from tenant %s namespace %s service %s, err %v",
					tenant, namespace, cmdutils.PulsarCtlConfig.WebServiceURL, err)
				os.Exit(1)
			}
			for _, function := range functions {
				functionConfig, err := functionAdmin.Functions().GetFunction(
					tenantNamespace[0], tenantNamespace[1], function)
				if err != nil {
					fmt.Printf("Get function %s config failed from tenant %s namespace %s service %s err %v",
						function, tenant, namespace, cmdutils.PulsarCtlConfig.WebServiceURL, err)
					os.Exit(1)
				}
				functionStatus, err := functionAdmin.Functions().GetFunctionStatus(
					tenantNamespace[0], tenantNamespace[1], function)
				if err != nil {
					fmt.Printf("Get function %s status failed from tenant %s namespace %s service %s err %v",
						function, tenant, namespace, cmdutils.PulsarCtlConfig.WebServiceURL, err)
					os.Exit(1)
				}
				if len(functionStatus.Instances) > 0 {
					workerID := functionStatus.Instances[0].Status.WorkerID
					workerIDList := strings.Split(workerID, "-")
					pulsarCluster := workerIDList[1]
					replicas := int32(functionConfig.Parallelism)
					sourceSpecs := make(map[string]v1alpha1.ConsumerConfig)
					for s := range functionConfig.InputSpecs {
						receiveQueueSize := int32(functionConfig.InputSpecs[s].ReceiverQueueSize)
						sourceSpecs[s] = v1alpha1.ConsumerConfig{
							SchemaType:        functionConfig.InputSpecs[s].SchemaType,
							SerdeClassName:    functionConfig.InputSpecs[s].SerdeClassName,
							IsRegexPattern:    functionConfig.InputSpecs[s].IsRegexPattern,
							ReceiverQueueSize: &receiveQueueSize,
						}
					}
					timeoutMs := int32(0)
					if functionConfig.TimeoutMs != nil {
						timeoutMs = int32(*functionConfig.TimeoutMs)
					}
					topicPattern := ""
					if functionConfig.TopicsPattern != nil {
						topicPattern = *functionConfig.TopicsPattern
					}
					topics := functionConfig.Inputs
					if topics == nil {
						topics = make([]string, 0, len(functionConfig.InputSpecs))
						for k := range functionConfig.InputSpecs {
							topics = append(topics, k)
						}
					}
					funcConfig := make(map[string]string)
					for key, value := range functionConfig.UserConfig {
						strKey := fmt.Sprintf("%v", key)
						strValue := fmt.Sprintf("%v", value)
						funcConfig[strKey] = strValue
					}
					maxMessageRetry := int32(0)
					if functionConfig.MaxMessageRetries != nil {
						maxMessageRetry = int32(*functionConfig.MaxMessageRetries)
					}
					labels := make(map[string]string)

					labels["pulsar-cluster"] = pulsarCluster
					labels["pulsar-component"] = functionConfig.Name
					labels["pulsar-namespace"] = tenantNamespace[1]
					labels["pulsar-tenant"] = tenant
					sha1Value := SHA1(pulsarCluster + "-" + tenantNamespace[0] +
						"-" + tenantNamespace[1] + "-" + functionConfig.Name)
					functionSpec := v1alpha1.FunctionSpec{
						Name:                "function-" + sha1Value[0: 8],
						ClassName:           functionConfig.ClassName,
						Tenant:              functionConfig.Tenant,
						ClusterName:         pulsarCluster,
						AutoAck:             &functionConfig.AutoAck,
						CleanupSubscription: functionConfig.CleanupSubscription,
						RetainOrdering:      functionConfig.RetainOrdering,
						// Need to be added in pulsarctl
						// https://github.com/streamnative/pulsarctl/blob/master/pkg/pulsar/utils/function_confg.go
						// RetainKeyOrdering:   functionConfig.RetainKeyOrdering,
						Replicas: &replicas,
						Input: v1alpha1.InputConf{
							Topics:              topics,
							TopicPattern:        topicPattern,
							CustomSerdeSources:  functionConfig.CustomSerdeInputs,
							CustomSchemaSources: functionConfig.CustomSchemaInputs,
							SourceSpecs:         sourceSpecs,
						},
						// Disable maxReplicas
						//MaxReplicas: &replicas,
						Timeout:     timeoutMs,
						Output: v1alpha1.OutputConf{
							Topic:              functionConfig.Output,
							SinkSerdeClassName: functionConfig.OutputSerdeClassName,
							SinkSchemaType:     functionConfig.OutputSchemaType,
							// Need to be added in pulsarctl
							// https://github.com/streamnative/pulsarctl/blob/master/pkg/pulsar/utils/function_confg.go
							// CustomSchemaSinks:  functionConfig.CustomSchemaOutputs,
						},
						DeadLetterTopic: functionConfig.DeadLetterTopic,
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%f", functionConfig.Resources.CPU)),
								corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dM", functionConfig.Resources.RAM/1024/1024)),
							},
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse(fmt.Sprintf("%f", functionConfig.Resources.CPU)),
								corev1.ResourceMemory: resource.MustParse(fmt.Sprintf("%dM", functionConfig.Resources.RAM/1024/1024)),
							},
						},
						FuncConfig:      funcConfig,
						MaxMessageRetry: maxMessageRetry,
						Pod: v1alpha1.PodPolicy{
							Labels: labels,
						},
					}
					if functionConfig.ProcessingGuarantees != "" {
						switch strings.ToLower(functionConfig.ProcessingGuarantees) {
						case "atleast_once":
							functionSpec.ProcessingGuarantee = v1alpha1.AtleastOnce
						case "atmost_once":
							functionSpec.ProcessingGuarantee = v1alpha1.AtmostOnce
						case "effectively_once":
							functionSpec.ProcessingGuarantee = v1alpha1.EffectivelyOnce
						}
					}
					if functionConfig.SubName != "" {
						functionSpec.SubscriptionName = functionConfig.SubName
					}
					if functionConfig.RuntimeFlags != "" {
						functionSpec.RuntimeFlags = functionConfig.RuntimeFlags
					}
					if functionConfig.Jar != nil && *functionConfig.Jar != "" {
						functionSpec.Java.Jar = *functionConfig.Jar
					}
					if functionConfig.Py != nil && *functionConfig.Py != "" {
						functionSpec.Python.Py = *functionConfig.Py
					}
					if functionConfig.Go != nil && *functionConfig.Go != "" {
						functionSpec.Golang.Go = *functionConfig.Go
					}
					functionSpec.Pulsar = &v1alpha1.PulsarMessaging{
						PulsarConfig: pulsarCluster + "-function-mesh-config",
						AuthSecret:   "",
						TLSSecret:    "",
					}
					typeMeta := metav1.TypeMeta{
						APIVersion: "compute.functionmesh.io/v1alpha1",
						Kind:       "Function",
					}
					objectMeta := metav1.ObjectMeta{
						Name:      "function-" + sha1Value[0: 8],
						Labels:    labels,
					}
					functionData := v1alpha1.Function{
						TypeMeta:   typeMeta,
						ObjectMeta: objectMeta,
						Spec:       functionSpec,
					}
					if functionConfig.LogTopic != "" {
						functionSpec.LogTopic = functionConfig.LogTopic
					}
					data, err := json.Marshal(&functionData)
					if err != nil {
						fmt.Printf("Convert function %s config to json failed"+
							" from tenant %s namespace %s service %s err %v",
							function, tenant, namespace, cmdutils.PulsarCtlConfig.WebServiceURL, err)
						os.Exit(1)
					}
					y, err := yaml.JSONToYAML(data)
					if err != nil {
						fmt.Printf("Convert function %s config to yaml failed"+
							" from tenant %s namespace %s service %s err %v",
							function, tenant, namespace, cmdutils.PulsarCtlConfig.WebServiceURL, err)
						os.Exit(1)
					}
					path := "functions/" + namespace
					err = os.MkdirAll(path, os.ModePerm)
					if err != nil {
						fmt.Printf("Create directory failed for function %s from tenant %s namespace %s service %s err %v",
							function, tenant, namespace, cmdutils.PulsarCtlConfig.WebServiceURL, err)
						os.Exit(1)
					}
					filePath := path + "/" + function + ".yaml"
					f, err := os.Create(filePath)
					if err != nil {
						fmt.Printf("Create yaml file failed for function %s from tenant %s namespace %s service %s err %v",
							function, tenant, namespace, cmdutils.PulsarCtlConfig.WebServiceURL, err)
						os.Exit(1)
					}
					_, err = f.WriteString(string(y))
					if err != nil {
						fmt.Printf("Write yaml file failed for function %s from tenant %s namespace %s service %s err %v",
							function, tenant, namespace, cmdutils.PulsarCtlConfig.WebServiceURL, err)
						os.Exit(1)
					}
					f.Sync()
					f.Close()
				}
			}
		}
	}
}
