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

package webhook

import (
	"context"
	"fmt"
	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var connectorcataloglog = logf.Log.WithName("connectorcatalog-resource")

type ConnectorWebhook struct {
	v1alpha1.ConnectorCatalog
}

func (webhook *ConnectorWebhook) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&webhook.ConnectorCatalog).
		WithDefaulter(webhook).
		WithValidator(webhook).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-compute-functionmesh-io-v1alpha1-connectorcatalog,mutating=true,failurePolicy=fail,sideEffects=None,groups=compute.functionmesh.io,resources=connectorcatalogs,verbs=create;update,versions=v1alpha1,name=mconnectorcatalog.kb.io,admissionReviewVersions=v1

var _ admission.CustomDefaulter = &ConnectorWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (webhook *ConnectorWebhook) Default(ctx context.Context, obj runtime.Object) error {
	req, err := admission.RequestFromContext(ctx)
	if err != nil {
		return fmt.Errorf("expected admission.Request in ctx: %w", err)
	}
	if req.Kind.Kind != connectorCatalogKind {
		return fmt.Errorf("expected Kind %q got %q", connectorCatalogKind, req.Kind.Kind)
	}

	r := obj.(*v1alpha1.ConnectorCatalog) //nolint:ifshort
	connectorcataloglog.Info("default", "name", r.Name)

	for _, d := range r.Spec.ConnectorDefinitions {
		if d.Version == "" {
			d.Version = d.ImageTag
		}
	}
	return nil
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-compute-functionmesh-io-v1alpha1-connectorcatalog,mutating=false,failurePolicy=fail,sideEffects=None,groups=compute.functionmesh.io,resources=connectorcatalogs,verbs=create;update,versions=v1alpha1,name=vconnectorcatalog.kb.io,admissionReviewVersions=v1

var _ admission.CustomValidator = &ConnectorWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (webhook *ConnectorWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) error {
	req, err := admission.RequestFromContext(ctx)
	if err != nil {
		return fmt.Errorf("expected admission.Request in ctx: %w", err)
	}
	if req.Kind.Kind != connectorCatalogKind {
		return fmt.Errorf("expected Kind %q got %q", connectorCatalogKind, req.Kind.Kind)
	}

	r := obj.(*v1alpha1.ConnectorCatalog) //nolint:ifshort
	connectorcataloglog.Info("validate create", "name", r.Name)
	var allErrs field.ErrorList

	if r.Spec.ConnectorDefinitions == nil {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("connectorDefinitions"), r.Spec.ConnectorDefinitions,
			"connectorDefinitions is not provided"))
	}

	for _, def := range r.Spec.ConnectorDefinitions {
		if def.ID == "" {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("connectorDefinitions"), r.Spec.ConnectorDefinitions, "No Id specified"))
		}

		if def.Version == "" {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("connectorDefinitions"), r.Spec.ConnectorDefinitions, "No Version specified"))
		}

		if def.ImageRepository == "" {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("connectorDefinitions"), r.Spec.ConnectorDefinitions, "No ImageRepository specified"))
		}

		if def.ImageTag == "" {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("connectorDefinitions"), r.Spec.ConnectorDefinitions, "No ImageTag specified"))
		}

		if def.Name == "" {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("connectorDefinitions"), r.Spec.ConnectorDefinitions, "No Name specified"))
		}

		if def.Description == "" {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("connectorDefinitions"), r.Spec.ConnectorDefinitions, "No Description specified"))
		}

		if def.SourceClass == "" && def.SinkClass == "" {
			allErrs = append(allErrs, field.Invalid(field.NewPath("spec").Child("connectorDefinitions"), r.Spec.ConnectorDefinitions, "Either SourceClass or SinkClass must be specified"))
		}
	}

	if len(allErrs) == 0 {
		return nil
	}

	return apierrors.NewInvalid(schema.GroupKind{Group: "compute.functionmesh.io", Kind: "ConnectorCatalog"}, r.Name, allErrs)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (webhook *ConnectorWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) error {
	req, err := admission.RequestFromContext(ctx)
	if err != nil {
		return fmt.Errorf("expected admission.Request in ctx: %w", err)
	}
	if req.Kind.Kind != connectorCatalogKind {
		return fmt.Errorf("expected Kind %q got %q", connectorCatalogKind, req.Kind.Kind)
	}

	r := oldObj.(*v1alpha1.ConnectorCatalog) //nolint:ifshort
	connectorcataloglog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (webhook *ConnectorWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) error {
	req, err := admission.RequestFromContext(ctx)
	if err != nil {
		return fmt.Errorf("expected admission.Request in ctx: %w", err)
	}
	if req.Kind.Kind != connectorCatalogKind {
		return fmt.Errorf("expected Kind %q got %q", connectorCatalogKind, req.Kind.Kind)
	}

	r := obj.(*v1alpha1.ConnectorCatalog) //nolint:ifshort
	connectorcataloglog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
