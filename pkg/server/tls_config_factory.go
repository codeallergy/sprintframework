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

package server

import (
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"github.com/codeallergy/glue"
	"github.com/codeallergy/sprint"
	"reflect"
	"github.com/pkg/errors"
)

type implTlsConfigFactory struct {

	Properties  glue.Properties `inject`
	NodeService sprint.NodeService `inject`

	CertificateManager sprint.CertificateManager `inject`
	DomainService      sprint.CertificateService `inject`

	beanName          string
}

func TlsConfigFactory(beanName string) glue.FactoryBean {
	return &implTlsConfigFactory{beanName: beanName}
}

func (t *implTlsConfigFactory) Object() (obj interface{}, err error) {

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
		GetCertificate: t.CertificateManager.GetCertificate,
		Rand:         rand.Reader,
		InsecureSkipVerify: insecure,
	}

	tlsConfig.NextProtos = AppendH2ToNextProtos(tlsConfig.NextProtos)
	return tlsConfig, nil
}

func (t *implTlsConfigFactory) ObjectType() reflect.Type {
	return sprint.TlsConfigClass
}

func (t *implTlsConfigFactory) ObjectName() string {
	return t.beanName
}

func (t *implTlsConfigFactory) Singleton() bool {
	return true
}


