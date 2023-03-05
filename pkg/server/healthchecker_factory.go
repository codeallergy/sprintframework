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
	"github.com/codeallergy/glue"
	"github.com/codeallergy/sprint"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"reflect"
	"github.com/pkg/errors"
)

type implHealthcheckerFactory struct {
	glue.FactoryBean
	GrpcServer    *grpc.Server         `inject`

	enableServices  bool
}

func HealthcheckerFactory(enableServices bool) glue.FactoryBean {
	return &implHealthcheckerFactory{enableServices: enableServices}
}

func (t *implHealthcheckerFactory) Object() (object interface{}, err error) {

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

	srv := health.NewServer()

	srv.SetServingStatus(
		"",
		grpc_health_v1.HealthCheckResponse_SERVING,
	)

	grpc_health_v1.RegisterHealthServer(t.GrpcServer, srv)

	if t.enableServices {
		for serviceName := range t.GrpcServer.GetServiceInfo() {
			srv.SetServingStatus(
				serviceName,
				grpc_health_v1.HealthCheckResponse_SERVING,
			)
		}
	}

	return srv, nil
}

func (t *implHealthcheckerFactory) ObjectType() reflect.Type {
	return sprint.HealthCheckerClass
}

func (t *implHealthcheckerFactory) ObjectName() string {
	return ""
}

func (t *implHealthcheckerFactory) Singleton() bool {
	return true
}
