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

package v1alpha1

import (
	"fmt"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var sinklog = logf.Log.WithName("sink-resource")

func (r *Sink) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-compute-functionmesh-io-v1alpha1-sink,mutating=true,failurePolicy=fail,groups=compute.functionmesh.io,resources=sinks,verbs=create;update,versions=v1alpha1,name=msink.kb.io,sideEffects=none,admissionReviewVersions={v1beta1,v1}

var _ webhook.Defaulter = &Sink{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Sink) Default() {
	sinklog.Info("default", "name", r.Name)

	if !(r.Spec.Replicas != nil && r.Spec.MinReplicas != nil) {
		if r.Spec.MinReplicas != nil && r.Spec.Replicas == nil {
			r.Spec.Replicas = new(int32)
			*r.Spec.Replicas = *r.Spec.MinReplicas
		} else if r.Spec.MinReplicas == nil && r.Spec.Replicas != nil {
			r.Spec.MinReplicas = new(int32)
			*r.Spec.MinReplicas = *r.Spec.Replicas
		} else {
			r.Spec.Replicas = new(int32)
			*r.Spec.Replicas = 1
			r.Spec.MinReplicas = new(int32)
			*r.Spec.MinReplicas = 1
		}
	}

	if r.Spec.AutoAck == nil {
		r.Spec.AutoAck = new(bool)
		*r.Spec.AutoAck = true
	}

	if r.Spec.ProcessingGuarantee == "" {
		r.Spec.ProcessingGuarantee = AtleastOnce
	}

	if r.Spec.Name == "" {
		r.Spec.Name = r.Name
	}

	if r.Spec.ClusterName == "" {
		r.Spec.ClusterName = DefaultCluster
	}

	if r.Spec.Tenant == "" {
		r.Spec.Tenant = DefaultTenant
	}

	if r.Spec.Namespace == "" {
		r.Spec.Namespace = DefaultNamespace
	}

	if r.Spec.Resources.Requests != nil {
		if r.Spec.Resources.Requests.Cpu() == nil {
			r.Spec.Resources.Requests.Cpu().Set(DefaultResourceCPU)
		}

		if r.Spec.Resources.Requests.Memory() == nil {
			r.Spec.Resources.Requests.Memory().Set(DefaultResourceMemory)
		}
	}

	if r.Spec.Resources.Limits == nil {
		paddingResourceLimit(&r.Spec.Resources)
	}

	if r.Spec.Input.TypeClassName == "" {
		r.Spec.Input.TypeClassName = "[B"
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-compute-functionmesh-io-v1alpha1-sink,mutating=false,failurePolicy=fail,groups=compute.functionmesh.io,resources=sinks,versions=v1alpha1,name=vsink.kb.io,sideEffects=none,admissionReviewVersions={v1beta1,v1}

var _ webhook.Validator = &Sink{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Sink) ValidateCreate() error {
	sinklog.Info("validate create sink", "name", r.Name)
	var allErrs field.ErrorList
	var fieldErr *field.Error
	var fieldErrs []*field.Error

	if len(r.Name) > maxNameLength {
		allErrs = append(allErrs, field.Invalid(field.NewPath("name"), r.Name, fmt.Sprintf("sink name must be no more than %d characters", maxNameLength)))
	}

	if r.Spec.SinkConfig == nil {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spec").Child("sinkConfig"), r.Spec.SinkConfig, "sink config is not provided"))
	}

	if r.Spec.Runtime.Java == nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("runtime", "java"), r.Spec.Runtime.Java,
			"sink must have java runtime specified"))
	}

	fieldErrs = validateJavaRuntime(r.Spec.Java, r.Spec.ClassName)
	if len(fieldErrs) > 0 {
		allErrs = append(allErrs, fieldErrs...)
	}

	fieldErrs = validateReplicasAndMinReplicasAndMaxReplicas(r.Spec.Replicas, r.Spec.MinReplicas, r.Spec.MaxReplicas)
	if len(fieldErrs) > 0 {
		allErrs = append(allErrs, fieldErrs...)
	}

	if r.Spec.Pod.VPA != nil {
		if r.Spec.MaxReplicas != nil {
			allErrs = append(allErrs,
				field.Invalid(field.NewPath("spec").Child("pod").Child("vpa"), r.Spec.Pod.VPA, "you can not enable HPA and VPA at the same time"))
		}

		if r.Spec.Pod.VPA.ResourcePolicy != nil {
			for _, c := range r.Spec.Pod.VPA.ResourcePolicy.ContainerPolicies {
				if c.ContainerName == "" {
					allErrs = append(allErrs,
						field.Invalid(field.NewPath("spec").Child("pod").Child("vpa").Child("resourcePolicy").Child("containerPolicy"), r.Spec.Pod.VPA.ResourcePolicy.ContainerPolicies, "container name must be specified"))
					break
				}
			}
		}
	}

	fieldErr = validateResourceRequirement(r.Spec.Resources)
	if fieldErr != nil {
		allErrs = append(allErrs, fieldErr)
	}

	fieldErr = validateAutoAck(r.Spec.AutoAck)
	if fieldErr != nil {
		allErrs = append(allErrs, fieldErr)
	}

	fieldErr = validateTimeout(r.Spec.Timeout, r.Spec.ProcessingGuarantee)
	if fieldErr != nil {
		allErrs = append(allErrs, fieldErr)
	}

	fieldErrs = validateMaxMessageRetry(r.Spec.MaxMessageRetry, r.Spec.ProcessingGuarantee, r.Spec.DeadLetterTopic)
	if len(fieldErrs) > 0 {
		allErrs = append(allErrs, fieldErrs...)
	}

	fieldErr = validateSinkConfig(r.Spec.SinkConfig)
	if fieldErr != nil {
		allErrs = append(allErrs, fieldErr)
	}

	fieldErr = validateSecretsMap(r.Spec.SecretsMap)
	if fieldErr != nil {
		allErrs = append(allErrs, fieldErr)
	}

	fieldErrs = validateInputOutput(&r.Spec.Input, nil)
	if len(fieldErrs) > 0 {
		allErrs = append(allErrs, fieldErrs...)
	}

	fieldErr = validateDeadLetterTopic(r.Spec.DeadLetterTopic)
	if fieldErr != nil {
		allErrs = append(allErrs, fieldErr)
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "compute.functionmesh.io", Kind: "Sink"}, r.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Sink) ValidateUpdate(old runtime.Object) error {
	sinklog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Sink) ValidateDelete() error {
	sinklog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
