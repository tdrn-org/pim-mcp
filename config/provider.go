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

package config

import (
	"fmt"
	"log/slog"

	"github.com/tdrn-org/pim-mcp/internal/domain"
)

type ProviderConfig struct {
	Adapter    ProviderAdapter    `toml:"adapter"`
	AccessMode ProviderAccessMode `toml:"access_mode"`
	MSGraph    MSGraphConfig      `toml:"msgraph"`
}

type MSGraphConfig struct {
	ClientID         string `toml:"client_id"`
	ClientSecret     string `toml:"client_secret"`
	TenantID         string `toml:"tenant_id"`
	SensitivityLimit int    `toml:"sensitivity_limit"`
}

type ProviderAdapter string

const (
	ProviderAdapterDemo    ProviderAdapter = "demo"
	ProviderAdapterMSGraph ProviderAdapter = "msgraph"
)

var knownProviderAdapters map[string]ProviderAdapter = map[string]ProviderAdapter{
	string(ProviderAdapterDemo):    ProviderAdapterDemo,
	string(ProviderAdapterMSGraph): ProviderAdapterMSGraph,
}

func (p *ProviderAdapter) Value() string {
	for value, adapter := range knownProviderAdapters {
		if *p == adapter {
			return value
		}
	}
	slog.Warn("unexpected provider adapter", slog.Any("p", *p))
	return ""
}

func (p *ProviderAdapter) MarshalTOML() ([]byte, error) {
	return []byte(`"` + p.Value() + `"`), nil
}

func (p *ProviderAdapter) UnmarshalTOML(value any) error {
	adapterString, ok := value.(string)
	if !ok {
		return fmt.Errorf("unexpected provider adapter type %v", value)
	}
	adapter, ok := knownProviderAdapters[adapterString]
	if !ok {
		return fmt.Errorf("unknown provider adapter: '%s'", adapterString)
	}
	*p = adapter
	return nil
}

type ProviderAccessMode string

const (
	ProviderAccessModeRO ProviderAccessMode = ProviderAccessMode(domain.ReadOnly)
	ProviderAccessModeRW ProviderAccessMode = ProviderAccessMode(domain.ReadWrite)
)

var knownProviderAccessModes map[string]ProviderAccessMode = map[string]ProviderAccessMode{
	string(ProviderAccessModeRO): ProviderAccessModeRO,
	string(ProviderAccessModeRW): ProviderAccessModeRW,
}

func (m *ProviderAccessMode) Value() string {
	for value, mode := range knownProviderAccessModes {
		if *m == mode {
			return value
		}
	}
	slog.Warn("unexpected provider access mode", slog.Any("m", *m))
	return ""
}

func (m *ProviderAccessMode) MarshalTOML() ([]byte, error) {
	return []byte(`"` + m.Value() + `"`), nil
}

func (m *ProviderAccessMode) UnmarshalTOML(value any) error {
	modeString, ok := value.(string)
	if !ok {
		return fmt.Errorf("unexpected provider access mode type %v", value)
	}
	mode, ok := knownProviderAccessModes[modeString]
	if !ok {
		return fmt.Errorf("unknown provider access mode: '%s'", modeString)
	}
	*m = mode
	return nil
}
