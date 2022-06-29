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

func MakeSourceHPA(source *v1alpha1.Source) *autov2beta2.HorizontalPodAutoscaler {
	objectMeta := MakeSourceObjectMeta(source)
	targetRef := autov2beta2.CrossVersionObjectReference{
		Kind:       source.Kind,
		Name:       source.Name,
		APIVersion: source.APIVersion,
	}
	if isBuiltinHPAEnabled(source.Spec.Replicas, source.Spec.MaxReplicas, source.Spec.Pod) {
		return makeBuiltinHPA(objectMeta, *source.Spec.Replicas, *source.Spec.MaxReplicas, targetRef,
			source.Spec.Pod.BuiltinAutoscaler)
	} else if !isDefaultHPAEnabled(source.Spec.Replicas, source.Spec.MaxReplicas, source.Spec.Pod) {
		return makeHPA(objectMeta, *source.Spec.Replicas, *source.Spec.MaxReplicas, source.Spec.Pod, targetRef)
	}
	return makeDefaultHPA(objectMeta, *source.Spec.Replicas, *source.Spec.MaxReplicas, targetRef)
}

func MakeSourceService(source *v1alpha1.Source) *corev1.Service {
	labels := makeSourceLabels(source)
	objectMeta := MakeSourceObjectMeta(source)
	return MakeService(objectMeta, labels)
}

func MakeSourceStatefulSet(source *v1alpha1.Source) *appsv1.StatefulSet {
	objectMeta := MakeSourceObjectMeta(source)
	return MakeStatefulSet(objectMeta, source.Spec.Replicas, MakeSourceContainer(source),
		makeSourceVolumes(source), makeSourceLabels(source), source.Spec.Pod)
}

func MakeSourceObjectMeta(source *v1alpha1.Source) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		Name:      makeJobName(source.Name, v1alpha1.SourceComponent),
		Namespace: source.Namespace,
		Labels:    makeSourceLabels(source),
		OwnerReferences: []metav1.OwnerReference{
			*metav1.NewControllerRef(source, source.GroupVersionKind()),
		},
	}
}

func MakeSourceContainer(source *v1alpha1.Source) *corev1.Container {
	imagePullPolicy := source.Spec.ImagePullPolicy
	if imagePullPolicy == "" {
		imagePullPolicy = corev1.PullIfNotPresent
	}
	return &corev1.Container{
		// TODO new container to pull user code image and upload jars into bookkeeper
		Name:            "pulsar-source",
		Image:           getSourceRunnerImage(&source.Spec),
		Command:         makeSourceCommand(source),
		Ports:           []corev1.ContainerPort{GRPCPort, MetricsPort},
		Env:             generateContainerEnv(source.Spec.SecretsMap),
		Resources:       source.Spec.Resources,
		ImagePullPolicy: imagePullPolicy,
		EnvFrom: generateContainerEnvFrom(source.Spec.Pulsar.PulsarConfig, source.Spec.Pulsar.AuthSecret,
			source.Spec.Pulsar.TLSSecret),
		VolumeMounts: makeSourceVolumeMounts(source),
	}
}

func makeSourceLabels(source *v1alpha1.Source) map[string]string {
	labels := make(map[string]string)
	labels["component"] = ComponentSource
	labels["name"] = source.Name
	labels["namespace"] = source.Namespace

	return labels
}

func makeSourceVolumes(source *v1alpha1.Source) []corev1.Volume {
	return generatePodVolumes(source.Spec.Pod.Volumes, source.Spec.Output.ProducerConf, nil, source.Spec.TlsTrustCert)
}

func makeSourceVolumeMounts(source *v1alpha1.Source) []corev1.VolumeMount {
	return generateContainerVolumeMounts(source.Spec.VolumeMounts, source.Spec.Output.ProducerConf, nil, source.Spec.TlsTrustCert)
}

func makeSourceCommand(source *v1alpha1.Source) []string {
	spec := source.Spec
	return MakeJavaFunctionCommand(spec.Java.JarLocation, spec.Java.Jar,
		spec.Name, spec.ClusterName, generateSourceDetailsInJSON(source),
		getDecimalSIMemory(spec.Resources.Requests.Memory()), spec.Java.ExtraDependenciesDir, string(source.UID),
		spec.Pulsar.AuthSecret != "", spec.Pulsar.TLSSecret != "", spec.SecretsMap, nil, spec.TlsTrustCert)
}

func generateSourceDetailsInJSON(source *v1alpha1.Source) string {
	sourceDetails := convertSourceDetails(source)
	marshaler := &jsonpb.Marshaler{}
	json, error := marshaler.MarshalToString(sourceDetails)
	if error != nil {
		// TODO
		panic(error)
	}
	return json
}
