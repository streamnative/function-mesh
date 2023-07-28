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
)

type Bookies interface {
	// List lists all the available bookies
	List(bkdata.BookieType, bool) (map[string]string, error)

	// DiskUsageInfo gets the bookies disk usage info of a cluster
	DiskUsageInfo() (map[string]string, error)
}

type bookies struct {
	bk       *bookieClient
	basePath string
	params   map[string]string
}

func (c *bookieClient) Bookies() Bookies {
	return &bookies{
		bk:       c,
		basePath: "/bookie",
		params:   make(map[string]string),
	}
}

func (b *bookies) List(t bkdata.BookieType, show bool) (map[string]string, error) {
	endpoint := b.bk.endpoint(b.basePath, "/list_bookies")
	b.params["type"] = t.String()
	b.params["print_hostnames"] = strconv.FormatBool(show)
	bookies := make(map[string]string)
	_, err := b.bk.Client.GetWithQueryParams(endpoint, &bookies, b.params, true)
	return bookies, err
}

func (b *bookies) DiskUsageInfo() (map[string]string, error) {
	endpoint := b.bk.endpoint(b.basePath, "/list_bookie_info")
	info := make(map[string]string)
	err := b.bk.Client.Get(endpoint, &info)
	return info, err
}
