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

package bkdata

import (
	"strings"

	"github.com/pkg/errors"
)

type BookieType string

const (
	rw BookieType = "rw"
	ro BookieType = "ro"
)

func ParseBookieType(t string) (BookieType, error) {
	switch strings.ToLower(t) {
	case rw.String():
		return rw, nil
	case ro.String():
		return ro, nil
	default:
		return "", errors.Errorf("invalid bookie type %s, the bookie type only can "+
			"be specified as 'rw' or 'ro'", t)
	}
}

func (t BookieType) String() string {
	return string(t)
}
