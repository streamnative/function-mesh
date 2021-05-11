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

package bookkeeper

import (
	"fmt"
	"net/http"
	"path"

	"github.com/streamnative/pulsarctl/pkg/bookkeeper/bkdata"
	"github.com/streamnative/pulsarctl/pkg/cli"
)

type Client interface {
	// Bookie related commands
	Bookie() Bookie
	// Bookies related commands
	Bookies() Bookies
	// Ledger related commands
	Ledger() Ledger
	// AutoRecovery related commands
	AutoRecovery() AutoRecovery
}

type bookieClient struct {
	Client     *cli.Client
	APIVersion bkdata.APIVersion
}

func New(config *Config) (Client, error) {
	if len(config.WebServiceURL) == 0 {
		config.WebServiceURL = DefaultWebServiceURL
	}

	bkClient := &bookieClient{
		APIVersion: config.APIVersion,
		Client: &cli.Client{
			ServiceURL:  config.WebServiceURL,
			VersionInfo: ReleaseVersion,
			HTTPClient: &http.Client{
				Timeout: config.HTTPTimeout,
			},
		},
	}

	return bkClient, nil
}

func (c *bookieClient) endpoint(componentPath string, parts ...string) string {
	return path.Join(makeHTTPPath(c.APIVersion.String(), componentPath), path.Join(parts...))
}

func makeHTTPPath(apiVersion string, componentPath string) string {
	return fmt.Sprintf("/api/%s%s", apiVersion, componentPath)
}
