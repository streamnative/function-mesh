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
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/streamnative/function-mesh/api/v1alpha1"
	"github.com/streamnative/function-mesh/controllers/spec"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Source Controller", func() {
	var objectMeta = metav1.ObjectMeta{
		Name:      "test-source",
		Namespace: "default",
		UID:       "dead-beef",
	}
	pulsarConfig := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pulsar",
			Namespace: "default",
		},
		Data: map[string]string{
			"webServiceURL":    "http://test-pulsar-broker.default.svc.cluster.local:8080",
			"brokerServiceURL": "pulsar://test-pulsar-broker.default.svc.cluster.local:6650",
		},
	}
	var pulsar = &v1alpha1.PulsarMessaging{
		PulsarConfig: pulsarConfig.Name,
		//AuthConfig: "test-auth",
	}

	Context("Simple Source Item", func() {
		replicas := int32(1)
		maxReplicas := int32(1)
		source := &v1alpha1.Source{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Sink",
				APIVersion: "cloud.streamnative.io/v1alpha1",
			},
			ObjectMeta: objectMeta,
			Spec: v1alpha1.SourceSpec{
				Name:        objectMeta.Name,
				ClassName:   "org.apache.pulsar.io.debezium.mongodb.DebeziumMongoDbSource",
				Tenant:      "public",
				ClusterName: "test-pulsar",
				SourceType:  "org.apache.pulsar.common.schema.KeyValue",
				SinkType:    "org.apache.pulsar.common.schema.KeyValue",
				Output: v1alpha1.OutputConf{
					Topic: "persistent://public/default/destination",
					ProducerConf: &v1alpha1.ProducerConfig{
						MaxPendingMessages:                 1000,
						MaxPendingMessagesAcrossPartitions: 50000,
						UseThreadLocalProducers:            true,
					},
				},
				SourceConfig: map[string]string{
					"mongodb.hosts":      "rs0/mongo-dbz-0.mongo.default.svc.cluster.local:27017,rs0/mongo-dbz-1.mongo.default.svc.cluster.local:27017,rs0/mongo-dbz-2.mongo.default.svc.cluster.local:27017",
					"mongodb.name":       "dbserver1",
					"mongodb.user":       "debezium",
					"mongodb.password":   "dbz",
					"mongodb.task.id":    "1",
					"database.whitelist": "inventory",
					"pulsar.service.url": "pulsar://test-pulsar-broker.default.svc.cluster.local:6650",
				},
				Replicas:    &replicas,
				MaxReplicas: &maxReplicas,
				Messaging: v1alpha1.Messaging{
					Pulsar: pulsar,
				},
				Runtime: v1alpha1.Runtime{
					Java: &v1alpha1.JavaRuntime{
						Jar:         "connectors/pulsar-io-elastic-search-2.7.0-rc-pm-3.nar",
						JarLocation: "",
					},
				},
			},
		}
		if source.Status.Conditions == nil {
			source.Status.Conditions = make(map[v1alpha1.Component]v1alpha1.ResourceCondition)
		}
		statefulSet := spec.MakeSourceStatefulSet(source)

		It("Should create pulsar configmap successfully", func() {
			Expect(k8sClient.Create(context.Background(), pulsarConfig)).Should(Succeed())
		})

		It("Should create successfully", func() {
			Expect(k8sClient.Create(context.Background(), statefulSet)).Should(Succeed())
		})

		It("Should delete successfully", func() {
			Expect(k8sClient.Delete(context.Background(), statefulSet)).Should(Succeed())
		})

		It("Should delete pulsar configmap successfully", func() {
			Expect(k8sClient.Delete(context.Background(), pulsarConfig)).Should(Succeed())
		})

	})
})
