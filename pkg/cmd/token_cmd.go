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
	"github.com/pkg/errors"
	"github.com/codeallergy/glue"
	"github.com/codeallergy/sprint"
	"github.com/codeallergy/sprintframework/pkg/util"
	"strconv"
	"strings"
	"time"
)

type implTokenCommand struct {
	Application      sprint.Application      `inject`
	ApplicationFlags sprint.ApplicationFlags `inject`
	Properties       glue.Properties      `inject`
}

func TokenCommand() sprint.Command {
	return &implTokenCommand{}
}

func (t *implTokenCommand) BeanName() string {
	return "token"
}

func (t *implTokenCommand) Desc() string {
	return "token generator commands [boot, auth, verify]"
}

func (t *implTokenCommand) Run(args []string) (err error) {

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

	if len(args) == 0 {
		return errors.Errorf("token command needs argument [boot, auth]")
	}
	cmd := args[0]
	args = args[1:]

	switch cmd {
	case "boot":
		return t.generateBootstrapToken(args)
	case "auth":
		return t.generateAuthToken(args)
	case "verify":
		return t.verifyAuthToken(args)
	default:
		return errors.Errorf("unknown sub-command '%s' for token command", cmd)
	}
}

func (t *implTokenCommand) generateBootstrapToken(args []string) error {
	if token, err := util.GenerateToken(); err == nil {
		println(token)
		return nil
	} else {
		return err
	}
}

func (t *implTokenCommand) generateAuthToken(args []string) error {

	if len(args) < 4 {
		return errors.Errorf("Usage: ./%s token auth username roles context ttl-in-days", t.Application.Executable())
	}

	username := args[0]
	roles := args[1]
	context := args[2]
	ttlDaysStr := args[3]

	contextMap := make(map[string]string)
	pairs := strings.Split(context, ",")
	for _, pair := range pairs {
		i := strings.IndexByte(pair, '=')
		if i == -1 {
			contextMap[pair] = ""
		} else {
			contextMap[pair[0:i]] = pair[i+1:]
		}
	}

	secret := util.PromptPassword("Enter JWT secret key: ")
	secretKey, err := base64.RawURLEncoding.DecodeString(secret)
	if err != nil {
		return err
	}

	ttlDays, err := strconv.ParseInt(ttlDaysStr, 10, 64)
	if err != nil {
		return errors.Errorf("error on parsing days '%s', %v", ttlDaysStr, err)
	}

	indexedRoles := make(map[string]bool)
	for _, role := range strings.Split(roles, ",") {
		indexedRoles[strings.TrimSpace(role)] = true
	}

	user := &sprint.AuthorizedUser{
		Username:  username,
		Roles:     indexedRoles,
		Context:   contextMap,
		ExpiresAt: time.Now().Unix() + ttlDays*24*3600,
	}

	token, err := util.GenerateAuthToken(secretKey, user)
	if err != nil {
		return err
	}

	println(token)
	return nil
}

func (t *implTokenCommand) verifyAuthToken(args []string) error {

	var authToken string
	if len(args) > 0 {
		authToken = args[0]
	} else {
		tokenKey := strings.ToUpper(fmt.Sprintf("%s_auth", t.Application.Name()))
		authToken = t.Properties.GetString(tokenKey, "")
	}

	if authToken == "" {
		return errors.New("auth token not found")
	}

	secret := util.PromptPassword("Enter JWT secret key: ")
	secretKey, err := base64.RawURLEncoding.DecodeString(secret)
	if err != nil {
		return err
	}

	user, err := util.VerifyAuthToken(secretKey, authToken)
	if err != nil {
		errors.Errorf("verify error, %v", err)
	}

	fmt.Printf("%s, %+v, %s, expires at %s\n", user.Username, user.Roles, user.Context, time.Unix(user.ExpiresAt, 0).String())
	return nil
}
