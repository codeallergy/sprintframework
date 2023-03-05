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
	"github.com/pkg/errors"
)

type implJobCommand struct {
	Context glue.Context `inject`
}

func JobCommand() sprint.Command {
	return &implJobCommand{}
}

func (t *implJobCommand) BeanName() string {
	return "job"
}

func (t *implJobCommand) Desc() string {
	return "job management - [list, run, cancel]"
}

func (t *implJobCommand) Run(args []string) error {

	if len(args) < 1 {
		return errors.New("job management commands: [list, run, cancel]")
	}

	command := args[0]
	args = args[1:]

	return doWithControlClient(t.Context, func(client sprint.ControlClient) error {
		output, err := client.JobCommand(command, args)
		if err != nil {
			return err
		}
		println(output)
		return nil
	})

}