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
	"github.com/pkg/errors"
	"github.com/codeallergy/sprint"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"net"
	"net/http"
	"strings"
)

type implHttpServer struct {

	Log       *zap.Logger              `inject`

	srv       *http.Server
	listener  net.Listener

	running   atomic.Bool
}

func NewHttpServer(srv *http.Server) sprint.Server {
	return &implHttpServer{srv: srv}
}

func (t *implHttpServer) PostConstruct() error {
	t.running.Store(false)
	return nil
}

func (t *implHttpServer) Bind() (err error) {

	t.listener, err = net.Listen("tcp4", t.srv.Addr)
	if err != nil {
		return errors.Errorf("can not bind to port '%s', %v", t.srv.Addr, err)
	}

	return nil
}

func (t *implHttpServer) Active() bool {
	return t.running.Load()
}

func (t *implHttpServer) ListenAddress() net.Addr {
	if t.listener != nil {
		return t.listener.Addr()
	} else {
		return EmptyAddr{}
	}
}

func (t *implHttpServer) Stop() {
	if t.running.CAS(true, false) {
		if t.listener != nil {
			t.listener.Close()
		}
		t.srv.Close()
	}
}

func (t *implHttpServer) Destroy() error {
	t.Stop()
	return nil
}

func (t *implHttpServer) Serve() (err error) {

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

	t.Log.Info("HttpServerServe", zap.String("addr", t.srv.Addr), zap.Bool("tls", t.srv.TLSConfig != nil))

	t.running.Store(true)
	if t.srv.TLSConfig != nil {
		err = t.srv.ServeTLS(t.listener, "", "")
	} else {
		err = t.srv.Serve(t.listener)
	}

	t.running.Store(false)
	if err != nil && strings.Contains(err.Error(), "closed") {
		return nil
	}
	return err
}
