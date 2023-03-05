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
	"github.com/pkg/errors"
	"github.com/codeallergy/glue"
	"github.com/codeallergy/properties"
	"github.com/codeallergy/sprint"
	"path/filepath"
	"reflect"
)

var (
	CertFile = "client.crt"
	KeyFile  = "client.key"
)

type tlsConfigFactory struct {
	Application sprint.Application `inject`
	Properties  glue.Properties `inject`

	CompanyName   string        `value:"application.company,default=sprint"`

	beanName string
}

func TlsConfigFactory(beanName string) glue.FactoryBean {
	return &tlsConfigFactory{beanName: beanName}
}

func (t *tlsConfigFactory) Object() (object interface{}, err error) {

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

	appDir := properties.Locate(t.CompanyName).GetDir(t.Application.Name())

	certFile := filepath.Join(appDir, CertFile)
	keyFile := filepath.Join(appDir, KeyFile)

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, errors.Errorf("LoadX509KeyPair for implControlClient SSL from %s and %s failed, %v", certFile, keyFile, err)
	}

	insecure := t.Properties.GetBool(fmt.Sprintf("%s.insecure", t.beanName), false)

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: insecure,
		Rand:               rand.Reader,
	}

	tlsConfig.NextProtos = appendH2ToNextProtos(tlsConfig.NextProtos)
	return tlsConfig, err
}

func (t *tlsConfigFactory) ObjectType() reflect.Type {
	return sprint.TlsConfigClass
}

func (t *tlsConfigFactory) ObjectName() string {
	return t.beanName
}

func (t *tlsConfigFactory) Singleton() bool {
	return true
}

