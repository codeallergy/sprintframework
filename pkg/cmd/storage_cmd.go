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
	"github.com/pkg/errors"
	"github.com/codeallergy/glue"
	"github.com/codeallergy/sprint"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"os"
)

type implStorageCommand struct {
	Context          glue.Context           `inject`
}

type coreStorageContext struct {
	StorageService sprint.StorageService `inject`
}

func StorageCommand() sprint.Command {
	return &implStorageCommand{}
}

func (t *implStorageCommand) BeanName() string {
	return "storage"
}

func (t *implStorageCommand) Desc() string {
	return "storage management commands: [console, list, dump, restore, compact, drop, clean]"
}

func (t *implStorageCommand) Run(args []string) error {

	if len(args) < 1 {
		return errors.New(t.Desc())
	}

	cmd := args[0]
	args = args[1:]

	err := doWithControlClient(t.Context, func(client sprint.ControlClient) error {
		if cmd == "console" {
			return client.StorageConsole(os.Stdout, os.Stderr)
		} else {
			output, err := client.StorageCommand(cmd, args)
			if err != nil {
				return err
			}
			println(output)
			return nil
		}
	})
	if err == nil {
		return nil
	}
	if status.Code(err) != codes.Unavailable {
		return err
	}

	c := new(coreStorageContext)
	return doInCore(t.Context, c, func(core glue.Context) error {
		if cmd == "console" {
			return c.StorageService.LocalConsole(os.Stdout, os.Stderr)
		} else {
			content, err :=  c.StorageService.ExecuteCommand(cmd, args)
			if err != nil {
				return err
			}
			println(content)
			return nil
		}
	})

}



