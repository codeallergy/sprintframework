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

package util

import (
	rt "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
)

func FindGatewayHandler(srv *http.Server, pattern string) (*rt.ServeMux, error) {
	handler := srv.Handler

	switch mux := handler.(type) {
	case *rt.ServeMux:
		return mux, nil
	case *http.ServeMux:
		return findGatewayAPIHandler(mux, pattern)
	default:
		return nil, errors.Errorf("unknown server handler '%v'", handler)
	}
}


func findGatewayAPIHandler(mux *http.ServeMux, pattern string) (*rt.ServeMux, error) {

	u, err := url.Parse("http://localhost:/api/")
	if err != nil {
		return nil, errors.Errorf("parsing configuration URL error, %v", err)
	}
	req := &http.Request{
		Method:           "GET",
		URL:              u,
		Host:             "localhost",
		RequestURI:       pattern,
	}

	handler, foundPattern := mux.Handler(req)
	if foundPattern != pattern {
		return nil, errors.Errorf("invalid configuration of http mux, found pattern '%s' whereas expected '%s'", foundPattern, pattern)
	}

	if handler == nil {
		return nil, errors.Errorf("handler not found for pattern '%s'", pattern)
	}

	rtMux, ok := handler.(*rt.ServeMux)
	if !ok {
		return nil, errors.Errorf("non gateway mux '%v' found on pattern '%s'", handler, pattern)
	}

	return rtMux, nil
}
