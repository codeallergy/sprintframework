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
package dns

import (
	"github.com/codeallergy/glue"
	"github.com/codeallergy/sprintframework/pkg/core/dns/netlify"
	"github.com/codeallergy/sprint"
)

type dnsProviderScanner struct {
	Scan     []interface{}
}

func DNSProviderScanner(scan... interface{}) glue.Scanner {
	return &dnsProviderScanner{
		Scan: scan,
	}
}

func (t *dnsProviderScanner) Beans() []interface{} {

	beans := []interface{}{
		netlify.NetlifyProvider(),
		&struct {
			DNSProviders []sprint.DNSProvider `inject`
		}{},
	}

	return append(beans, t.Scan...)
}

