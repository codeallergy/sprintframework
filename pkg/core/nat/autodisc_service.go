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
	"sync"
	"time"
)

type implAutodiscService struct {
	what string // type of interface being auto discovered
	once sync.Once
	doit func() sprint.NatService

	mu    sync.Mutex
	found sprint.NatService
}

func startAutoDiscovery(what string, doit func() sprint.NatService) sprint.NatService {
	return &implAutodiscService{what: what, doit: doit}
}

func (t *implAutodiscService) AllowMapping() bool {
	if err := t.wait(); err != nil {
		return false
	}
	return t.found.AllowMapping()
}

func (t *implAutodiscService) AddMapping(protocol string, extport, intport int, name string, lifetime time.Duration) error {
	if err := t.wait(); err != nil {
		return err
	}
	return t.found.AddMapping(protocol, extport, intport, name, lifetime)
}

func (t *implAutodiscService) DeleteMapping(protocol string, extport, intport int) error {
	if err := t.wait(); err != nil {
		return err
	}
	return t.found.DeleteMapping(protocol, extport, intport)
}

func (t *implAutodiscService) ExternalIP() (net.IP, error) {
	if err := t.wait(); err != nil {
		return nil, err
	}
	return t.found.ExternalIP()
}

func (t *implAutodiscService) ServiceName() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.found == nil {
		return t.what
	}
	return t.found.ServiceName()
}

// wait blocks until auto-discovery has been performed.
func (t *implAutodiscService) wait() error {
	t.once.Do(func() {
		t.mu.Lock()
		t.found = t.doit()
		t.mu.Unlock()
	})
	if t.found == nil {
		return errors.Errorf("no %s router discovered", t.what)
	}
	return nil
}
