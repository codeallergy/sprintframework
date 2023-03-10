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
	"fmt"
	"github.com/codeallergy/glue"
	"github.com/codeallergy/sprint"
	"github.com/codeallergy/sprintframework/pkg/util"
	"github.com/codeallergy/uuid"
	"github.com/pkg/errors"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"runtime"
	"strconv"
	"sync"
	"time"
)

const oneMb = 1024 * 1024

type implNodeService struct {
	Application      sprint.Application      `inject`
	Properties       glue.Properties      `inject`
	ConfigRepository sprint.ConfigRepository `inject`
	Log              *zap.Logger            `inject`

	initOnce sync.Once

	nodeIdHex string
	nodeId    uint64

	lastTimestamp atomic.Int64
	clock         atomic.Int32
}

func NodeService() sprint.NodeService {
	return &implNodeService{}
}

func (t *implNodeService) BeanName() string {
	return "node_service"
}

func (t *implNodeService) GetStats(cb func(name, value string) bool) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	cb("id", t.nodeIdHex)
	cb("numGoroutine", strconv.Itoa(runtime.NumGoroutine()))
	cb("numCPU", strconv.Itoa(runtime.NumCPU()))
	cb("numCgoCall", strconv.FormatInt(runtime.NumCgoCall(), 10))
	cb("goVersion", runtime.Version())
	cb("memAlloc", fmt.Sprintf("%dmb", m.Alloc / oneMb))
	cb("memTotalAlloc", fmt.Sprintf("%dmb", m.TotalAlloc / oneMb))
	cb("memSys", fmt.Sprintf("%dmb", m.Sys / oneMb))
	cb("memNumGC", strconv.Itoa(int(m.NumGC)))

	return nil
}

func (t *implNodeService) PostConstruct() (err error) {

	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v
			case string:
				err = errors.New(v)
			default:
				err = errors.Errorf("%v", v)
			}
		}
	}()

	t.nodeIdHex = t.Properties.GetString("node.id", "")
	if t.nodeIdHex == "" {
		t.nodeIdHex, err = util.GenerateNodeId()
		if err != nil {
			return errors.Errorf("generate node id, %v", err)
		}
		err = t.ConfigRepository.Set("node.id", t.nodeIdHex)
		if err != nil {
			return errors.Errorf("set property 'node.id' with value '%s', %v", t.nodeIdHex, err)
		}
	}
	t.nodeId, err = util.ParseNodeId(t.nodeIdHex)
	return err
}

func (t *implNodeService) NodeId() uint64 {
	return t.nodeId
}

func (t *implNodeService) NodeIdHex() string {
	return t.nodeIdHex
}

func (t *implNodeService) Issue() uuid.UUID {

	id := uuid.New(uuid.TimebasedVer1)
	id.SetTime(time.Now())
	id.SetNode(int64(t.nodeId))

	for {

		curr := id.UnixTime100Nanos()
		old := t.lastTimestamp.Load()
		if old == curr {
			id.SetClockSequence(int(t.clock.Inc()))
			break
		}

		if t.lastTimestamp.CAS(old, curr) {
			t.clock.Store(0)
			break
		}

		old = t.lastTimestamp.Load()
		if old > curr {
			id.SetTime(time.Now())
		}

	}

	return id

}

func (t *implNodeService) Parse(id uuid.UUID) (timestampMillis int64, nodeId int64, clock int) {
	return id.UnixTimeMillis(), id.Node(), id.ClockSequence()
}

