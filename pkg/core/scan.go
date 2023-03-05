/*
 * Copyright (c) 2022-2023 Zander Schwid & Co. LLC.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 */

package core

import (
	"github.com/codeallergy/sprint"
	"github.com/codeallergy/sprintframework/pkg/core/dns"
	"github.com/codeallergy/sprintframework/pkg/core/nat"
	"github.com/codeallergy/sealmod"
)

type coreScanner struct {
	Scan     []interface{}
}

func CoreScanner(scan... interface{}) sprint.CoreScanner {
	return &coreScanner {
		Scan: scan,
	}
}

func (t *coreScanner) CoreBeans() []interface{} {

	beans := []interface{}{
		LogFactory(),
		NodeService(),
		ConfigRepository(10000),
		JobService(),
		StorageService(),
		WhoisService(),
		dns.DNSProviderScanner(),
		sealmod.SealService(),
		CertificateIssueService(),
		CertificateRepository(),
		CertificateService(),
		CertificateManager(),
		nat.NatServiceFactory(),
		DynDNSService(),
		MailService(),
		&struct {
			ClientScanners []sprint.ClientScanner `inject`
			ServerScanners []sprint.ServerScanner `inject`
		}{},
	}

	return append(beans, t.Scan...)
}

