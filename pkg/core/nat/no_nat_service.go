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
	"errors"
	"github.com/codeallergy/sprint"
	"net"
	"time"
)

var (
	ErrNoNatService = errors.New("no nat service")
)

type implNonatService struct {
}

func NoNatService() sprint.NatService {
	return &implNonatService{}
}

func (t *implNonatService) ServiceName() string {
	return "no_nat"
}

func (t *implNonatService) AllowMapping() bool {
	return false
}

func (t *implNonatService) AddMapping(protocol string, extport, intport int, name string, lifetime time.Duration) error {
	return nil
}

func (t *implNonatService) DeleteMapping(protocol string, extport, intport int) error {
	return nil
}

func (t *implNonatService) ExternalIP() (net.IP, error) {
	return nil, ErrNoNatService
}


