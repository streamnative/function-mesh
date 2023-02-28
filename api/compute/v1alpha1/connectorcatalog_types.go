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

	//+listType=map
	//+listMapKey=id
	ConnectorDefinitions []ConnectorDefinition `json:"connectorDefinitions"`
}

// ConnectorCatalogStatus defines the observed state of ConnectorCatalog
type ConnectorCatalogStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

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
	Attributes map[string]string `json:"attributes,omitempty"`
}

type ConnectorDefinition struct {
	Id                     string                  `json:"id"`
	Version                string                  `json:"version,omitempty"`
	ImageRegistry          string                  `json:"imageRegistry,omitempty"`
	ImageRepository        string                  `json:"imageRepository,omitempty"`
	ImageTag               string                  `json:"imageTag,omitempty"`
	TypeClassName          string                  `json:"typeClassName,omitempty"`
	SourceTypeClassName    string                  `json:"sourceTypeClassName,omitempty"`
	SinkTypeClassName      string                  `json:"sinkTypeClassName,omitempty"`
	JarFullName            string                  `json:"jarFullName,omitempty"`
	DefaultSchemaType      string                  `json:"defaultSchemaType,omitempty"`
	DefaultSerdeClassName  string                  `json:"defaultSerdeClassName,omitempty"`
	Name                   string                  `json:"name,omitempty"`
	Description            string                  `json:"description,omitempty"`
	SourceClass            string                  `json:"sourceClass,omitempty"`
	SinkClass              string                  `json:"sinkClass,omitempty"`
	SourceConfigClass      string                  `json:"sourceConfigClass,omitempty"`
	SinkConfigClass        string                  `json:"sinkConfigClass,omitempty"`
	ConfigFieldDefinitions []ConfigFieldDefinition `json:"configFieldDefinitions,omitempty"`
}

func init() {
	SchemeBuilder.Register(&ConnectorCatalog{}, &ConnectorCatalogList{})
}
