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
	"fmt"
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

func MakeSourceHPA(source *v1alpha1.Source) *autov2.HorizontalPodAutoscaler {
	objectMeta := MakeSourceObjectMeta(source)
	targetRef := autov2.CrossVersionObjectReference{
		Kind:       source.Kind,
		Name:       source.Name,
		APIVersion: source.APIVersion,
	}
	if isBuiltinHPAEnabled(source.Spec.MinReplicas, source.Spec.MaxReplicas, source.Spec.Pod) {
		return makeBuiltinHPA(objectMeta, *source.Spec.MinReplicas, *source.Spec.MaxReplicas, targetRef,
			source.Spec.Pod.BuiltinAutoscaler)
	} else if !isDefaultHPAEnabled(source.Spec.MinReplicas, source.Spec.MaxReplicas, source.Spec.Pod) {
		return makeHPA(objectMeta, *source.Spec.MinReplicas, *source.Spec.MaxReplicas, source.Spec.Pod, targetRef)
	}
	return makeDefaultHPA(objectMeta, *source.Spec.MinReplicas, *source.Spec.MaxReplicas, targetRef)
}

func MakeSourceService(source *v1alpha1.Source) *corev1.Service {
	labels := makeSourceLabels(source)
	objectMeta := MakeSourceObjectMeta(source)
	return MakeService(objectMeta, labels)
}

func MakeSourceStatefulSet(ctx context.Context, cli client.Client, source *v1alpha1.Source) (*appsv1.StatefulSet, error) {
	objectMeta := MakeSourceObjectMeta(source)
	statefulSet := MakeStatefulSet(objectMeta, source.Spec.Replicas, source.Spec.DownloaderImage, makeSourceContainer(source),
		makeFilebeatContainer(source.Spec.VolumeMounts, source.Spec.Pod.Env, source.Spec.Name, source.Spec.LogTopic, source.Spec.LogTopicAgent,
			source.Spec.Pulsar.TLSConfig, source.Spec.Pulsar.AuthConfig, source.Spec.Pulsar.PulsarConfig, source.Spec.Pulsar.TLSSecret,
			source.Spec.Pulsar.AuthSecret, source.Spec.FilebeatImage),
		makeSourceVolumes(source, source.Spec.Pulsar.AuthConfig), makeSourceLabels(source), source.Spec.Pod, *source.Spec.Pulsar,
		source.Spec.Java, source.Spec.Python, source.Spec.Golang, source.Spec.VolumeMounts, nil, nil)

	globalMeshConfigVersion, namespacedMeshConfigVersion, err := PatchStatefulSet(ctx, cli, source.Namespace, statefulSet)
	if err != nil {
		return nil, err
	}
	if globalMeshConfigVersion != "" {
		source.Status.GlobalMeshConfigRevision = globalMeshConfigVersion
	}
	if namespacedMeshConfigVersion != "" {
		source.Status.NamespacedMeshConfigRevision = namespacedMeshConfigVersion
	}

	return statefulSet, nil
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

func makeSourceContainer(source *v1alpha1.Source) *corev1.Container {
	imagePullPolicy := source.Spec.ImagePullPolicy
	if imagePullPolicy == "" {
		imagePullPolicy = corev1.PullIfNotPresent
	}
	probe := MakeLivenessProbe(source.Spec.Pod.Liveness)
	allowPrivilegeEscalation := false
	mounts := makeSourceVolumeMounts(source, source.Spec.Pulsar.AuthConfig)
	if utils.EnableInitContainers {
		mounts = append(mounts, generateDownloaderVolumeMountsForRuntime(source.Spec.Java, nil, nil)...)
	}
	return &corev1.Container{
		// TODO new container to pull user code image and upload jars into bookkeeper
		Name:            "pulsar-source",
		Image:           getSourceRunnerImage(&source.Spec),
		Command:         makeSourceCommand(source),
		Ports:           []corev1.ContainerPort{GRPCPort, MetricsPort},
		Env:             generateBasicContainerEnv(source.Spec.SecretsMap, source.Spec.Pod.Env),
		Resources:       source.Spec.Resources,
		ImagePullPolicy: imagePullPolicy,
		EnvFrom: generateContainerEnvFrom(source.Spec.Pulsar.PulsarConfig, source.Spec.Pulsar.AuthSecret,
			source.Spec.Pulsar.TLSSecret),
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

func makeSourceLabels(source *v1alpha1.Source) map[string]string {
	jobName := makeJobName(source.Name, v1alpha1.SourceComponent)
	labels := map[string]string{
		"app.kubernetes.io/name":            jobName,
		"app.kubernetes.io/instance":        jobName,
		"compute.functionmesh.io/component": ComponentSource,
		"compute.functionmesh.io/name":      source.Name,
		"compute.functionmesh.io/namespace": source.Namespace,
		// The following will be deprecated after two releases
		"component": ComponentSource,
		"name":      source.Name,
		"namespace": source.Namespace,
	}
	return labels
}

func makeSourceVolumes(source *v1alpha1.Source, authConfig *v1alpha1.AuthConfig) []corev1.Volume {
	return generatePodVolumes(
		source.Spec.Pod.Volumes,
		source.Spec.Output.ProducerConf,
		nil,
		source.Spec.Pulsar.TLSConfig,
		authConfig,
		getRuntimeLogConfigNames(source.Spec.Java, source.Spec.Python, source.Spec.Golang),
		source.Spec.LogTopicAgent)
}

func makeSourceVolumeMounts(source *v1alpha1.Source, authConfig *v1alpha1.AuthConfig) []corev1.VolumeMount {
	return generateContainerVolumeMounts(
		source.Spec.VolumeMounts,
		source.Spec.Output.ProducerConf,
		nil,
		source.Spec.Pulsar.TLSConfig,
		authConfig,
		getRuntimeLogConfigNames(source.Spec.Java, source.Spec.Python, source.Spec.Golang),
		source.Spec.LogTopicAgent)
}

func makeSourceCommand(source *v1alpha1.Source) []string {
	spec := source.Spec
	hasPulsarctl := source.Spec.ImageHasPulsarctl
	hasWget := source.Spec.ImageHasWget
	if match, _ := regexp.MatchString(RunnerImageHasPulsarctl, source.Spec.Image); match {
		hasPulsarctl = true
		hasWget = true
	}
	return MakeJavaFunctionCommand(spec.Java.JarLocation, spec.Java.Jar,
		spec.Name, spec.ClusterName,
		generateJavaLogConfigCommand(spec.Java, spec.LogTopicAgent),
		parseJavaLogLevel(spec.Java),
		generateSourceDetailsInJSON(source),
		spec.Java.ExtraDependenciesDir, string(source.UID),
		spec.Java.JavaOpts, hasPulsarctl, hasWget, spec.Pulsar.AuthSecret != "", spec.Pulsar.TLSSecret != "",
		spec.SecretsMap, spec.StateConfig, spec.Pulsar.TLSConfig, spec.Pulsar.AuthConfig, nil,
		generateJavaLogConfigFileName(spec.Java))
}

func generateSourceDetailsInJSON(source *v1alpha1.Source) string {
	sourceDetails := convertSourceDetails(source)
	json, err := protojson.Marshal(sourceDetails)
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

func MakeSourceCleanUpJob(source *v1alpha1.Source) *v1.Job {
	labels := makeSourceLabels(source)
	labels["owner"] = string(source.GetUID())
	objectMeta := &metav1.ObjectMeta{
		Name:      makeJobName(source.Name, v1alpha1.SourceComponent) + "-cleanup",
		Namespace: source.Namespace,
		Labels:    labels,
	}
	container := makeSourceContainer(source)
	container.LivenessProbe = nil
	container.Name = CleanupContainerName
	authConfig := source.Spec.Pulsar.CleanupAuthConfig
	if authConfig == nil {
		authConfig = source.Spec.Pulsar.AuthConfig
	}
	volumeMounts := makeSourceVolumeMounts(source, authConfig)
	container.VolumeMounts = volumeMounts
	if source.Spec.CleanupImage != "" {
		container.Image = source.Spec.CleanupImage
	}

	hasPulsarctl := source.Spec.ImageHasPulsarctl
	if match, _ := regexp.MatchString(RunnerImageHasPulsarctl, source.Spec.Image); match {
		hasPulsarctl = true
	}
	command := getCleanUpCommand(hasPulsarctl,
		source.Spec.Pulsar.AuthSecret != "",
		source.Spec.Pulsar.TLSSecret != "",
		source.Spec.Pulsar.TLSConfig,
		source.Spec.Pulsar.AuthConfig,
		[]string{computeBatchSourceIntermediateTopicName(source)},
		"",
		computeBatchSourceInstanceSubscriptionName(source),
		source.Spec.Tenant,
		source.Spec.Namespace,
		source.Spec.Name, true)
	container.Command = command
	return makeCleanUpJob(objectMeta, container, makeSourceVolumes(source, authConfig), makeSourceLabels(source), source.Spec.Pod)
}

func computeBatchSourceInstanceSubscriptionName(source *v1alpha1.Source) string {
	return fmt.Sprintf("BatchSourceExecutor-%s/%s/%s", source.Spec.Tenant, source.Spec.Namespace, source.Spec.Name)
}

func computeBatchSourceIntermediateTopicName(source *v1alpha1.Source) string {
	return fmt.Sprintf("persistent://%s/%s/%s-intermediate", source.Spec.Tenant, source.Spec.Namespace, source.Spec.Name)
}
