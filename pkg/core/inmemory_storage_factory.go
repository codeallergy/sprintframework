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
	"github.com/pkg/errors"
	"github.com/codeallergy/glue"
	"github.com/codeallergy/cachestore"
	"reflect"
)

type implInmemoryStorageFactory struct {
	beanName        string
}

func InmemoryStorageFactory(beanName string) glue.FactoryBean {
	return &implInmemoryStorageFactory{beanName: beanName}
}

func (t *implInmemoryStorageFactory) Object() (object interface{}, err error) {

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

	return cachestore.New(t.beanName), nil
}

func (t *implInmemoryStorageFactory) ObjectType() reflect.Type {
	return cachestore.ObjectType()
}

func (t *implInmemoryStorageFactory) ObjectName() string {
	return t.beanName
}

func (t *implInmemoryStorageFactory) Singleton() bool {
	return true
}

