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
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type implStopCommand struct {
	Application      sprint.Application      `inject`
	ApplicationFlags sprint.ApplicationFlags `inject`
	Context          glue.Context         `inject`

	RunDir           string       `value:"application.run.dir,default="`
}

func StopCommand() sprint.Command {
	return &implStopCommand{}
}

func (t *implStopCommand) BeanName() string {
	return "stop"
}

func (t *implStopCommand) Desc() string {
	return "stop server"
}

func (t *implStopCommand) Run(args []string) error {

	err := doWithControlClient(t.Context, func(client sprint.ControlClient) error {
		status, err := client.Shutdown(false)
		if err == nil {
			println(status)
		}
		return err
	})

	if err != nil {
		return t.KillServer()
	}

	return nil
}

func (t *implStopCommand) KillServer() error {

	runDir := t.RunDir
	if runDir == "" {
		runDir = filepath.Join(t.Application.ApplicationDir(), "run")
	}
	pidFile := filepath.Join(runDir, fmt.Sprintf("%s.pid", t.Application.Name()))

	blob, err := ioutil.ReadFile(pidFile)
	if err != nil {
		return err
	}

	pid := string(blob)

	if _, err := strconv.Atoi(pid); err != nil {
		return errors.Errorf("Invalid pid %s, %v", pid, err)
	}

	cmd := exec.Command("kill", "-2", pid)
	if err := cmd.Start(); err != nil {
		return err
	}

	if err := os.Remove(pidFile); err != nil {
		return errors.Errorf("Can not remove file %s, %v", pidFile, err)
	}

	return cmd.Wait()

}
