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
	"github.com/streamnative/pulsarctl/pkg/bookkeeper/bkdata"
)

type Bookie interface {
	// LastLogMark gets the last log marker of a bookie
	LastLogMark() (map[string]string, error)

	// ListDiskFile gets all the files on disk of a bookie
	ListDiskFile(bkdata.FileType) (map[string]string, error)

	// ExpandStorage expands storage for a bookie
	ExpandStorage() error

	// GC triggers garbage collection for a bookie
	GC() error

	// GCStatus gets the garbage collection status
	GCStatus() (map[string]string, error)

	// GCDetails gets the garbage collection details
	GCDetails() ([]bkdata.GCStatus, error)

	// State gets the state of a bookie
	State() (*bkdata.State, error)
}

type bookie struct {
	bk       *bookieClient
	basePath string
	params   map[string]string
}

func (c *bookieClient) Bookie() Bookie {
	return &bookie{
		bk:       c,
		basePath: "/bookie",
		params:   make(map[string]string),
	}
}

func (b *bookie) LastLogMark() (map[string]string, error) {
	endpoint := b.bk.endpoint(b.basePath, "/last_log_mark")
	marker := make(map[string]string)
	err := b.bk.Client.Get(endpoint, &marker)
	return marker, err
}

func (b *bookie) ListDiskFile(fileType bkdata.FileType) (map[string]string, error) {
	endpoint := b.bk.endpoint(b.basePath, "/list_disk_file")
	b.params["file_type"] = fileType.String()
	files := make(map[string]string)
	_, err := b.bk.Client.GetWithQueryParams(endpoint, &files, b.params, true)
	return files, err
}

func (b *bookie) ExpandStorage() error {
	endpoint := b.bk.endpoint(b.basePath, "/expand_storage")
	return b.bk.Client.Put(endpoint, nil)
}

func (b *bookie) GC() error {
	endpoint := b.bk.endpoint(b.basePath, "/gc")
	return b.bk.Client.Put(endpoint, nil)
}

func (b *bookie) GCStatus() (map[string]string, error) {
	endpoint := b.bk.endpoint(b.basePath, "/gc")
	status := make(map[string]string)
	err := b.bk.Client.Get(endpoint, &status)
	return status, err
}

func (b *bookie) GCDetails() ([]bkdata.GCStatus, error) {
	endpoint := b.bk.endpoint(b.basePath, "/gc_details")
	details := make([]bkdata.GCStatus, 0)
	err := b.bk.Client.Get(endpoint, &details)
	return details, err
}

func (b *bookie) State() (*bkdata.State, error) {
	endpoint := b.bk.endpoint(b.basePath, "/state")
	var state bkdata.State
	err := b.bk.Client.Get(endpoint, &state)
	return &state, err
}
