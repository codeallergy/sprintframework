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
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"github.com/codeallergy/glue"
	"github.com/codeallergy/sprint"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"reflect"
	"github.com/pkg/errors"
)

type implGrpcServerFactory struct {
	Log                     *zap.Logger                   `inject`
	AuthorizationMiddleware sprint.AuthorizationMiddleware `inject`

	beanName  string
}

func GrpcServerFactory(beanName string) glue.FactoryBean {
	return &implGrpcServerFactory{beanName: beanName}
}

func (t *implGrpcServerFactory) Object() (object interface{}, err error) {

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

	srv, err := t.createServer()
	if err != nil {
		return nil, err
	}

	return srv, nil
}

func (t *implGrpcServerFactory) ObjectType() reflect.Type {
	return sprint.GrpcServerClass
}

func (t *implGrpcServerFactory) ObjectName() string {
	return t.beanName
}

func (t *implGrpcServerFactory) Singleton() bool {
	return true
}

func (t *implGrpcServerFactory) createServer() (*grpc.Server, error) {

	var opts []grpc.ServerOption

	opts = append(opts, grpc.StreamInterceptor(grpc_auth.StreamServerInterceptor(t.AuthorizationMiddleware.Authenticate)))
	opts = append(opts, grpc.UnaryInterceptor(grpc_auth.UnaryServerInterceptor(t.AuthorizationMiddleware.Authenticate)))

	return grpc.NewServer(opts...), nil
}
