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
	"github.com/gogo/protobuf/jsonpb"
	"github.com/streamnative/function-mesh/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	autov2beta2 "k8s.io/api/autoscaling/v2beta2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func MakeSinkHPA(sink *v1alpha1.Sink) *autov2beta2.HorizontalPodAutoscaler {
	objectMeta := MakeSinkObjectMeta(sink)
	targetRef := autov2beta2.CrossVersionObjectReference{
		Kind:       sink.Kind,
		Name:       sink.Name,
		APIVersion: sink.APIVersion,
	}
	if isBuiltinHPAEnabled(sink.Spec.MinReplicas, sink.Spec.MaxReplicas, sink.Spec.Pod) {
		return makeBuiltinHPA(objectMeta, *sink.Spec.MinReplicas, *sink.Spec.MaxReplicas, targetRef,
			sink.Spec.Pod.BuiltinAutoscaler)
	} else if !isDefaultHPAEnabled(sink.Spec.MinReplicas, sink.Spec.MaxReplicas, sink.Spec.Pod) {
		return makeHPA(objectMeta, *sink.Spec.MinReplicas, *sink.Spec.MaxReplicas, sink.Spec.Pod, targetRef)
	}
	return makeDefaultHPA(objectMeta, *sink.Spec.MinReplicas, *sink.Spec.MaxReplicas, targetRef)
}

func MakeSinkService(sink *v1alpha1.Sink) *corev1.Service {
	labels := MakeSinkLabels(sink)
	objectMeta := MakeSinkObjectMeta(sink)
	return MakeService(objectMeta, labels)
}

func MakeSinkStatefulSet(sink *v1alpha1.Sink) *appsv1.StatefulSet {
	objectMeta := MakeSinkObjectMeta(sink)
	return MakeStatefulSet(objectMeta, sink.Spec.Replicas, MakeSinkContainer(sink),
		makeSinkVolumes(sink), MakeSinkLabels(sink), sink.Spec.Pod)
}

func MakeSinkServiceName(sink *v1alpha1.Sink) string {
	objectMeta := MakeSinkObjectMeta(sink)
	return MakeHeadlessServiceName(objectMeta.Name)
}

func MakeSinkObjectMeta(sink *v1alpha1.Sink) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		Name:      makeJobName(sink.Name, v1alpha1.SinkComponent),
		Namespace: sink.Namespace,
		Labels:    MakeSinkLabels(sink),
		OwnerReferences: []metav1.OwnerReference{
			*metav1.NewControllerRef(sink, sink.GroupVersionKind()),
		},
	}
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
		Env:             generateBasicContainerEnv(sink.Spec.SecretsMap, sink.Spec.Pod.Env),
		Resources:       sink.Spec.Resources,
		ImagePullPolicy: imagePullPolicy,
		EnvFrom: generateContainerEnvFrom(sink.Spec.Pulsar.PulsarConfig, sink.Spec.Pulsar.AuthSecret,
			sink.Spec.Pulsar.TLSSecret),
		VolumeMounts: makeSinkVolumeMounts(sink),
	}
}

func MakeSinkLabels(sink *v1alpha1.Sink) map[string]string {
	jobName := makeJobName(sink.Name, v1alpha1.SinkComponent)
	labels := map[string]string{
		"app.kubernetes.io/name":            jobName,
		"app.kubernetes.io/instance":        jobName,
		"compute.functionmesh.io/component": ComponentSink,
		"compute.functionmesh.io/name":      sink.Name,
		"compute.functionmesh.io/namespace": sink.Namespace,
		// The following will be deprecated after two releases
		"component": ComponentSink,
		"name":      sink.Name,
		"namespace": sink.Namespace,
	}
	return labels
}

func makeSinkVolumes(sink *v1alpha1.Sink) []corev1.Volume {
	return generatePodVolumes(
		sink.Spec.Pod.Volumes,
		nil,
		sink.Spec.Input.SourceSpecs,
		sink.Spec.Pulsar.TLSConfig,
		getRuntimeLogConfigNames(sink.Spec.Java, sink.Spec.Python, sink.Spec.Golang))
}

func makeSinkVolumeMounts(sink *v1alpha1.Sink) []corev1.VolumeMount {
	return generateContainerVolumeMounts(
		sink.Spec.VolumeMounts,
		nil,
		sink.Spec.Input.SourceSpecs,
		sink.Spec.Pulsar.TLSConfig,
		getRuntimeLogConfigNames(sink.Spec.Java, sink.Spec.Python, sink.Spec.Golang))
}

func MakeSinkCommand(sink *v1alpha1.Sink) []string {
	spec := sink.Spec
	return MakeJavaFunctionCommand(spec.Java.JarLocation, spec.Java.Jar,
		spec.Name, spec.ClusterName,
		generateJavaLogConfigCommand(sink.Spec.Java),
		parseJavaLogLevel(sink.Spec.Java),
		generateSinkDetailsInJSON(sink),
		getDecimalSIMemory(spec.Resources.Requests.Memory()), spec.Java.ExtraDependenciesDir, string(sink.UID),
		spec.Pulsar.AuthSecret != "", spec.Pulsar.TLSSecret != "", spec.SecretsMap, nil, spec.Pulsar.TLSConfig, spec.Pulsar.AuthConfig)
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
