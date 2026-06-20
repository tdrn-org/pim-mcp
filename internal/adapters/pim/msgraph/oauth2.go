/*
 * Copyright 2026 Holger de Carne
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package msgraph

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"golang.org/x/oauth2"
)

const graphScope string = "https://graph.microsoft.com/.default"

var graphTokenRequestOptions policy.TokenRequestOptions = policy.TokenRequestOptions{
	Scopes: []string{
		graphScope,
	},
}

func marshalToken(token *oauth2.Token) (string, error) {
	buffer := &strings.Builder{}
	err := json.NewEncoder(buffer).Encode(token)
	if err != nil {
		return "", fmt.Errorf("failed to marshal OAuth2 token (cause: %w)", err)
	}
	return buffer.String(), nil
}

func unmarshalToken(s string) (*oauth2.Token, error) {
	token := &oauth2.Token{}
	err := json.NewDecoder(strings.NewReader(s)).Decode(token)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal OAuth2 token (cause: %w)", err)
	}
	return token, nil
}

func (p *Provider) oauth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     p.cfg.MSGraph.ClientID,
		ClientSecret: p.cfg.MSGraph.ClientSecret,
		RedirectURL:  p.runtime.BaseURL().JoinPath("/msgraph/callback").String(),
		Scopes: []string{
			"offline_access",
			fmt.Sprintf("%s/access_as_user", p.cfg.MSGraph.ClientID),
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize", p.cfg.MSGraph.TenantID),
			TokenURL: fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", p.cfg.MSGraph.TenantID),
		},
	}
}
