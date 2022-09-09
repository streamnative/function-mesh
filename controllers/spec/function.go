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
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// log is for logging in this package.
var log = logf.Log.WithName("function-resource")

func MakeFunctionHPA(function *v1alpha1.Function) *autov2beta2.HorizontalPodAutoscaler {
	objectMeta := MakeFunctionObjectMeta(function)
	targetRef := autov2beta2.CrossVersionObjectReference{
		Kind:       function.Kind,
		Name:       function.Name,
		APIVersion: function.APIVersion,
	}
	if isBuiltinHPAEnabled(function.Spec.MinReplicas, function.Spec.MaxReplicas, function.Spec.Pod) {
		return makeBuiltinHPA(objectMeta, *function.Spec.MinReplicas, *function.Spec.MaxReplicas, targetRef,
			function.Spec.Pod.BuiltinAutoscaler)
	} else if !isDefaultHPAEnabled(function.Spec.MinReplicas, function.Spec.MaxReplicas, function.Spec.Pod) {
		return makeHPA(objectMeta, *function.Spec.MinReplicas, *function.Spec.MaxReplicas, function.Spec.Pod, targetRef)
	}
	return makeDefaultHPA(objectMeta, *function.Spec.MinReplicas, *function.Spec.MaxReplicas, targetRef)
}

func MakeFunctionService(function *v1alpha1.Function) *corev1.Service {
	labels := makeFunctionLabels(function)
	objectMeta := MakeFunctionObjectMeta(function)
	return MakeService(objectMeta, labels)
}

func MakeFunctionStatefulSet(function *v1alpha1.Function) *appsv1.StatefulSet {
	objectMeta := MakeFunctionObjectMeta(function)
	return MakeStatefulSet(objectMeta, function.Spec.Replicas,
		MakeFunctionContainer(function), makeFunctionVolumes(function), makeFunctionLabels(function), function.Spec.Pod)
}

func MakeFunctionObjectMeta(function *v1alpha1.Function) *metav1.ObjectMeta {
	return &metav1.ObjectMeta{
		Name:      makeJobName(function.Name, v1alpha1.FunctionComponent),
		Namespace: function.Namespace,
		Labels:    makeFunctionLabels(function),
		OwnerReferences: []metav1.OwnerReference{
			*metav1.NewControllerRef(function, function.GroupVersionKind()),
		},
	}
}

func makeFunctionVolumes(function *v1alpha1.Function) []corev1.Volume {
	return generatePodVolumes(function.Spec.Pod.Volumes,
		function.Spec.Output.ProducerConf,
		function.Spec.Input.SourceSpecs,
		function.Spec.Pulsar.TLSConfig,
		function.Spec.Pulsar.AuthConfig,
		getRuntimeLogConfigNames(function.Spec.Java, function.Spec.Python, function.Spec.Golang))
}

func makeFunctionVolumeMounts(function *v1alpha1.Function) []corev1.VolumeMount {
	return generateContainerVolumeMounts(function.Spec.VolumeMounts,
		function.Spec.Output.ProducerConf,
		function.Spec.Input.SourceSpecs,
		function.Spec.Pulsar.TLSConfig,
		function.Spec.Pulsar.AuthConfig,
		getRuntimeLogConfigNames(function.Spec.Java, function.Spec.Python, function.Spec.Golang))
}

func MakeFunctionContainer(function *v1alpha1.Function) *corev1.Container {
	imagePullPolicy := function.Spec.ImagePullPolicy
	if imagePullPolicy == "" {
		imagePullPolicy = corev1.PullIfNotPresent
	}
	return &corev1.Container{
		// TODO new container to pull user code image and upload jars into bookkeeper
		Name:            "pulsar-function",
		Image:           getFunctionRunnerImage(&function.Spec),
		Command:         makeFunctionCommand(function),
		Ports:           []corev1.ContainerPort{GRPCPort, MetricsPort},
		Env:             generateContainerEnv(function),
		Resources:       function.Spec.Resources,
		ImagePullPolicy: imagePullPolicy,
		EnvFrom: generateContainerEnvFrom(function.Spec.Pulsar.PulsarConfig, function.Spec.Pulsar.AuthSecret,
			function.Spec.Pulsar.TLSSecret),
		VolumeMounts: makeFunctionVolumeMounts(function),
	}
}

func makeFunctionLabels(function *v1alpha1.Function) map[string]string {
	jobName := makeJobName(function.Name, v1alpha1.FunctionComponent)
	labels := map[string]string{
		"app.kubernetes.io/name":            jobName,
		"app.kubernetes.io/instance":        jobName,
		"compute.functionmesh.io/app":       AppFunctionMesh,
		"compute.functionmesh.io/component": ComponentFunction,
		"compute.functionmesh.io/name":      function.Name,
		"compute.functionmesh.io/namespace": function.Namespace,
		// The following will be deprecated after two releases
		"app":       AppFunctionMesh,
		"component": ComponentFunction,
		"name":      function.Name,
		"namespace": function.Namespace,
	}
	return labels
}

func makeFunctionCommand(function *v1alpha1.Function) []string {
	spec := function.Spec

	if spec.Java != nil {
		if spec.Java.Jar != "" {
			return MakeJavaFunctionCommand(spec.Java.JarLocation, spec.Java.Jar,
				spec.Name, spec.ClusterName,
				generateJavaLogConfigCommand(function.Spec.Java),
				parseJavaLogLevel(function.Spec.Java),
				generateFunctionDetailsInJSON(function),
				getDecimalSIMemory(spec.Resources.Requests.Memory()), spec.Java.ExtraDependenciesDir, string(function.UID),
				spec.Pulsar.AuthSecret != "", spec.Pulsar.TLSSecret != "", function.Spec.SecretsMap,
				function.Spec.StateConfig, function.Spec.Pulsar.TLSConfig, function.Spec.Pulsar.AuthConfig)
		}
	} else if spec.Python != nil {
		if spec.Python.Py != "" {
			return MakePythonFunctionCommand(spec.Python.PyLocation, spec.Python.Py,
				spec.Name, spec.ClusterName,
				generatePythonLogConfigCommand(function.Name, function.Spec.Python),
				generateFunctionDetailsInJSON(function), string(function.UID),
				spec.Pulsar.AuthSecret != "", spec.Pulsar.TLSSecret != "", function.Spec.SecretsMap,
				function.Spec.StateConfig, function.Spec.Pulsar.TLSConfig, function.Spec.Pulsar.AuthConfig)
		}
	} else if spec.Golang != nil {
		if spec.Golang.Go != "" {
			return MakeGoFunctionCommand(spec.Golang.GoLocation, spec.Golang.Go,
				function)
		}
	}

	return nil
}

func generateFunctionDetailsInJSON(function *v1alpha1.Function) string {
	functionDetails := convertFunctionDetails(function)
	marshaler := &jsonpb.Marshaler{}
	json, err := marshaler.MarshalToString(functionDetails)
	if err != nil {
		// TODO
		panic(err)
	}
	log.Info(json)
	return json
}
