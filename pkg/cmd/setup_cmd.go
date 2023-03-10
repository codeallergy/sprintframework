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
	"encoding/base64"
	"fmt"
	"github.com/codeallergy/glue"
	"github.com/codeallergy/sprint"
	"github.com/codeallergy/sprintframework/pkg/util"
	"strings"
	"time"
	"os/user"
)

type implSetupCommand struct {
	Context     glue.Context    `inject`
	Application sprint.Application `inject`
	Properties  glue.Properties `inject`
}

type coreSetupContext struct {
	ConfigRepository sprint.ConfigRepository `inject`
}

func SetupCommand() sprint.Command {
	return &implSetupCommand{}
}

func (t *implSetupCommand) BeanName() string {
	return "setup"
}

func (t *implSetupCommand) Desc() string {
	return "setup command"
}

func (t *implSetupCommand) Run(args []string) error {

	boot, err := util.GenerateToken()
	if err != nil {
		return err
	}

	env := strings.ToUpper(fmt.Sprintf("%s_%s", t.Application.Name(), "boot"))
	fmt.Printf("export %s=%s\n", env, boot)

	t.Properties.Set("application.boot", boot)

	var secretKey []byte

	c := new(coreSetupContext)
	err = doInCore(t.Context, c, func(core glue.Context) error {
		secret, err := c.ConfigRepository.Get("jwt.secret.key")
		if err != nil {
			return err
		}
		if secret == "" {
			secret, err = util.GenerateToken()
			if err != nil {
				return err
			}
			err = c.ConfigRepository.Set("jwt.secret.key", secret)
			if err != nil {
				return err
			}
		}
		secretKey, err = base64.RawURLEncoding.DecodeString(secret)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	user, err := user.Current()
	if err != nil {
		return err
	}

	roles := map[string]bool {
		"USER": true,
		"ADMIN": true,
	}

	authUser := &sprint.AuthorizedUser{
		Username:  user.Username,
		Roles:     roles,
		Context:   nil,
		ExpiresAt: time.Now().Unix() + 356*24*3600,
	}

	auth, err := util.GenerateAuthToken(secretKey, authUser)
	if err != nil {
		return err
	}

	env = strings.ToUpper(fmt.Sprintf("%s_%s", t.Application.Name(), "auth"))
	fmt.Printf("export %s=%s\n", env, auth)

	return nil
}