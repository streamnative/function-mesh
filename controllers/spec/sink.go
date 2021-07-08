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

import (
	"fmt"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/streamnative/function-mesh/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	autov1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MakeSinkHPA(sink *v1alpha1.Sink) *autov1.HorizontalPodAutoscaler {
	objectMeta := MakeSinkObjectMeta(sink)
	return MakeHPA(objectMeta, *sink.Spec.Replicas, *sink.Spec.MaxReplicas, sink.Kind)
}

func MakeSinkService(sink *v1alpha1.Sink, replica int) *corev1.Service {
	labels := MakeSinkServiceLabels(sink, replica)
	objectMeta := MakeSinkServiceObjectMeta(sink, replica)
	return MakeService(objectMeta, labels)
}

func MakeSinkStatefulSet(sink *v1alpha1.Sink) *appsv1.StatefulSet {
	objectMeta := MakeSinkObjectMeta(sink)
	return MakeStatefulSet(objectMeta, sink.Spec.Replicas, MakeSinkContainer(sink),
		makeSinkVolumes(sink), MakeSinkLabels(sink), sink.Spec.Pod)
}

func MakeSinkObjectMeta(sink *v1alpha1.Sink) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		Name:      makeJobName(sink.Name, v1alpha1.SinkComponent),
		Namespace: sink.Namespace,
		OwnerReferences: []metav1.OwnerReference{
			*metav1.NewControllerRef(sink, sink.GroupVersionKind()),
		},
	}
}

func MakeSinkServiceObjectMeta(sink *v1alpha1.Sink, replica int) *metav1.ObjectMeta {
	meta := MakeSinkObjectMeta(sink)
	meta.Name = MakeSinkServiceName(sink, replica)
	return meta
}

func MakeSinkServiceName(sink *v1alpha1.Sink, replica int) string {
	return fmt.Sprintf("%s-%d", makeJobName(sink.Name, v1alpha1.SinkComponent), replica)
}

func MakeSinkContainer(sink *v1alpha1.Sink) *corev1.Container {
	imagePullPolicy := sink.Spec.ImagePullPolicy
	if imagePullPolicy == "" {
		imagePullPolicy = corev1.PullIfNotPresent
	}
	return &corev1.Container{
		// TODO new container to pull user code image and upload jars into bookkeeper
		Name:            "pulsar-sink",
		Image:           getSinkRunnerImage(&sink.Spec),
		Command:         MakeSinkCommand(sink),
		Ports:           []corev1.ContainerPort{GRPCPort, MetricsPort},
		Env:             generateContainerEnv(sink.Spec.SecretsMap),
		Resources:       sink.Spec.Resources,
		ImagePullPolicy: imagePullPolicy,
		EnvFrom:         generateContainerEnvFrom(sink.Spec.Pulsar.PulsarConfig, sink.Spec.Pulsar.AuthSecret, sink.Spec.Pulsar.TLSSecret),
		VolumeMounts:    makeSinkVolumeMounts(sink),
	}
}

func MakeSinkLabels(sink *v1alpha1.Sink) map[string]string {
	labels := make(map[string]string)
	labels["component"] = ComponentSink
	labels["name"] = sink.Name
	labels["namespace"] = sink.Namespace

	return labels
}

func MakeSinkServiceLabels(sink *v1alpha1.Sink, replica int) map[string]string {
	podName := MakeSinkServiceName(sink, replica)
	labels := MakeSinkLabels(sink)
	labels["statefulset.kubernetes.io/pod-name"] = podName

	return labels
}

func makeSinkVolumes(sink *v1alpha1.Sink) []corev1.Volume {
	return generatePodVolumes(sink.Spec.Pod.Volumes, nil, sink.Spec.Input.SourceSpecs)
}

func makeSinkVolumeMounts(sink *v1alpha1.Sink) []corev1.VolumeMount {
	return generateContainerVolumeMounts(sink.Spec.VolumeMounts, nil, sink.Spec.Input.SourceSpecs)
}

func MakeSinkCommand(sink *v1alpha1.Sink) []string {
	spec := sink.Spec
	return MakeJavaFunctionCommand(spec.Java.JarLocation, spec.Java.Jar,
		spec.Name, spec.ClusterName, generateSinkDetailsInJSON(sink),
		spec.Resources.Requests.Memory().ToDec().String(), spec.Java.ExtraDependenciesDir,
		spec.Pulsar.AuthSecret != "", spec.Pulsar.TLSSecret != "")
}

func generateSinkDetailsInJSON(sink *v1alpha1.Sink) string {
	sourceDetails := convertSinkDetails(sink)
	marshaler := &jsonpb.Marshaler{}
	json, error := marshaler.MarshalToString(sourceDetails)
	if error != nil {
		// TODO
		panic(error)
	}
	return json
}
