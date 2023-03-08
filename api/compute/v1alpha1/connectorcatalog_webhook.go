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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var connectorcataloglog = logf.Log.WithName("connectorcatalog-resource")

func (r *ConnectorCatalog) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

//+kubebuilder:webhook:path=/mutate-compute-functionmesh-io-v1alpha1-connectorcatalog,mutating=true,failurePolicy=fail,sideEffects=None,groups=compute.functionmesh.io,resources=connectorcatalogs,verbs=create;update,versions=v1alpha1,name=mconnectorcatalog.kb.io,admissionReviewVersions=v1

var _ webhook.Defaulter = &ConnectorCatalog{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *ConnectorCatalog) Default() {
	connectorcataloglog.Info("default", "name", r.Name)

	for _, d := range r.Spec.ConnectorDefinitions {
		if d.Version == "" {
			d.Version = d.ImageTag
		}
	}
}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-compute-functionmesh-io-v1alpha1-connectorcatalog,mutating=false,failurePolicy=fail,sideEffects=None,groups=compute.functionmesh.io,resources=connectorcatalogs,verbs=create;update,versions=v1alpha1,name=vconnectorcatalog.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &ConnectorCatalog{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *ConnectorCatalog) ValidateCreate() error {
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
func (r *ConnectorCatalog) ValidateUpdate(old runtime.Object) error {
	connectorcataloglog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *ConnectorCatalog) ValidateDelete() error {
	connectorcataloglog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
