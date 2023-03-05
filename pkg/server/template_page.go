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

package server

import (
	"fmt"
	"github.com/go-errors/errors"
	"github.com/codeallergy/glue"
	"github.com/codeallergy/sprint"
	"html/template"
	"net/http"
)

type implTemplatePage struct {
	glue.InitializingBean

	pattern      string
	templateFile string
	tpl          *template.Template

	ResourceService sprint.ResourceService `inject`
}

func TemplatePage(pattern, templateFile string) sprint.Page {
	return &implTemplatePage{
		pattern: pattern,
		templateFile: templateFile,
	}
}

func (t *implTemplatePage) PostConstruct() (err error) {
	t.tpl, err = t.ResourceService.HtmlTemplate(t.templateFile)
	if err != nil {
		return errors.Errorf("template index file '%s' error, %v", t.templateFile, err)
	}
	return
}

func (t *implTemplatePage) Pattern() string {
	return t.pattern
}

func (t *implTemplatePage) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	defer func() {
		if r := recover(); r != nil {
			http.Error(w, fmt.Sprintf("%v", r), http.StatusInternalServerError)
		}
	}()

	r.ParseForm()
	t.tpl.Execute(w, r)
}
