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

package app

import (
	"flag"
	"github.com/codeallergy/glue"
	"github.com/codeallergy/sprint"
	"reflect"
)

type implFlagSetFactory struct {
	Registrars []sprint.FlagSetRegistrar `inject`
}

func FlagSetFactory() glue.FactoryBean {
	return &implFlagSetFactory{}
}

func (t *implFlagSetFactory) Object() (interface{}, error) {
	fs := flag.NewFlagSet("sprint", flag.ContinueOnError)
	for _, reg := range t.Registrars {
		reg.RegisterFlags(fs)
	}
	return fs, nil
}

func (t *implFlagSetFactory) ObjectType() reflect.Type {
	return sprint.FlagSetClass
}

func (t *implFlagSetFactory) ObjectName() string {
	return ""
}

func (t *implFlagSetFactory) Singleton() bool {
	return true
}
