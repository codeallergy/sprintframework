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
	"github.com/codeallergy/sprint"
)

type clientScanner struct {
	Scan         []interface{}
}

func ClientScanner(scan... interface{}) sprint.ClientScanner {
	return &clientScanner{
		Scan: scan,
	}
}

func (t *clientScanner) ClientBeans() []interface{} {
	beans := []interface{}{
		&struct {
			ControlClient []sprint.ControlClient `inject`
		}{},
	}
	return append(beans, t.Scan...)
}

type controlClientScanner struct {
	Scan         []interface{}
}

func ControlClientScanner(scan... interface{}) sprint.ClientScanner {
	return &controlClientScanner{
		Scan: scan,
	}
}

func (t *controlClientScanner) ClientBeans() []interface{} {
	beans := []interface{}{
		GrpcClientFactory("control-grpc-client"),
		ControlClient(),
		&struct {
			ControlClient []sprint.ControlClient `inject`
		}{},
	}
	return append(beans, t.Scan...)
}

