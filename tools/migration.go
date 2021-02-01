package main

import (
	"fmt"
	cmdutils "github.com/streamnative/pulsarctl/pkg/cmdutils"
	"github.com/streamnative/pulsarctl/pkg/pulsar/common"
	"os"
	"strings"
	"gopkg.in/yaml.v2"
	"github.com/streamnative/function-mesh/api/v1alpha1"
)

func main() {
	admin := cmdutils.NewPulsarClient()
	functionAdmin := cmdutils.NewPulsarClientWithAPIVersion(common.V3)
	tenants, err := admin.Tenants().List()
	if err != nil {
		fmt.Printf("List tenant failed from service %s\n", cmdutils.PulsarCtlConfig.WebServiceURL)
		os.Exit(1)
	}
	clusters, err := admin.Clusters().List()
	if err != nil {
		fmt.Printf("List clusters failed from service %s\n", cmdutils.PulsarCtlConfig.WebServiceURL)
		os.Exit(1)
	}
	var pulsarCluster = ""
	for _, cluster := range clusters {
		pulsarCluster = cluster
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
				replicas := int32(functionConfig.Parallelism)
				sourceSpecs := make(map[string]v1alpha1.ConsumerConfig)
				for s := range functionConfig.InputSpecs {
					//sourceSpec[s] = functionConfig.InputSpecs[s].SchemaType
					receiveQueueSize := int32(functionConfig.InputSpecs[s].ReceiverQueueSize)
					sourceSpecs[s] = v1alpha1.ConsumerConfig{
						SchemaType:         functionConfig.InputSpecs[s].SchemaType,
						SerdeClassName:     functionConfig.InputSpecs[s].SerdeClassName,
						IsRegexPattern:     functionConfig.InputSpecs[s].IsRegexPattern,
						ReceiverQueueSize:  receiveQueueSize,
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
				functionSpec := v1alpha1.FunctionSpec{
					Name:           functionConfig.Name,
					ClassName:      functionConfig.ClassName,
					Tenant:         functionConfig.Tenant,
					ClusterName:    pulsarCluster,
					AutoAck:        &functionConfig.AutoAck,
					RetainOrdering: functionConfig.RetainOrdering,
					Replicas:       &replicas,
					Input: v1alpha1.InputConf{
						Topics:              functionConfig.Inputs,
						TopicPattern:        topicPattern,
						CustomSerdeSources:  functionConfig.CustomSerdeInputs,
						CustomSchemaSources: functionConfig.CustomSchemaInputs,
						SourceSpecs:         sourceSpecs,
					},
					MaxReplicas: &replicas,
					LogTopic: functionConfig.LogTopic,
					Timeout: timeoutMs,

					//SinkType: functionConfig.
				}
				data, err := yaml.Marshal(&functionSpec)
				if err != nil {
					fmt.Printf("Convert function %s config to yaml failed" +
						" from tenant %s namespace %s service %s err %v",
						function, tenant, namespace, cmdutils.PulsarCtlConfig.WebServiceURL, err)
					os.Exit(1)
				}
				fmt.Println(string(data))
			}
		}
	}
}