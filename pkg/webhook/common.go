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
	"fmt"
	"strings"

	"github.com/streamnative/function-mesh/api/compute/v1alpha1"
	pctlutil "github.com/streamnative/pulsarctl/pkg/pulsar/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
	vpav1 "k8s.io/autoscaler/vertical-pod-autoscaler/pkg/apis/autoscaling.k8s.io/v1"
)

const (
	functionKind         = "Function"
	sinkKind             = "Sink"
	sourceKind           = "Source"
	connectorCatalogKind = "ConnectorCatalog"

	maxNameLength = 43

	PackageURLHTTP     string = "http://"
	PackageURLHTTPS    string = "https://"
	PackageURLFunction string = "function://"
	PackageURLSource   string = "source://"
	PackageURLSink     string = "sink://"

	DefaultTenant    string = "public"
	DefaultNamespace string = "default"
	DefaultCluster   string = "kubernetes"

	DefaultResourceCPU    int64 = 1
	DefaultResourceMemory int64 = 1073741824
)

func validPackageLocation(packageLocation string) error {
	if hasPackageTypePrefix(packageLocation) {
		err := isValidPulsarPackageURL(packageLocation)
		if err != nil {
			return err
		}
	} else {
		if !isFunctionPackageURLSupported(packageLocation) {
			return fmt.Errorf("invalid function package url %s, supported url (http/https)", packageLocation)
		}
	}

	return nil
}

func hasPackageTypePrefix(packageLocation string) bool {
	lowerCase := strings.ToLower(packageLocation)
	return strings.HasPrefix(lowerCase, PackageURLFunction) ||
		strings.HasPrefix(lowerCase, PackageURLSource) ||
		strings.HasPrefix(lowerCase, PackageURLSink)
}

func isValidPulsarPackageURL(packageLocation string) error {
	parts := strings.Split(packageLocation, "://")
	if len(parts) != 2 {
		return fmt.Errorf("invalid package name %s", packageLocation)
	}
	if !hasPackageTypePrefix(packageLocation) {
		return fmt.Errorf("invalid package name %s", packageLocation)
	}
	rest := parts[1]
	if !strings.Contains(rest, "@") {
		rest += "@"
	}
	packageParts := strings.Split(rest, "@")
	if len(packageParts) != 2 {
		return fmt.Errorf("invalid package name %s", packageLocation)
	}
	partsWithoutVersion := strings.Split(packageParts[0], "/")
	if len(partsWithoutVersion) != 3 {
		return fmt.Errorf("invalid package name %s", packageLocation)
	}
	return nil
}

func isFunctionPackageURLSupported(packageLocation string) bool {
	// TODO: support file:// schema
	lowerCase := strings.ToLower(packageLocation)
	return strings.HasPrefix(lowerCase, PackageURLHTTP) ||
		strings.HasPrefix(lowerCase, PackageURLHTTPS)
}

func validResourceRequirement(requirements corev1.ResourceRequirements) bool {
	if requirements.Limits.Cpu().Sign() == 0 {
		return validResource(requirements.Requests) && validResource(requirements.Limits) &&
			requirements.Requests.Memory().Cmp(*requirements.Limits.Memory()) <= 0
	}
	return validResource(requirements.Requests) && validResource(requirements.Limits) &&
		requirements.Requests.Memory().Cmp(*requirements.Limits.Memory()) <= 0 &&
		requirements.Requests.Cpu().Cmp(*requirements.Limits.Cpu()) <= 0
}

func validResource(resources corev1.ResourceList) bool {
	// memory > 0 and cpu & storage >= 0
	return resources.Cpu().Sign() >= 0 &&
		resources.Memory().Sign() == 1 &&
		resources.Storage().Sign() >= 0
}

func paddingResourceLimit(requirement *corev1.ResourceRequirements) {
	// TODO: better padding calculation
	requirement.Limits.Memory().Set(requirement.Requests.Memory().Value())
	requirement.Limits.Storage().Set(requirement.Requests.Storage().Value())
}

func collectAllInputTopics(inputs v1alpha1.InputConf) []string {
	ret := []string{}
	if len(inputs.Topics) > 0 {
		ret = append(ret, inputs.Topics...)
	}
	if inputs.TopicPattern != "" {
		ret = append(ret, inputs.TopicPattern)
	}
	if len(inputs.CustomSerdeSources) > 0 {
		for k := range inputs.CustomSerdeSources {
			ret = append(ret, k)
		}
	}
	if len(inputs.CustomSchemaSources) > 0 {
		for k := range inputs.CustomSchemaSources {
			ret = append(ret, k)
		}
	}
	if len(inputs.SourceSpecs) > 0 {
		for k := range inputs.SourceSpecs {
			ret = append(ret, k)
		}
	}
	return ret
}

func isValidTopicName(topicName string) error {
	_, err := pctlutil.GetTopicName(topicName)
	return err
}

func validateResourcePolicy(resourcePolicy *vpav1.PodResourcePolicy) field.ErrorList {
	var errs field.ErrorList
	if resourcePolicy != nil {
		for _, c := range resourcePolicy.ContainerPolicies {
			if c.ContainerName == "" {
				errs = append(errs,
					field.Invalid(field.NewPath("spec").Child("pod").Child("vpa").Child("resourcePolicy").Child("containerPolicy"), resourcePolicy.ContainerPolicies, "container name must be specified"))
				break
			}
		}
	}
	return errs
}
