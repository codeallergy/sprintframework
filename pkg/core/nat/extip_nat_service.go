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

package nat

import (
	"github.com/pkg/errors"
	"github.com/codeallergy/sprint"
	"net"
	"time"
)

type implExternalIPService struct {
	ip  net.IP
}

func ExternalIPService(address string) (sprint.NatService, error) {
	ip := net.ParseIP(address)
	if ip == nil {
		return nil, errors.Errorf("invalid IP address '%s'", address)
	}
	return &implExternalIPService{ip: ip}, nil
}

func (t *implExternalIPService) ServiceName() string {
	return "ext_ip"
}

func (t *implExternalIPService) AllowMapping() bool {
	return false
}

func (t *implExternalIPService) AddMapping(protocol string, extport, intport int, name string, lifetime time.Duration) error {
	return nil
}

func (t *implExternalIPService) DeleteMapping(protocol string, extport, intport int) error {
	return nil
}

func (t *implExternalIPService) ExternalIP() (net.IP, error) {
	return t.ip, nil
}
