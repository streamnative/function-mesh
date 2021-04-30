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

import "github.com/streamnative/pulsarctl/pkg/bookkeeper/bkdata"

type AutoRecovery interface {
	// RecoverBookie is used to recovering ledger data for a failed bookie
	RecoverBookie([]string, bool) error

	// ListUnderReplicatedLedger is used to listing all the underreplicated ledgers
	// which have been marked for rereplication
	ListUnderReplicatedLedger(string, string) ([]int64, error)

	// PrintListUnderReplicatedLedger is used to printing the replicate list of the replicated ledgers
	PrintListUnderReplicatedLedger(string, string) (map[int64][]string, error)

	// WhoIsAuditor is used to getting which bookie is the auditor
	WhoIsAuditor() (string, error)

	// TriggerAudit is used to triggering audit by resetting the lostBookieRecoveryDelay
	TriggerAudit() error

	// GetLostBookieRecoveryDelay is used to getting the lostBookieRecoveryDelay of a bookie
	GetLostBookieRecoveryDelay() (string, error)

	// SetLostBookieRecoveryDelay is used to setting the lastBookieRecoverDelay of a bookie
	SetLostBookieRecoveryDelay(int) error

	// Decommission is used to decommissioning a bookie
	Decommission(string) error
}

type autoRecovery struct {
	bk       *bookieClient
	basePath string
	params   map[string]string
}

func (c *bookieClient) AutoRecovery() AutoRecovery {
	return &autoRecovery{
		bk:       c,
		basePath: "/autorecovery",
		params:   make(map[string]string),
	}
}

func (a *autoRecovery) RecoverBookie(src []string, deleteCookie bool) error {
	endpoint := a.bk.endpoint(a.basePath, "/bookie")
	request := bkdata.RecoveryRequest{
		BookieSrc:    src,
		DeleteCookie: deleteCookie,
	}
	return a.bk.Client.Put(endpoint, &request)
}

func (a *autoRecovery) ListUnderReplicatedLedger(missingReplica, excludingMissingReplica string) ([]int64, error) {
	endpoint := a.bk.endpoint(a.basePath, "/list_under_replicated_ledger")
	a.params["missingreplica"] = missingReplica
	a.params["excludingmissingreplica"] = excludingMissingReplica
	resp := make([]int64, 0)
	_, err := a.bk.Client.GetWithQueryParams(endpoint, &resp, a.params, true)
	return resp, err
}

func (a *autoRecovery) PrintListUnderReplicatedLedger(missingReplica,
	excludingMissingReplica string) (map[int64][]string, error) {

	endpoint := a.bk.endpoint(a.basePath, "/list_under_replicated_ledger")
	a.params["missingreplica"] = missingReplica
	a.params["excludingmissingreplica"] = excludingMissingReplica
	a.params["printmissingreplica"] = "true"
	resp := make(map[int64][]string)
	_, err := a.bk.Client.GetWithQueryParams(endpoint, &resp, a.params, true)
	return resp, err
}

func (a *autoRecovery) WhoIsAuditor() (string, error) {
	endpoint := a.bk.endpoint(a.basePath, "/who_is_auditor")
	resp, err := a.bk.Client.GetWithQueryParams(endpoint, nil, nil, false)
	return string(resp), err
}

func (a *autoRecovery) TriggerAudit() error {
	endpoint := a.bk.endpoint(a.basePath, "/trigger_audit")
	return a.bk.Client.Put(endpoint, nil)
}

func (a *autoRecovery) GetLostBookieRecoveryDelay() (string, error) {
	endpoint := a.bk.endpoint(a.basePath, "/lost_bookie_recovery_delay")
	resp, err := a.bk.Client.GetWithQueryParams(endpoint, nil, nil, false)
	return string(resp), err
}

func (a *autoRecovery) SetLostBookieRecoveryDelay(delay int) error {
	endpoint := a.bk.endpoint(a.basePath, "/lost_bookie_recovery_delay")
	req := bkdata.LostBookieRecoverDelayRequest{
		DelaySeconds: delay,
	}
	return a.bk.Client.Put(endpoint, &req)
}

func (a *autoRecovery) Decommission(src string) error {
	endpoint := a.bk.endpoint(a.basePath, "/decommission")
	req := bkdata.DecommissionRequest{
		BookieSrc: src,
	}
	return a.bk.Client.Put(endpoint, &req)
}
