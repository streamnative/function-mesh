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
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMatchImage(t *testing.T) {
	imageCapabilities := DefaultImageCapabilities()

	assert.Equal(t, false, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.8.1"))
	assert.Equal(t, false, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.8.2"))
	assert.Equal(t, false, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.8.3.19"))

	assert.Equal(t, false, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.9.1"))
	assert.Equal(t, false, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.9.2"))
	assert.Equal(t, false, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.9.3.13.1"))
	assert.Equal(t, false, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.9.3.13"))
	assert.Equal(t, false, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.9.3.14"))
	assert.Equal(t, false, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.9.3.15"))
	assert.Equal(t, true, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.9.3.16"))
	assert.Equal(t, true, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.9.3.16.2"))
	assert.Equal(t, true, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.9.3.16.2.test"))
	assert.Equal(t, true, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.9.3.20"))
	assert.Equal(t, true, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.9.3.30"))
	assert.Equal(t, true, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.9.4.30"))
	assert.Equal(t, true, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.9.4.0"))
	assert.Equal(t, true, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.9.4.1"))
	assert.Equal(t, true, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.9.4.190"))
	assert.Equal(t, false, imageCapabilities.InferTypeClassName.MatchImage("custom/pulsar:2.9.4.190"))

	assert.Equal(t, false, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.10.1"))
	assert.Equal(t, false, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.10.1.19"))
	assert.Equal(t, true, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.10.2.1"))
	assert.Equal(t, true, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.10.3.1"))
	assert.Equal(t, true, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.10.4.2"))
	assert.Equal(t, true, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.10.4.20"))
	assert.Equal(t, false, imageCapabilities.InferTypeClassName.MatchImage("custom/pulsar:2.10.4.20"))

	assert.Equal(t, true, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.11.0.1"))
	assert.Equal(t, true, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:2.11.1"))
	assert.Equal(t, false, imageCapabilities.InferTypeClassName.MatchImage("xxxxx/pulsar:2.11.1"))

	assert.Equal(t, true, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:3.1.1"))
	assert.Equal(t, true, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:4.1.0"))
	assert.Equal(t, true, imageCapabilities.InferTypeClassName.MatchImage("streamnative/pulsar:5.2.1"))
	assert.Equal(t, false, imageCapabilities.InferTypeClassName.MatchImage("yyyyyy/pulsar:5.2.1"))
}
