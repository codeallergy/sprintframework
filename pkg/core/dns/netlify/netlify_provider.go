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

package netlify

import (
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/netlify"
	"github.com/pkg/errors"
	"github.com/codeallergy/glue"
	"github.com/codeallergy/sprint"
	"os"
	"strings"
)

type implNetlifyProvider struct {
	Properties   glue.Properties  `inject`
}

func NetlifyProvider() sprint.DNSProvider {
	return &implNetlifyProvider{}
}

func (t *implNetlifyProvider) BeanName() string {
	return "netlify_provider"
}

func (t *implNetlifyProvider) Detect(whois *sprint.Whois) bool {
	for _, ns := range whois.NServer {
		if strings.HasSuffix(strings.ToLower(ns), ".nsone.net") {
			return true
		}
	}
	return false
}

func (t *implNetlifyProvider) RegisterChallenge(legoClient interface{}, token string) error {

	client, ok := legoClient.(*lego.Client)
	if !ok {
		return errors.Errorf("expected *lego.Client instance")
	}

	if token == "" {
		token = t.Properties.GetString("netlify.token", "")
	}

	if token == "" {
		token = os.Getenv("NETLIFY_TOKEN")
	}

	if token == "" {
		return errors.New("netlify token not found")
	}

	conf := netlify.NewDefaultConfig()
	conf.Token = token

	prov, err := netlify.NewDNSProviderConfig(conf)
	if err != nil {
		return err
	}

	return client.Challenge.SetDNS01Provider(prov)
}


func (t *implNetlifyProvider) NewClient() (sprint.DNSProviderClient, error) {

	token := t.Properties.GetString("netlify.token", "")

	if token == "" {
		token = os.Getenv("NETLIFY_TOKEN")
	}

	if token == "" {
		return nil, errors.New("netlify.token is empty in config and empty system env NETLIFY_TOKEN")
	}

	return NewClient(token), nil
}