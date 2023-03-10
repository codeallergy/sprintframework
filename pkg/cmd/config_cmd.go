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
	"github.com/codeallergy/sprintframework/pkg/app"
	"github.com/codeallergy/sprint"
	"github.com/codeallergy/sprintframework/pkg/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
)

type implConfigCommand struct {
	Context     glue.Context    `inject`
	Application sprint.Application `inject`
}

type coreConfigContext struct {
	ConfigRepository sprint.ConfigRepository `inject`
}

func ConfigCommand() sprint.Command {
	return &implConfigCommand{}
}

func (t *implConfigCommand) BeanName() string {
	return "config"
}

func (t *implConfigCommand) Desc() string {
	return "config commands: [get, set, dump, list]"
}

func (t *implConfigCommand) Run(args []string) error {
	if len(args) == 0 {
		return errors.Errorf("config command needs argument, %s", t.Desc())
	}
	cmd := args[0]
	args = args[1:]
	switch cmd {
	case "get":
		return t.getConfig(args)

	case "set":
		return t.setConfig(args)

	case "dump", "list":
		return t.dumpConfig(cmd, args)

	default:
		return errors.Errorf("unknown sub-command for config '%s'", cmd)
	}

	return nil
}

func (t *implConfigCommand) getConfig(args []string) error {
	if len(args) < 1 {
		return errors.Errorf("'config get' command expected key argument: %v", args)
	}
	key := args[0]

	var value string
	err := doWithControlClient(t.Context, func(client sprint.ControlClient) (err error) {
		value, err = client.ConfigCommand("get", []string {key})
		return
	})
	if err != nil && status.Code(err) == codes.Unavailable {
		value, err = t.getFromStorage(key)
	}
	if err != nil {
		return err
	}
	println(value)
	return nil
}

func (t *implConfigCommand) setConfig(args []string) error {
	if len(args) < 1 {
		return errors.Errorf("'config set' command expected key argument: %v", args)
	}

	key := args[0]
	args = args[1:]

	var value string
	if len(args) < 1 {
		if app.IsPEMProperty(key) {
			var err error
			value, err = util.PromptPEM("Enter PEM key: ")
			if err != nil {
				return err
			}
		} else if app.IsPasswordProperty(key) {
			value = util.PromptPassword("Enter password: ")
		} else {
			value = util.Prompt("Enter value: ")
		}

	} else {
		value = args[0]
		args = args[1:]
	}

	// value is the file path
	if strings.HasPrefix(value, "@") {
		filePath := value[1:]
		binVal, err := ioutil.ReadFile(filePath)
		if err != nil {
			return errors.Errorf("i/o error on reading value from file '%s', %v", filePath, err)
		}
		value = string(binVal)
	}

	err := doWithControlClient(t.Context, func(client sprint.ControlClient) error {
		_, err := client.ConfigCommand("set", []string{ key, value })
		return err
	})

	if err != nil && status.Code(err) == codes.Unavailable  {
		fmt.Printf("Error on gRPC: %v\n", err)
		err = t.setInStorage(key, value)
	}
	if err != nil {
		return err
	}
	println("SUCCESS")
	return nil
}

func (t *implConfigCommand) dumpConfig(cmd string, args []string) error {
	err := doWithControlClient(t.Context, func(client sprint.ControlClient) error {
		content, err := client.ConfigCommand(cmd, args)
		if err == nil {
			println(content)
		}
		return err
	})
	if err != nil && status.Code(err) == codes.Unavailable  {
		return t.dumpFromStorage(cmd, args, os.Stdout)
	}
	return err
}

func (t *implConfigCommand) dumpFromStorage(cmd string, args []string, writer io.StringWriter) (err error) {

	var prefix string
	if len(args) > 0 {
		prefix = args[0]
		args = args[1:]
	}

	limit := math.MaxInt64
	if cmd == "list" {
		limit = 80
	}

	if len(args) > 0 {
		limit, err = strconv.Atoi(args[0])
		if err != nil {
			return errors.Errorf("parsing limit '%s', %v", args[0], err)
		}
	}

	c := new(coreConfigContext)
	return doInCore(t.Context, c, func(core glue.Context) error {
		return c.ConfigRepository.EnumerateAll(prefix, func(key, value string) bool {
			if len(value) > limit {
				value = value[:limit] + "..."
				value = strings.ReplaceAll(value, "\n", " ")
			}
			writer.WriteString(fmt.Sprintf("%s: %s\n", key, value))
			return true
		})
	})
}

func (t *implConfigCommand) getFromStorage(key string) (value string, err error) {
	c := new(coreConfigContext)
	err = doInCore(t.Context, c, func(core glue.Context) error {
		value, err = c.ConfigRepository.Get(key)
		return err
	})
	return
}

func (t *implConfigCommand) setInStorage(key, value string) error {
	c := new(coreConfigContext)
	return doInCore(t.Context, c, func(core glue.Context) error {
		return c.ConfigRepository.Set(key, value)
	})
}
