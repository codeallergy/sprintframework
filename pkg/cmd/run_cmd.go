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
	"fmt"
	"github.com/pkg/errors"
	"github.com/codeallergy/glue"
	"github.com/codeallergy/sprint"
	"go.uber.org/zap"
	"io"
	"log"
	"os"
	"path/filepath"
)

type implRunCommand struct {
	Application                       sprint.Application                       `inject`
	ApplicationFlags                  sprint.ApplicationFlags                  `inject`
	SystemEnvironmentPropertyResolver sprint.SystemEnvironmentPropertyResolver `inject`
	Context                           glue.Context                          `inject`
	StartCommand                      *implStartCommand                       `inject`
	CoreScanner                       sprint.CoreScanner                       `inject`

	LogDir         string        `value:"application.log.dir,default="`
	LogDirPerm     os.FileMode   `value:"application.perm.log.dir,default=-rwxrwxr-x"`
	LogFilePerm    os.FileMode   `value:"application.perm.log.file,default=-rw-rw-r--"`

	startupLog  *log.Logger
	logFile     *os.File
	logWriter   io.Writer
}

func RunCommand() sprint.Command {
	return &implRunCommand{}
}

func (t *implRunCommand) BeanName() string {
	return "run"
}

func (t *implRunCommand) Desc() string {
	return "run server"
}

func (t *implRunCommand) createLogFile() (string, error) {

	logDir := t.LogDir
	if logDir == "" {
		logDir = filepath.Join(t.Application.ApplicationDir(), "log")
	}

	if _, err := os.Stat(logDir); err != nil {
		if err = os.MkdirAll(logDir, t.LogDirPerm); err != nil {
			return "", err
		}
	}

	logFile := fmt.Sprintf("%s-startup.log", t.Application.Name())
	return logFile, nil
}

func (t *implRunCommand) lazyStartupLog() *log.Logger {
	var err error
	if t.startupLog == nil {

		t.logWriter = os.Stdout

		if t.ApplicationFlags.Daemon() {
			var fileName string
			fileName, err = t.createLogFile()
			if err == nil {
				t.logFile, err = os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, t.LogFilePerm)
				if err == nil {
					t.logWriter = t.logFile
				}
			}
		}

		t.startupLog = log.New(t.logWriter,
			"ERROR: ",
			log.Ldate|log.Ltime|log.Lshortfile)

		if err != nil {
			t.startupLog.Printf("Startup log file creation error, %v\n", err)
		}
	}
	return t.startupLog
}

func (t *implRunCommand) Destroy() error {
	if t.logFile != nil {
		return t.logFile.Close()
	}
	return nil
}

func (t *implRunCommand) Run(args []string) (err error) {

	beans := t.CoreScanner.CoreBeans()
	if t.ApplicationFlags.Verbose() {
		verbose := glue.Verbose{ Log: t.lazyStartupLog() }
		beans = append([]interface{}{verbose}, beans...)
	}

	core, err := t.Context.Extend(beans...)
	if err != nil {
		msg := fmt.Sprintf("core creation context failed by %v, used environment variables %+v", err, t.SystemEnvironmentPropertyResolver.Environ(false))
		t.lazyStartupLog().Println(msg)
		return errors.New(msg)
	}

	logger, ok := findZapLogger(core)
	if !ok {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	defer func() {
		if r := recover(); r != nil {
			logger.Error("ApplicationRecover", zap.Error(err))
		}
	}()
	
	err = runServers(t.Application, core, logger)
	if err != nil {
		logger.Error("ApplicationEnded", zap.Bool("restarting", t.Application.Restarting()), zap.Strings("env", t.SystemEnvironmentPropertyResolver.Environ(false)), zap.Error(err))
	} else {
		logger.Info("ApplicationEnded", zap.Bool("restarting", t.Application.Restarting()), zap.Strings("env", t.SystemEnvironmentPropertyResolver.Environ(false)))
	}

	err = core.Close()
	if err != nil {
		logger.Error("CoreContextClosed", zap.Error(err))
	} else {
		logger.Info("CoreContextClosed", )
	}

	logger.Sync()

	if t.Application.Restarting() {
		logger.Info("ApplicationRestarting")
		err = t.StartCommand.Start(logger, true)
		if err != nil {
			logger.Error("ApplicationRestart", zap.Strings("env", t.SystemEnvironmentPropertyResolver.Environ(false)), zap.Error(err))
		} else {
			logger.Info("ApplicationRestarted", zap.Strings("env", t.SystemEnvironmentPropertyResolver.Environ(false)))
		}
	}

	return
}

func findZapLogger(core glue.Context) (*zap.Logger, bool) {
	list := core.Bean(sprint.LogClass, glue.DefaultLevel)
	if len(list) > 0 {
		if l, ok := list[0].Object().(*zap.Logger); ok {
			return l, true
		}
	}
	return nil, false
}