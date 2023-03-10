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
	"crypto/tls"
	"fmt"
	"github.com/pkg/errors"
	"github.com/codeallergy/glue"
	"github.com/codeallergy/sprint"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"reflect"
	"strings"
)

type implGrpcClientFactory struct {

	Application sprint.Application `inject`
	Properties  glue.Properties `inject`
	TlsConfig   *tls.Config       `inject:"optional"`

	beanName string
}

func GrpcClientFactory(beanName string) glue.FactoryBean {
	return &implGrpcClientFactory{
		beanName: beanName,
	}
}

func (t *implGrpcClientFactory) Object() (object interface{}, err error) {

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

	// try to get normal property
	connectAddr := t.Properties.GetString(fmt.Sprintf("%s.connect-address", t.beanName), "")
	if connectAddr == "" {
		// try to convert from server address
		grpcListenAddr := t.Properties.GetString( "control-grpc-server.listen-address", "")
		if grpcListenAddr == "" {
			return nil, errors.Errorf("property '%s.connect-address' is not found and property 'control-grpc-server.listen-address' is not found too'", t.beanName)
		}
		connectAddr = t.getConnectAddress(grpcListenAddr)
	}

	return t.doDial(connectAddr)
}

func (t *implGrpcClientFactory) ObjectType() reflect.Type {
	return sprint.GrpcClientClass
}

func (t *implGrpcClientFactory) ObjectName() string {
	return t.beanName
}

func (t *implGrpcClientFactory) Singleton() bool {
	return true
}

func (t *implGrpcClientFactory) getConnectAddress(listenAddr string) string {
	if strings.HasPrefix(listenAddr, "0.0.0.0:") {
		return "127.0.0.1" + listenAddr[7:]
	}
	if strings.HasPrefix(listenAddr, ":") {
		return "127.0.0.1" + listenAddr
	}
	return listenAddr
}

func (t *implGrpcClientFactory) doDial(connectAddr string) (*grpc.ClientConn, error) {

	var opts []grpc.DialOption

	if t.TlsConfig != nil {
		tlsCredentials := credentials.NewTLS(t.TlsConfig)
		opts = append(opts, grpc.WithTransportCredentials(tlsCredentials))
	} else {
		opts = append(opts, grpc.WithInsecure())
	}

	maxMessageSize := t.Properties.GetInt(fmt.Sprintf("%s.max.message.size", t.beanName), 0)
	if maxMessageSize != 0 {
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(maxMessageSize)))
	}

	authToken := t.Properties.GetString("application.auth", "")
	if authToken != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(&tokenAuth{token: authToken}))
	}

	return grpc.Dial(connectAddr, opts...)
}
