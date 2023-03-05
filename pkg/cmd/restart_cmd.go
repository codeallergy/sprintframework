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
	"github.com/codeallergy/glue"
	"github.com/codeallergy/sprint"
)

type implRestartCommand struct {
	Context glue.Context `inject`
}

func RestartCommand() sprint.Command {
	return &implRestartCommand{}
}

func (t *implRestartCommand) BeanName() string {
	return "restart"
}

func (t *implRestartCommand) Desc() string {
	return "restart server"
}

func (t *implRestartCommand) Run(args []string) error {

	return doWithControlClient(t.Context, func(client sprint.ControlClient) error {
		status, err := client.Shutdown(true)
		if err == nil {
			println(status)
		}
		return err
	})

}
