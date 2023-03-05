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

package util

import (
	"crypto/tls"
)

func ParseClientAuth(s string) tls.ClientAuthType {
	switch s {
	case "no_client_cert":
		return tls.NoClientCert
	case "request_client_cert":
		return tls.RequestClientCert
	case "require_any_client_cert":
		return tls.RequireAnyClientCert
	case "verify_client_cert":
		return tls.VerifyClientCertIfGiven
	case "require_verify_client_cert":
		return tls.RequireAndVerifyClientCert
	default:
		return tls.NoClientCert
	}
}
