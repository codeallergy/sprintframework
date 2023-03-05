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
)

type implCertCommand struct {
	Context     glue.Context    `inject`
	Application sprint.Application `inject`
}

type coreDomainContext struct {
	CertificateService sprint.CertificateService `inject`
}

func CertCommand() sprint.Command {
	return &implCertCommand{}
}

func (t *implCertCommand) BeanName() string {
	return "cert"
}

func (t *implCertCommand) Desc() string {
	return "cert commands: [list, dump, upload, create, renew, remove, client, acme, self, manager]"
}

func (t *implCertCommand) Run(args []string) error {
	if len(args) == 0 {
		return errors.Errorf("cert command needs argument, %s", t.Desc())
	}
	cmd := args[0]
	args = args[1:]

	err := doWithControlClient(t.Context, func(client sprint.ControlClient) error {
		content, err := client.CertificateCommand(cmd, args)
		if err == nil {
			println(content)
		}
		return err
	})
	if err == nil {
		return nil
	}
	if status.Code(err) != codes.Unavailable {
		return err
	}

	if cmd == "manager" {
		return errors.New("cert manager command available only on running server")
	}

	c := new(coreDomainContext)
	return doInCore(t.Context, c, func(core glue.Context) error {
		content, err :=  c.CertificateService.ExecuteCommand(cmd, args)
		if err != nil {
			return err
		}
		println(content)
		return nil
	})

}
