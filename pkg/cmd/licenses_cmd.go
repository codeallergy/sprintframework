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
	"github.com/codeallergy/sprint"
)

type implLicensesCommand struct {
	ResourceService sprint.ResourceService `inject`
}

func LicensesCommand() sprint.Command {
	return &implLicensesCommand{}
}

func (t *implLicensesCommand) BeanName() string {
	return "licenses"
}

func (t *implLicensesCommand) Desc() string {
	return "show all licenses"
}

func (t *implLicensesCommand) Run(args []string) error {
	content, err := t.ResourceService.GetLicenses("resources:licenses.txt")
	if err != nil {
		return err
	}
	print(content)
	return nil
}
