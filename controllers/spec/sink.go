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
	"context"
	"regexp"

	"github.com/streamnative/function-mesh/utils"
	"google.golang.org/protobuf/encoding/protojson"
	appsv1 "k8s.io/api/apps/v1"
	autov2 "k8s.io/api/autoscaling/v2"
	v1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
)

func MakeSinkHPA(sink *v1alpha1.Sink) *autov2.HorizontalPodAutoscaler {
	objectMeta := MakeSinkObjectMeta(sink)
	targetRef := autov2.CrossVersionObjectReference{
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
	labels := makeSinkLabels(sink)
	objectMeta := MakeSinkObjectMeta(sink)
	return MakeService(objectMeta, labels)
}

func MakeSinkStatefulSet(ctx context.Context, r client.Reader, sink *v1alpha1.Sink) *appsv1.StatefulSet {
	objectMeta := MakeSinkObjectMeta(sink)
	return MakeStatefulSet(ctx, r, objectMeta, sink.Spec.Replicas, sink.Spec.DownloaderImage, makeSinkContainer(sink),
		makeFilebeatContainer(sink.Spec.VolumeMounts, sink.Spec.Pod.Env, sink.Spec.Name, sink.Spec.LogTopic, sink.Spec.LogTopicAgent,
			sink.Spec.Pulsar.TLSConfig, sink.Spec.Pulsar.AuthConfig, sink.Spec.Pulsar.PulsarConfig, sink.Spec.Pulsar.TLSSecret,
			sink.Spec.Pulsar.AuthSecret, sink.Spec.FilebeatImage),
		makeSinkVolumes(sink, sink.Spec.Pulsar.AuthConfig), makeSinkLabels(sink), sink.Spec.Pod, *sink.Spec.Pulsar,
		sink.Spec.Java, sink.Spec.Python, sink.Spec.Golang, sink.Spec.VolumeMounts, nil, nil)
}

func MakeSinkServiceName(sink *v1alpha1.Sink) string {
	objectMeta := MakeSinkObjectMeta(sink)
	return MakeHeadlessServiceName(objectMeta.Name)
}

func MakeSinkObjectMeta(sink *v1alpha1.Sink) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		Name:      makeJobName(sink.Name, v1alpha1.SinkComponent),
		Namespace: sink.Namespace,
		Labels:    makeSinkLabels(sink),
		OwnerReferences: []metav1.OwnerReference{
			*metav1.NewControllerRef(sink, sink.GroupVersionKind()),
		},
	}
}

func makeSinkContainer(sink *v1alpha1.Sink) *corev1.Container {
	imagePullPolicy := sink.Spec.ImagePullPolicy
	if imagePullPolicy == "" {
		imagePullPolicy = corev1.PullIfNotPresent
	}
	probe := MakeLivenessProbe(sink.Spec.Pod.Liveness)
	allowPrivilegeEscalation := false
	mounts := makeSinkVolumeMounts(sink, sink.Spec.Pulsar.AuthConfig)
	if utils.EnableInitContainers {
		mounts = append(mounts, generateDownloaderVolumeMountsForRuntime(sink.Spec.Java, nil, nil)...)
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
		VolumeMounts:  mounts,
		LivenessProbe: probe,
		SecurityContext: &corev1.SecurityContext{
			Capabilities: &corev1.Capabilities{
				Drop: []corev1.Capability{"ALL"},
			},
			AllowPrivilegeEscalation: &allowPrivilegeEscalation,
		},
	}
}

func makeSinkLabels(sink *v1alpha1.Sink) map[string]string {
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

func MakeSinkCleanUpJob(sink *v1alpha1.Sink) *v1.Job {
	labels := makeSinkLabels(sink)
	labels["owner"] = string(sink.GetUID())
	objectMeta := &metav1.ObjectMeta{
		Name:      makeJobName(sink.Name, v1alpha1.SinkComponent) + "-cleanup",
		Namespace: sink.Namespace,
		Labels:    labels,
	}
	container := makeSinkContainer(sink)
	container.Name = CleanupContainerName
	container.LivenessProbe = nil
	authConfig := sink.Spec.Pulsar.CleanupAuthConfig
	if authConfig == nil {
		authConfig = sink.Spec.Pulsar.AuthConfig
	}
	volumeMounts := makeSinkVolumeMounts(sink, authConfig)
	container.VolumeMounts = volumeMounts
	if sink.Spec.CleanupImage != "" {
		container.Image = sink.Spec.CleanupImage
	}
	topicPattern := sink.Spec.Input.TopicPattern
	inputSpecs := generateInputSpec(sink.Spec.Input)
	inputTopics := make([]string, 0, len(inputSpecs))
	for topic, spec := range inputSpecs {
		if spec.IsRegexPattern {
			topicPattern = topic
		} else {
			inputTopics = append(inputTopics, topic)
		}
	}
	imageHasPulsarctl := sink.Spec.ImageHasPulsarctl
	if match, _ := regexp.MatchString(RunnerImageHasPulsarctl, sink.Spec.Image); match {
		imageHasPulsarctl = true
	}
	command := getCleanUpCommand(imageHasPulsarctl,
		sink.Spec.Pulsar.AuthSecret != "",
		sink.Spec.Pulsar.TLSSecret != "",
		sink.Spec.Pulsar.TLSConfig,
		authConfig,
		inputTopics,
		topicPattern,
		sink.Spec.SubscriptionName,
		sink.Spec.Tenant,
		sink.Spec.Namespace,
		sink.Spec.Name, false)
	container.Command = command
	return makeCleanUpJob(objectMeta, container, makeSinkVolumes(sink, authConfig), makeSinkLabels(sink), sink.Spec.Pod)
}

func makeSinkVolumes(sink *v1alpha1.Sink, authConfig *v1alpha1.AuthConfig) []corev1.Volume {
	return generatePodVolumes(
		sink.Spec.Pod.Volumes,
		nil,
		sink.Spec.Input.SourceSpecs,
		sink.Spec.Pulsar.TLSConfig,
		authConfig,
		getRuntimeLogConfigNames(sink.Spec.Java, sink.Spec.Python, sink.Spec.Golang),
		sink.Spec.LogTopicAgent)
}

func makeSinkVolumeMounts(sink *v1alpha1.Sink, authConfig *v1alpha1.AuthConfig) []corev1.VolumeMount {
	return generateContainerVolumeMounts(
		sink.Spec.VolumeMounts,
		nil,
		sink.Spec.Input.SourceSpecs,
		sink.Spec.Pulsar.TLSConfig,
		authConfig,
		getRuntimeLogConfigNames(sink.Spec.Java, sink.Spec.Python, sink.Spec.Golang),
		sink.Spec.LogTopicAgent)
}

func MakeSinkCommand(sink *v1alpha1.Sink) []string {
	spec := sink.Spec
	hasPulsarctl := sink.Spec.ImageHasPulsarctl
	hasWget := sink.Spec.ImageHasWget
	if match, _ := regexp.MatchString(RunnerImageHasPulsarctl, sink.Spec.Image); match {
		hasPulsarctl = true
		hasWget = true
	}
	return MakeJavaFunctionCommand(spec.Java.JarLocation, spec.Java.Jar,
		spec.Name, spec.ClusterName,
		generateJavaLogConfigCommand(spec.Java, spec.LogTopicAgent),
		parseJavaLogLevel(spec.Java),
		generateSinkDetailsInJSON(sink),
		spec.Java.ExtraDependenciesDir, string(sink.UID),
		spec.Java.JavaOpts, hasPulsarctl, hasWget, spec.Pulsar.AuthSecret != "", spec.Pulsar.TLSSecret != "",
		spec.SecretsMap, spec.StateConfig, spec.Pulsar.TLSConfig, spec.Pulsar.AuthConfig, nil,
		generateJavaLogConfigFileName(spec.Java))
}

func generateSinkDetailsInJSON(sink *v1alpha1.Sink) string {
	sinkDetails := convertSinkDetails(sink)
	json, err := protojson.Marshal(sinkDetails)
	if err != nil {
		panic(err)
	}
	if err != nil {
		// TODO
		panic(err)
	}
	log.Info(string(json))
	return string(json)
}
