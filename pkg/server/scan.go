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
	"github.com/codeallergy/sprint"
	"net/http"
	"google.golang.org/grpc"
)

type serverScanner struct {
	Scan         []interface{}
}

func ServerScanner(scan... interface{}) sprint.ServerScanner {
	return &serverScanner{
		Scan: scan,
	}
}

func (t *serverScanner) ServerBeans() []interface{} {
	beans := []interface{}{
		&struct {
			Servers     []sprint.Server `inject:"optional"`
			GrpcServers []*grpc.Server `inject:"optional"`
			HttpServers []*http.Server `inject:"optional"`
		}{},
	}
	return append(beans, t.Scan...)
}

type grpcServerScanner struct {
	beanName    string
	Scan         []interface{}
}

func GrpcServerScanner(beanName string, scan... interface{}) sprint.ServerScanner {
	return &grpcServerScanner{
		beanName: beanName,
		Scan: scan,
	}
}

func (t *grpcServerScanner) ServerBeans() []interface{} {
	beans := []interface{}{
		AuthorizationMiddleware(),
		GrpcServerFactory(t.beanName),
		&struct {
			Servers     []sprint.Server `inject:"optional"`
			GrpcServers []*grpc.Server `inject:"optional"`
			HttpServers []*http.Server `inject:"optional"`
		}{},
	}
	return append(beans, t.Scan...)
}

type httpServerScanner struct {
	beanName    string
	Scan         []interface{}
}

func HttpServerScanner(beanName string, scan... interface{}) sprint.ServerScanner {
	return &httpServerScanner{
		beanName: beanName,
		Scan: scan,
	}
}

func (t *httpServerScanner) ServerBeans() []interface{} {
	beans := []interface{}{
		HttpServerFactory(t.beanName),
		&struct {
			HttpServers []*http.Server `inject`
		}{},
	}
	return append(beans, t.Scan...)
}


