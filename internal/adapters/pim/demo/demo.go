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

package demo

import (
	"context"
	"log/slog"
	"net/url"

	"github.com/google/uuid"
	"github.com/tdrn-org/go-httpserver"
	"github.com/tdrn-org/pim-mcp/internal/adapters/pim"
	"github.com/tdrn-org/pim-mcp/internal/domain"
)

const Name = "demo"

type Runtime interface {
	BaseURL() *url.URL
	Logger() *slog.Logger
}

type Provider struct {
	runtime Runtime
	id      string
}

func NewProvider(runtime Runtime) *Provider {
	return &Provider{
		runtime: runtime,
		id:      uuid.NewString(),
	}
}

func (p *Provider) ID() string {
	return p.id
}

func (*Provider) Name() string {
	return Name
}

func (p *Provider) Capabilities() domain.ProviderCapabilities {
	return domain.AllProviderCapabilities(domain.ReadWrite)
}

func (p *Provider) Mount(server *httpserver.Instance) {
}

func (p *Provider) LoginURL() *url.URL {
	return p.runtime.BaseURL()
}

func (p *Provider) CheckCredentials(ctx context.Context, credentials string) *pim.CredentialInfo {
	return &pim.CredentialInfo{Valid: true}
}

func (p *Provider) RefreshCredentials(ctx context.Context, credentials string) string {
	return credentials
}
