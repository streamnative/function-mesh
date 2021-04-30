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
	"strconv"

	"github.com/streamnative/pulsarctl/pkg/bookkeeper/bkdata"
	"github.com/streamnative/pulsarctl/pkg/cli"
)

type Ledger interface {
	// Delete the specified ledger
	Delete(int64) error

	// List all the ledgers and get the metadata
	List(bool) (map[int64]string, error)

	// Get the metadata of a ledger
	Get(int64) (map[int64]bkdata.LedgerMetadata, error)

	// Read a range of entries from a ledger
	Read(int64, int64, int64) (map[string]string, error)
}

type ledger struct {
	client   *bookieClient
	request  *cli.Client
	basePath string
	params   map[string]string
}

func (c *bookieClient) Ledger() Ledger {
	return &ledger{
		client:   c,
		request:  c.Client,
		basePath: "/ledger",
		params:   make(map[string]string),
	}
}

func (c *ledger) Delete(ledgerID int64) error {
	endpoint := c.client.endpoint(c.basePath, "/delete")
	c.params["ledger_id"] = strconv.FormatInt(ledgerID, 10)
	return c.request.DeleteWithQueryParams(endpoint, c.params)
}

func (c *ledger) List(showMeta bool) (map[int64]string, error) {
	endpoint := c.client.endpoint(c.basePath, "list")
	c.params["print_metadata"] = strconv.FormatBool(showMeta)
	var metadata map[int64]string
	_, err := c.request.GetWithQueryParams(endpoint, &metadata, c.params, true)
	return metadata, err
}

func (c *ledger) Get(ledgerID int64) (map[int64]bkdata.LedgerMetadata, error) {
	endpoint := c.client.endpoint(c.basePath, "metadata")
	c.params["ledger_id"] = strconv.FormatInt(ledgerID, 10)
	var metadata map[int64]bkdata.LedgerMetadata
	_, err := c.request.GetWithQueryParams(endpoint, &metadata, c.params, true)
	return metadata, err
}

func (c *ledger) Read(ledgerID int64, start int64, end int64) (map[string]string, error) {
	endpoint := c.client.endpoint(c.basePath, "read")
	c.params["ledger_id"] = strconv.FormatInt(ledgerID, 10)
	if start >= 0 {
		c.params["start_entry_id"] = strconv.FormatInt(start, 10)
	}
	if end >= 0 {
		c.params["end_entry_id"] = strconv.FormatInt(end, 10)
	}
	info := make(map[string]string)
	_, err := c.request.GetWithQueryParams(endpoint, &info, c.params, true)
	return info, err
}
