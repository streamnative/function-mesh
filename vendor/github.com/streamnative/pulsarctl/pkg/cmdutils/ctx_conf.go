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

package cmdutils

type Config struct {
	AuthInfos      map[string]*AuthInfo `yaml:"auth-info"`
	Contexts       map[string]*Context  `yaml:"contexts"`
	CurrentContext string               `yaml:"current-context"`
}

type AuthInfo struct {
	LocationOfOrigin           string
	TLSTrustCertsFilePath      string `yaml:"tls_trust_certs_file_path"`
	TLSAllowInsecureConnection bool   `yaml:"tls_allow_insecure_connection"`
	Token                      string `yaml:"token"`
	TokenFile                  string `yaml:"tokenFile"`

	// OAuth2 configuration
	IssuerEndpoint string `yaml:"issuer_endpoint"`
	ClientID       string `yaml:"client_id"`
	Audience       string `yaml:"audience"`
	KeyFile        string `yaml:"key_file"`
}

type Context struct {
	BrokerServiceURL string `yaml:"admin-service-url"`
	BookieServiceURL string `yaml:"bookie-service-url"`
}

type ConfigOverrides struct {
	AuthInfo       AuthInfo
	Context        Context
	CurrentContext string
	Timeout        string
}

func NewConfig() *Config {
	return &Config{
		AuthInfos: make(map[string]*AuthInfo),
		Contexts:  make(map[string]*Context),
	}
}
