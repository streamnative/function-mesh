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

type FileType string

const (
	journal  FileType = "journal"
	entryLog FileType = "entrylog"
	index    FileType = "index"
)

func ParseFileType(fileType string) (FileType, error) {
	switch strings.ToLower(fileType) {
	case journal.String():
		return journal, nil
	case entryLog.String():
		return entryLog, nil
	case index.String():
		return index, nil
	default:
		return "", errors.Errorf("invalid file type %s, the file type only can be specified as 'journal', "+
			"'entrylog', 'index'", fileType)
	}
}

func (t FileType) String() string {
	return string(t)
}
