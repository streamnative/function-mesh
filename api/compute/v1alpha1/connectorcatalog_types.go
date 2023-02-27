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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ConnectorCatalogSpec defines the desired state of ConnectorCatalog
type ConnectorCatalogSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ConnectorDefinitions []ConnectorDefinition `json:"connectorDefinitions"`
}

// ConnectorCatalogStatus defines the observed state of ConnectorCatalog
type ConnectorCatalogStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ConnectorCatalog is the Schema for the connectorcatalogs API
type ConnectorCatalog struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConnectorCatalogSpec   `json:"spec,omitempty"`
	Status ConnectorCatalogStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ConnectorCatalogList contains a list of ConnectorCatalog
type ConnectorCatalogList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConnectorCatalog `json:"items"`
}

type ConfigFieldDefinition struct {
	FieldName  string            `json:"fieldName"`
	TypeName   string            `json:"typeName"`
	Attributes map[string]string `json:"attributes"`
}

type ConnectorDefinition struct {
	Id                     string                  `json:"id"`
	Version                string                  `json:"version"`
	ImageRegistry          string                  `json:"imageRegistry"`
	ImageRepository        string                  `json:"imageRepository"`
	ImageTag               string                  `json:"imageTag"`
	TypeClassName          string                  `json:"typeClassName"`
	SourceTypeClassName    string                  `json:"sourceTypeClassName"`
	SinkTypeClassName      string                  `json:"sinkTypeClassName"`
	JarFullName            string                  `json:"jarFullName"`
	DefaultSchemaType      string                  `json:"defaultSchemaType"`
	DefaultSerdeClassName  string                  `json:"defaultSerdeClassName"`
	Name                   string                  `json:"name"`
	Description            string                  `json:"description"`
	SourceClass            string                  `json:"sourceClass"`
	SinkClass              string                  `json:"sinkClass"`
	SourceConfigClass      string                  `json:"sourceConfigClass"`
	SinkConfigClass        string                  `json:"sinkConfigClass"`
	ConfigFieldDefinitions []ConfigFieldDefinition `json:"configFieldDefinitions"`
}

func init() {
	SchemeBuilder.Register(&ConnectorCatalog{}, &ConnectorCatalogList{})
}
