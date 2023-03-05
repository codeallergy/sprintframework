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
	"github.com/codeallergy/glue"
	"github.com/codeallergy/sprint"
	"github.com/go-errors/errors"
	"net/http"
	"strings"
)

type implRedirectHttpsPage struct {
	Properties glue.Properties `inject`

	beanName       string
	redirectAddr   string
	redirectSuffix string
}

func RedirectHttpsPage(beanName string) sprint.Page {
	return &implRedirectHttpsPage{
		beanName: beanName,
	}
}

func (t *implRedirectHttpsPage) BeanName() string {
	return t.beanName
}

func (t *implRedirectHttpsPage) PostConstruct() (err error) {
	t.redirectAddr = t.Properties.GetString(fmt.Sprintf("%s.%s", t.beanName, "redirect-address"), "")
	if t.redirectAddr == "" {
		return errors.Errorf("property '%s.redirect-address' is not found in context", t.beanName)
	}

	i := strings.IndexByte(t.redirectAddr, ':')
	if i != -1 {
		t.redirectSuffix = t.redirectAddr[i:]
	} else {
		t.redirectSuffix = ""
	}

	return
}

func (t *implRedirectHttpsPage) Pattern() string {
	return "/"
}

func (t *implRedirectHttpsPage) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	defer func() {
		if r := recover(); r != nil {
			http.Error(w, fmt.Sprintf("%v", r), http.StatusInternalServerError)
		}
	}()

	hostname := strings.Split(req.Host, ":")[0]
	url := fmt.Sprintf("https://%s%s%s", hostname, t.redirectSuffix, req.RequestURI)
	http.Redirect(w, req, url, http.StatusMovedPermanently)
}
