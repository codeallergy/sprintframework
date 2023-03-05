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

package cmd

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/codeallergy/glue"
	"github.com/codeallergy/sprintframework/pkg/server"
	"github.com/codeallergy/sprint"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func doWithControlClient(parent glue.Context, cb func(sprint.ControlClient) error) error {

	var verbose bool
	list := parent.Bean(sprint.ApplicationFlagsClass, glue.DefaultLevel)
	if len(list) > 0 {
		flags := list[0].Object().(sprint.ApplicationFlags)
		if flags.Verbose() {
			verbose = true
		}
	}

	list = parent.Bean(sprint.ClientScannerClass, glue.DefaultLevel)
	if len(list) != 1 {
		return errors.Errorf("application context should have one client scanner, but found '%d'", len(list))
	}
	bean := list[0]

	scanner, ok := bean.Object().(sprint.ClientScanner)
	if !ok {
		return errors.Errorf("invalid object '%v' found instead of sprint.ClientScanner in application context", bean.Class())
	}

	beans := scanner.ClientBeans()
	if verbose {
		verbose := glue.Verbose{ Log: log.Default() }
		beans = append([]interface{}{verbose}, beans...)
	}

	ctx, err := parent.Extend(beans...)
	if err != nil {
		return err
	}
	defer ctx.Close()

	list = ctx.Bean(sprint.ControlClientClass, glue.DefaultLevel)
	if len(list) != 1 {
		return errors.Errorf("client context should have one sprint.ControlClient inference, but found '%d'", len(list))
	}
	bean = list[0]

	if client, ok := bean.Object().(sprint.ControlClient); ok {
		return cb(client)
	} else {
		return errors.Errorf("invalid object '%v' found instead of sprint.ControlClient in client context", bean.Class())
	}

}

func doWithServers(core glue.Context, cb func([]sprint.Server) error) (err error) {

	var contextList []glue.Context

	defer func() {

		var listErr []error
		if r := recover(); r != nil {
			listErr = append(listErr, errors.Errorf("recovered on error: %v", r))
		}

		for _, ctx := range contextList {
			if e := ctx.Close(); e != nil {
				listErr = append(listErr, e)
			}
		}

		if len(listErr) > 0 {
			err = errors.Errorf("%v", listErr)
		}

	}()

	list := core.Bean(sprint.ServerScannerClass, glue.DefaultLevel)
	if len(list) == 0 {
		return errors.New("no one sprint.ServerScanner found in core context")
	}

	for i, s := range list {
		scanner, ok := s.Object().(sprint.ServerScanner)
		if !ok {
			return errors.Errorf("invalid object found for sprint.ServerScanner on position %d in core context", i)
		}
		ctx, err := core.Extend(scanner.ServerBeans()...)
		if err != nil {
			return errors.Errorf("server creation context %v failed by %v", s, err)
		}
		contextList = append(contextList, ctx)
	}

	var serverList []sprint.Server
	for _, ctx := range contextList {

		for i, bean := range ctx.Bean(sprint.ServerClass, glue.DefaultLevel) {
			if srv, ok := bean.Object().(sprint.Server); ok {
				serverList = append(serverList, srv)
			} else {
				return errors.Errorf("invalid object found for sprint.Server on position %d in server context", i)
			}
		}

		for i, bean := range ctx.Bean(sprint.GrpcServerClass, glue.DefaultLevel) {
			if srv, ok := bean.Object().(*grpc.Server); ok {
				s := server.NewGrpcServer(bean.Name(), srv)
				if err := ctx.Inject(s); err != nil {
					return errors.Errorf("injection error for server '%s' of *grpc.Server on position %d in server context, %v", bean.Name(), i, err)
				}
				serverList = append(serverList, s)
			} else {
				return errors.Errorf("invalid object found for *grpc.Server on position %d in server context", i)
			}
		}

		for i, bean := range ctx.Bean(sprint.HttpServerClass, glue.DefaultLevel) {
			if srv, ok := bean.Object().(*http.Server); ok {
				s := server.NewHttpServer(srv)
				if err := ctx.Inject(s); err != nil {
					return errors.Errorf("injection error for server '%s' of *http.Server on position %d in server context, %v", srv.Addr, i, err)
				}
				serverList = append(serverList, s)
			} else {
				return errors.Errorf("invalid object found for *http.Server on position %d in server context", i)
			}
		}

	}

	return cb(serverList)
}

func runServers(application sprint.Application, core glue.Context, log *zap.Logger) error {

	return doWithServers(core, func(servers []sprint.Server) error {

		defer func() {
			if r := recover(); r != nil {
				switch v := r.(type) {
				case error:
					log.Error("Recover", zap.Error(v))
				case string:
					log.Error("Recover", zap.String("error", v))
				default:
					log.Error("Recover", zap.String("error", fmt.Sprintf("%v", v)))
				}
			}
		}()

		if len(servers) == 0 {
			return errors.New("sprint.Server instances are not found in server context")
		}

		c, cancel := context.WithCancel(context.Background())
		defer cancel()

		var boundServers []sprint.Server
		for _, server := range servers {
			if err := server.Bind(); err != nil {
				log.Error("Bind", zap.Error(err))
			} else {
				boundServers = append(boundServers, server)
			}
		}

		cnt := 0
		g, _ := errgroup.WithContext(c)

		for _, server := range boundServers {
			g.Go(server.Serve)
			cnt++
		}
		log.Info("ApplicationStarted", zap.Int("Servers", cnt))

		go func() {

			signalCh := make(chan os.Signal, 10)
			signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

			var signal os.Signal

			waitAgain:
			select {
			case signal = <- signalCh:
			case <- application.Done():
				signal = syscall.SIGABRT
			}

			if signal == syscall.SIGHUP {
				list := core.Bean(sprint.LumberjackClass, 1)
				if len(list) > 0 {
					for _, bean := range list {
						if logger, ok := bean.Object().(*lumberjack.Logger); ok {
							logger.Rotate()
						}
					}
					goto waitAgain
				}
				// no lumberjack found, restart application
				application.Shutdown(true)
			}

			log.Info("StopApplication", zap.String("signal", signal.String()))
			total := 0
			for _, server := range boundServers {
				server.Stop()
				total++
			}
			log.Info("ServersStopped", zap.Int("cnt", total))
			log.Sync()
			cancel()

		}()

		return g.Wait()
	})

}

func doInCore(parent glue.Context, withBean interface{}, cb func(core glue.Context) error) error {

	list := parent.Bean(sprint.CoreScannerClass, glue.DefaultLevel)
	if len(list) != 1 {
		return errors.Errorf("expected one core scanner in context, but found %d", len(list))
	}

	core, err := parent.Extend(list[0].Object().(sprint.CoreScanner).CoreBeans()...)
	if err != nil {
		return errors.Errorf("failed to create core context, %v", err)
	}
	defer core.Close()

	err = core.Inject(withBean)
	if err != nil {
		return err
	}

	return cb(core)
}


