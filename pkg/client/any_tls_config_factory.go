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

package client

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"github.com/codeallergy/glue"
	"github.com/codeallergy/sprint"
	"reflect"
	"github.com/pkg/errors"
)

type implAnyTlsConfigFactory struct {
	Properties    glue.Properties  `inject`
	beanName string
}

func AnyTlsConfigFactory(beanName string) glue.FactoryBean {
	return &implAnyTlsConfigFactory{beanName: beanName}
}

func (t *implAnyTlsConfigFactory) Object() (object interface{}, err error) {

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

	insecure := t.Properties.GetBool(fmt.Sprintf("%s.insecure", t.beanName), false)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecure,
		Rand:               rand.Reader,
	}

	tlsConfig.NextProtos = appendH2ToNextProtos(tlsConfig.NextProtos)
	return tlsConfig, nil
}

func (t *implAnyTlsConfigFactory) ObjectType() reflect.Type {
	return sprint.TlsConfigClass
}

func (t *implAnyTlsConfigFactory) ObjectName() string {
	return t.beanName
}

func (t *implAnyTlsConfigFactory) Singleton() bool {
	return true
}
