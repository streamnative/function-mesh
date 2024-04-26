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
var functionlog = logf.Log.WithName("function-resource")

type FunctionWebhook struct {
	v1alpha1.Function
}

func (webhook *FunctionWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&webhook.Function).
		WithDefaulter(webhook).
		WithValidator(webhook).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// +kubebuilder:webhook:path=/mutate-compute-functionmesh-io-v1alpha1-function,mutating=true,failurePolicy=fail,groups=compute.functionmesh.io,resources=functions,verbs=create;update,versions=v1alpha1,name=mfunction.kb.io,sideEffects=none,admissionReviewVersions={v1beta1,v1}

var _ admission.CustomDefaulter = &FunctionWebhook{}

// Default implements admission.CustomDefaulter so a webhook will be registered for the type
func (webhook *FunctionWebhook) Default(ctx context.Context, obj runtime.Object) error {
	req, err := admission.RequestFromContext(ctx)
	if err != nil {
		return fmt.Errorf("expected admission.Request in ctx: %w", err)
	}
	if req.Kind.Kind != functionKind {
		return fmt.Errorf("expected Kind %q got %q", functionKind, req.Kind.Kind)
	}

	r := obj.(*v1alpha1.Function) //nolint:ifshort
	functionlog.Info("default", "name", r.Name)

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

	if r.Spec.MaxPendingAsyncRequests == nil {
		maxPending := int32(1000)
		r.Spec.MaxPendingAsyncRequests = &maxPending
	}

	if r.Spec.ForwardSourceMessageProperty == nil {
		trueVal := true
		r.Spec.ForwardSourceMessageProperty = &trueVal
	}

	if r.Spec.Resources.Requests != nil {
		if r.Spec.Resources.Requests.Cpu() == nil || r.Spec.Resources.Requests.Cpu().IsZero() {
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

	if r.Spec.Output.TypeClassName == "" {
		r.Spec.Output.TypeClassName = "[B"
	}

	// mutate windowConfig
	if r.Spec.WindowConfig != nil {
		if r.Spec.WindowConfig.WindowLengthDurationMs != nil && r.Spec.WindowConfig.SlidingIntervalDurationMs == nil {
			r.Spec.WindowConfig.SlidingIntervalDurationMs = r.Spec.WindowConfig.WindowLengthDurationMs
		}
		if r.Spec.WindowConfig.WindowLengthCount != nil && r.Spec.WindowConfig.SlidingIntervalCount == nil {
			if r.Spec.WindowConfig.MaxLagMs == nil {
				*r.Spec.WindowConfig.MaxLagMs = 0
			}
			if r.Spec.WindowConfig.WatermarkEmitIntervalMs == nil {
				*r.Spec.WindowConfig.WatermarkEmitIntervalMs = 1000
			}
		}
	}

	if r.Spec.LogTopic != "" && r.Spec.LogTopicAgent == "" {
		r.Spec.LogTopicAgent = v1alpha1.RUNTIME
	}

	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-compute-functionmesh-io-v1alpha1-function,mutating=false,failurePolicy=fail,groups=compute.functionmesh.io,resources=functions,versions=v1alpha1,name=vfunction.kb.io,sideEffects=none,admissionReviewVersions={v1beta1,v1}

var _ admission.CustomValidator = &FunctionWebhook{}

// ValidateCreate implements admission.CustomValidator so a webhook will be registered for the type
func (webhook *FunctionWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	req, err := admission.RequestFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("expected admission.Request in ctx: %w", err)
	}
	if req.Kind.Kind != functionKind {
		return nil, fmt.Errorf("expected Kind %q got %q", functionKind, req.Kind.Kind)
	}

	r := obj.(*v1alpha1.Function) //nolint:ifshort
	functionlog.Info("validate create function", "name", r.Name)
	var allErrs field.ErrorList
	var fieldErr *field.Error
	var fieldErrs []*field.Error

	if len(r.Name) > maxNameLength {
		allErrs = append(allErrs, field.Invalid(field.NewPath("name"), r.Name, fmt.Sprintf("function name must be no more than %d characters", maxNameLength)))
	}

	if r.Name == "" {
		allErrs = append(allErrs, field.Invalid(field.NewPath("name"), r.Name, "function name is not provided"))
	}

	if r.Spec.Runtime.Java == nil && r.Spec.Runtime.Python == nil && r.Spec.Runtime.Golang == nil && r.Spec.GenericRuntime == nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("runtime", "java"), r.Spec.Runtime.Java,
			"runtime cannot be empty"))
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("runtime", "python"), r.Spec.Runtime.Python,
			"runtime cannot be empty"))
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("runtime", "golang"), r.Spec.Runtime.Golang,
			"runtime cannot be empty"))
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("runtime", "genericRuntime"), r.Spec.Runtime.GenericRuntime,
			"runtime cannot be empty"))
	}

	if (r.Spec.Runtime.Java != nil && r.Spec.Runtime.Python != nil) ||
		(r.Spec.Runtime.Java != nil && r.Spec.Runtime.Golang != nil) ||
		(r.Spec.Runtime.Python != nil && r.Spec.Runtime.Golang != nil) ||
		(r.Spec.Runtime.Java != nil && r.Spec.Runtime.Python != nil && r.Spec.Runtime.Golang != nil) {
		allErrs = append(allErrs,
			field.Invalid(field.NewPath("spec").Child("runtime"), r.Spec.Runtime, "you can only specify one runtime"))
	}

	fieldErrs = validateJavaRuntime(r.Spec.Java, r.Spec.ClassName)
	if len(fieldErrs) > 0 {
		allErrs = append(allErrs, fieldErrs...)
	}

	fieldErrs = validatePythonRuntime(r.Spec.Python, r.Spec.ClassName)
	if len(fieldErrs) > 0 {
		allErrs = append(allErrs, fieldErrs...)
	}

	fieldErrs = validateGolangRuntime(r.Spec.Golang)
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

	fieldErr = validateRetainKeyOrdering(r.Spec.RetainKeyOrdering, r.Spec.ProcessingGuarantee)
	if fieldErr != nil {
		allErrs = append(allErrs, fieldErr)
	}

	fieldErrs = validateRetainOrderingConflicts(r.Spec.RetainKeyOrdering, r.Spec.RetainOrdering)
	if len(fieldErrs) > 0 {
		allErrs = append(allErrs, fieldErrs...)
	}

	fieldErr = validateFunctionConfig(r.Spec.FuncConfig)
	if fieldErr != nil {
		allErrs = append(allErrs, fieldErr)
	}

	fieldErr = validateSecretsMap(r.Spec.SecretsMap)
	if fieldErr != nil {
		allErrs = append(allErrs, fieldErr)
	}

	fieldErrs = validateInputOutput(&r.Spec.Input, &r.Spec.Output)
	if len(fieldErrs) > 0 {
		allErrs = append(allErrs, fieldErrs...)
	}

	fieldErr = validateLogTopic(r.Spec.LogTopic)
	if fieldErr != nil {
		allErrs = append(allErrs, fieldErr)
	}

	fieldErr = validateDeadLetterTopic(r.Spec.DeadLetterTopic)
	if fieldErr != nil {
		allErrs = append(allErrs, fieldErr)
	}

	fieldErr = validateStatefulFunctionConfigs(r.Spec.StateConfig, r.Spec.Runtime)
	if fieldErr != nil {
		allErrs = append(allErrs, fieldErr)
	}

	fieldErr = validateWindowConfigs(r.Spec.WindowConfig)
	if fieldErr != nil {
		allErrs = append(allErrs, fieldErr)
	}

	fieldErr = validateMessaging(&r.Spec.Messaging)
	if fieldErr != nil {
		allErrs = append(allErrs, fieldErr)
	}

	fieldErr = validateBuiltinHPARules(r.Spec.Pod.BuiltinAutoscaler)
	if fieldErr != nil {
		allErrs = append(allErrs, fieldErr)
	}

	if len(allErrs) == 0 {
		return nil, nil
	}

	return nil, apierrors.NewInvalid(schema.GroupKind{Group: "compute.functionmesh.io", Kind: "FunctionWebhook"}, r.Name, allErrs)
}

// ValidateUpdate implements admission.CustomValidator so a webhook will be registered for the type
func (webhook *FunctionWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	req, err := admission.RequestFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("expected admission.Request in ctx: %w", err)
	}
	if req.Kind.Kind != functionKind {
		return nil, fmt.Errorf("expected Kind %q got %q", functionKind, req.Kind.Kind)
	}

	r := oldObj.(*v1alpha1.Function) //nolint:ifshort
	functionlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil, nil
}

// ValidateDelete implements admission.CustomValidator so a webhook will be registered for the type
func (webhook *FunctionWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	req, err := admission.RequestFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("expected admission.Request in ctx: %w", err)
	}
	if req.Kind.Kind != functionKind {
		return nil, fmt.Errorf("expected Kind %q got %q", functionKind, req.Kind.Kind)
	}

	r := obj.(*v1alpha1.Function) //nolint:ifshort
	functionlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
