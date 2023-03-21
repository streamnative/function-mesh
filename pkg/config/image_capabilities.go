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

// Package config defines config related tools
package config

import (
	"regexp"
)

type ImageCapabilities struct {
	// +optional
	InferTypeClassName *ImageCapability `json:"inferTypeClassName,omitempty"`
}

type ImageCapability struct {
	// +optional
	ImagePatterns []string `json:"imagePatterns,omitempty"`
}

func DefaultImageCapabilities() ImageCapabilities {
	return ImageCapabilities{
		InferTypeClassName: &ImageCapability{
			ImagePatterns: []string{
				// all 2.9 versions starting from 2.9.3.16
				`streamnative/.*?:2\.9\.(3\.(16|[2-9][0-9])|[4-9]\..*)`,
				// all 2.10 versions starting from 2.10.2.1
				`streamnative/.*?:2\.10\.[2-9]\..*`,
				// all 2.x versions starting from 2.11
				`streamnative/.*?:2\.1[1-9]\..*`,
				// all versions starting from 3.0
				`streamnative/.*?:([3-9]|[1-9][0-9]+)\..*`,
			},
		},
	}
}

func (c *ImageCapability) MatchImage(image string) bool {
	if c == nil {
		return false
	}
	for _, pattern := range c.ImagePatterns {
		if m, err := regexp.Match(pattern, []byte(image)); err != nil {
			continue
		} else if m {
			return true
		}
	}
	return false
}
