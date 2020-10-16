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
	"encoding/json"
	"errors"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var functionlog = logf.Log.WithName("function-resource")

func (r *Function) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// +kubebuilder:webhook:verbs=create;update,path=/validate-cloud-streamnative-io-streamnative-io-v1alpha1-function,mutating=false,failurePolicy=fail,groups=cloud.streamnative.io.streamnative.io,resources=functions,versions=v1alpha1,name=vfunction.kb.io

var _ webhook.Defaulter = &Function{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Function) Default() {
	functionlog.Info("default", "name", r.Name)

	if r.Spec.Replicas == nil {
		zeroVal := int32(0)
		r.Spec.Replicas = &zeroVal
	}

	if r.Spec.AutoAck == nil {
		trueVal := true
		r.Spec.AutoAck = &trueVal
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

	if r.Spec.MaxPendingAsyncRequests == nil {
		maxPending := int32(1000)
		r.Spec.MaxPendingAsyncRequests = &maxPending
	}

	if r.Spec.ForwardSourceMessageProperty == nil {
		trueVal := true
		r.Spec.ForwardSourceMessageProperty = &trueVal
	}

	if r.Spec.Resources.Cpu() == nil {
		r.Spec.Resources.Cpu().Set(int64(1))
	}

	if r.Spec.Resources.Memory() == nil {
		r.Spec.Resources.Memory().Set(int64(1073741824))
	}

}

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
// +kubebuilder:webhook:verbs=create;update,path=/validate-cloud-streamnative-io-streamnative-io-v1alpha1-function,mutating=false,failurePolicy=fail,groups=cloud.streamnative.io.streamnative.io,resources=functions,versions=v1alpha1,name=vfunction.kb.io

var _ webhook.Validator = &Function{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Function) ValidateCreate() error {
	functionlog.Info("validate create", "name", r.Name)

	if r.Spec.Java != nil {
		if r.Spec.ClassName == "" {
			return errors.New("class name cannot be empty")
		}
	}

	// TODO: verify source conf

	// TODO: allow 0 replicas, currently hpa's min value has to be 1
	if *r.Spec.Replicas == 0 {
		return errors.New("replicas cannot be zero")
	}

	if *r.Spec.MaxReplicas != 0 && *r.Spec.Replicas > *r.Spec.MaxReplicas {
		return errors.New("maxReplicas must be greater than or equal to replicas")
	}

	if !validResource(r.Spec.Resources) {
		return errors.New("resource request is invalid. each resource value must be positive")
	}

	if r.Spec.Timeout != 0 && r.Spec.ProcessingGuarantee != AtleastOnce {
		return errors.New("message timeout can only be set for AtleastOnce processing guarantee")
	}

	if r.Spec.MaxMessageRetry > 0 && r.Spec.ProcessingGuarantee == EffectivelyOnce {
		return errors.New("MaxMessageRetries and Effectively once are not compatible")
	}

	if r.Spec.MaxMessageRetry <= 0 && r.Spec.DeadLetterTopic != "" {
		return errors.New("dead letter topic is set but max message retry is set to infinity")
	}

	if r.Spec.RetainKeyOrdering && r.Spec.ProcessingGuarantee == EffectivelyOnce {
		return errors.New("when effectively once processing guarantee is specified, retain Key ordering cannot be set")
	}

	if r.Spec.RetainKeyOrdering && r.Spec.RetainOrdering {
		return errors.New("only one of retain ordering or retain key ordering can be set")
	}

	if r.Spec.Java == nil && r.Spec.Python == nil && r.Spec.Golang == nil {
		return errors.New("must specify a runtime from java, python or golang")
	}

	if r.Spec.FuncConfig != nil {
		_, err := json.Marshal(r.Spec.FuncConfig)
		if err != nil {
			return errors.New("provided config is wrong: " + err.Error())
		}
	}

	if r.Spec.SecretsMap != nil {
		_, err := json.Marshal(r.Spec.SecretsMap)
		if err != nil {
			return errors.New("provided secrets map is wrong: " + err.Error())
		}
	}
	// TODO python/golang specific check

	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Function) ValidateUpdate(old runtime.Object) error {
	functionlog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Function) ValidateDelete() error {
	functionlog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
