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
	"flag"
	"fmt"
	"github.com/codeallergy/sprint"
)

type implHelpCommand struct {
	Application sprint.Application `inject`
	FlagSet     *flag.FlagSet     `inject`
	Commands    []sprint.Command   `inject:"lazy"`
}

func HelpCommand() sprint.Command {
	return &implHelpCommand{}
}

func (t *implHelpCommand) BeanName() string {
	return "help"
}

func (t *implHelpCommand) Desc() string {
	return "help command"
}

func (t *implHelpCommand) Run(args []string) error {

	fmt.Printf("Usage: ./%s [command]\n", t.Application.Executable())

	for _, command := range t.Commands {
		fmt.Printf("    %s - %s\n", command.BeanName(), command.Desc())
	}

	fmt.Println("Flags:")
	t.FlagSet.PrintDefaults()
	return nil
}
