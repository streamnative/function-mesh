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

type BookieSocketAddress struct {
	Port     int    `json:"port"`
	HostName string `json:"hostname"`
}

type LedgerMetadata struct {
	StoreCtime            bool                            `json:"storeCtime"`
	HasPassword           bool                            `json:"hasPassword"`
	MetadataFormatVersion int                             `json:"metadataFormatVersion"`
	Ensemble              int                             `json:"ensembleSize"`
	WriteQuorum           int                             `json:"writeQuorumSize"`
	AckQuorum             int                             `json:"ackQuorumSize"`
	Length                int64                           `json:"length"`
	LastEntryID           int64                           `json:"lastEntryId"`
	Ctime                 int64                           `json:"ctime"`
	CToken                int64                           `json:"cToken"`
	State                 string                          `json:"state"`
	DigestType            string                          `json:"digestType"`
	Ensembles             map[int64][]BookieSocketAddress `json:"allEnsembles"`
	CurrentEnsemble       []BookieSocketAddress           `json:"currentEnsemble"`
	Password              []byte                          `json:"password"`
	CustomMetadata        map[string][]byte               `json:"customMetadata"`
}
