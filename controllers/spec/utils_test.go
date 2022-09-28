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

package spec

import (
	"testing"

	"github.com/streamnative/function-mesh/api/compute/v1alpha1"

	"github.com/stretchr/testify/assert"
)

func TestGetValFromPtrOrDefault(t *testing.T) {
	boolVal := true
	boolPtr := &boolVal
	assert.Equal(t, getBoolFromPtrOrDefault(boolPtr, false), boolVal)
	assert.Equal(t, getBoolFromPtrOrDefault(nil, boolVal), boolVal)

	var int32Val int32 = 100
	int32Ptr := &int32Val
	assert.Equal(t, getInt32FromPtrOrDefault(int32Ptr, 200), int32Val)
	assert.Equal(t, getInt32FromPtrOrDefault(nil, int32Val), int32Val)
}

func TestMarshalSecretsMap(t *testing.T) {
	secrets := map[string]v1alpha1.SecretRef{
		"foo": {
			Path: "path",
		},
	}
	marshaledSecrets := marshalSecretsMap(secrets)
	assert.Equal(t, marshaledSecrets, `{"foo":{"path":"path"}}`)

	marshaledSecretsNil := marshalSecretsMap(nil)
	assert.Equal(t, marshaledSecretsNil, `{}`)
}
