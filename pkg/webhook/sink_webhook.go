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

// Package webhook defines mutate and validate webhook for FunctionMesh types
package webhook

import (
	"context"
	"fmt"

	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var sinklog = logf.Log.WithName("sink-resource")

type SinkWebhook struct {
	v1alpha1.Sink
}

func (webhook *SinkWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&webhook.Sink).
		WithDefaulter(webhook).
		WithValidator(webhook).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:path=/mutate-compute-functionmesh-io-v1alpha1-sink,mutating=true,failurePolicy=fail,groups=compute.functionmesh.io,resources=sinks,verbs=create;update,versions=v1alpha1,name=msink.kb.io,sideEffects=none,admissionReviewVersions={v1beta1,v1}

var _ admission.CustomDefaulter = &SinkWebhook{}

// Default implements admission.CustomDefaulter so a webhook will be registered for the type
func (webhook *SinkWebhook) Default(ctx context.Context, obj runtime.Object) error {
	req, err := admission.RequestFromContext(ctx)
	if err != nil {
		return fmt.Errorf("expected admission.Request in ctx: %w", err)
	}
	if req.Kind.Kind != sinkKind {
		return fmt.Errorf("expected Kind %q got %q", sinkKind, req.Kind.Kind)
	}

	r := obj.(*v1alpha1.Sink) //nolint:ifshort
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
		r.Spec.ProcessingGuarantee = v1alpha1.AtleastOnce
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

	if r.Spec.LogTopic != "" && r.Spec.LogTopicAgent == "" {
		r.Spec.LogTopicAgent = v1alpha1.SIDECAR
	}

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-compute-functionmesh-io-v1alpha1-sink,mutating=false,failurePolicy=fail,groups=compute.functionmesh.io,resources=sinks,versions=v1alpha1,name=vsink.kb.io,sideEffects=none,admissionReviewVersions={v1beta1,v1}

var _ admission.CustomValidator = &SinkWebhook{}

// ValidateCreate implements admission.CustomValidator so a webhook will be registered for the type
func (webhook *SinkWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	req, err := admission.RequestFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("expected admission.Request in ctx: %w", err)
	}
	if req.Kind.Kind != sinkKind {
		return nil, fmt.Errorf("expected Kind %q got %q", sinkKind, req.Kind.Kind)
	}

	r := obj.(*v1alpha1.Sink) //nolint:ifshort
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

		resourceErrors := validateResourcePolicy(r.Spec.Pod.VPA.ResourcePolicy)
		if resourceErrors != nil {
			allErrs = append(allErrs, resourceErrors...)
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

	fieldErr = validateMessaging(&r.Spec.Messaging)
	if fieldErr != nil {
		allErrs = append(allErrs, fieldErr)
	}

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(schema.GroupKind{Group: "compute.functionmesh.io", Kind: "SinkWebhook"}, r.Name, allErrs)
}

// ValidateUpdate implements admission.CustomValidator so a webhook will be registered for the type
func (webhook *SinkWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	req, err := admission.RequestFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("expected admission.Request in ctx: %w", err)
	}
	if req.Kind.Kind != sinkKind {
		return nil, fmt.Errorf("expected Kind %q got %q", sinkKind, req.Kind.Kind)
	}

	r := oldObj.(*v1alpha1.Sink) //nolint:ifshort
	sinklog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements admission.CustomValidator so a webhook will be registered for the type
func (webhook *SinkWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	req, err := admission.RequestFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("expected admission.Request in ctx: %w", err)
	}
	if req.Kind.Kind != sinkKind {
		return nil, fmt.Errorf("expected Kind %q got %q", sinkKind, req.Kind.Kind)
	}

	r := obj.(*v1alpha1.Sink) //nolint:ifshort
	sinklog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
