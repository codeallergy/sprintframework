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
	rt "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"testing"
)

func TestHttpMux(t *testing.T) {

	mux := http.NewServeMux()
	api := rt.NewServeMux()
	mux.Handle("/api/", api)

	u, err := url.Parse("http://localhost:8443/api/")
	require.NoError(t, err)

	req := &http.Request{
		Method:     "GET",
		URL:        u,
		Host:       "localhost",
		RequestURI: "/api/",
	}

	handler, foundPattern := mux.Handler(req)
	require.Equal(t, "/api/", foundPattern)
	require.Equal(t, handler, api)

}

func TestHttpMuxRewrite(t *testing.T) {
	u, err := url.Parse("http://localhost:8443/")
	require.NoError(t, err)
	u.Path = "/index.html"
	require.Equal(t, u.RequestURI(), "/index.html")
}
