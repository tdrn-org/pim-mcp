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

type PIMConfig struct {
	Adapter    PIMAdapter    `toml:"adapter"`
	AccessMode PIMAccessMode `toml:"access_mode"`
	MSGraph    MSGraphConfig `toml:"msgraph"`
}

type MSGraphConfig struct {
	ClientID         string `toml:"client_id"`
	ClientSecret     string `toml:"client_secret"`
	TenantID         string `toml:"tenant_id"`
	SensitivityLimit int    `toml:"sensitivity_limit"`
}

type PIMAdapter string

const (
	PIMAdapterDemo    PIMAdapter = "demo"
	PIMAdapterMSGraph PIMAdapter = "msgraph"
)

var knownPIMAdapters map[string]PIMAdapter = map[string]PIMAdapter{
	string(PIMAdapterDemo):    PIMAdapterDemo,
	string(PIMAdapterMSGraph): PIMAdapterMSGraph,
}

func (p *PIMAdapter) Value() string {
	for value, adapter := range knownPIMAdapters {
		if *p == adapter {
			return value
		}
	}
	slog.Warn("unexpected PIM adapter", slog.Any("p", *p))
	return ""
}

func (p *PIMAdapter) MarshalTOML() ([]byte, error) {
	return []byte(`"` + p.Value() + `"`), nil
}

func (p *PIMAdapter) UnmarshalTOML(value any) error {
	adapterString, ok := value.(string)
	if !ok {
		return fmt.Errorf("unexpected PIM adapter type %v", value)
	}
	adapter, ok := knownPIMAdapters[adapterString]
	if !ok {
		return fmt.Errorf("unknown PIM adapter: '%s'", adapterString)
	}
	*p = adapter
	return nil
}

type PIMAccessMode string

const (
	PIMAccessModeRO PIMAccessMode = PIMAccessMode(domain.ReadOnly)
	PIMAccessModeRW PIMAccessMode = PIMAccessMode(domain.ReadWrite)
)

var knownPIMAccessModes map[string]PIMAccessMode = map[string]PIMAccessMode{
	string(PIMAccessModeRO): PIMAccessModeRO,
	string(PIMAccessModeRW): PIMAccessModeRW,
}

func (m *PIMAccessMode) Value() string {
	for value, mode := range knownPIMAccessModes {
		if *m == mode {
			return value
		}
	}
	slog.Warn("unexpected PIM access mode", slog.Any("m", *m))
	return ""
}

func (m *PIMAccessMode) MarshalTOML() ([]byte, error) {
	return []byte(`"` + m.Value() + `"`), nil
}

func (m *PIMAccessMode) UnmarshalTOML(value any) error {
	modeString, ok := value.(string)
	if !ok {
		return fmt.Errorf("unexpected PIM access mode type %v", value)
	}
	mode, ok := knownPIMAccessModes[modeString]
	if !ok {
		return fmt.Errorf("unknown PIM access mode: '%s'", modeString)
	}
	*m = mode
	return nil
}
