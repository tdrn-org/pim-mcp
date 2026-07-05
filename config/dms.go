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

type DMSConfig struct {
	Adapter      DMSAdapter         `toml:"adapter"`
	AccessMode   DMSAccessMode      `toml:"access_mode"`
	PaperlessNGX PaperlessNGXConfig `toml:"paperlessngx"`
}

type PaperlessNGXConfig struct {
	APIURL URLSpec `toml:"api_url"`
	APIKey string  `toml:"api_key"`
}

type DMSAdapter string

const (
	DMSAdapterNone         DMSAdapter = "none"
	DMSAdapterPaperlessNGX DMSAdapter = "paperlessngx"
)

var knownDMSAdapters map[string]DMSAdapter = map[string]DMSAdapter{
	string(DMSAdapterNone):         DMSAdapterNone,
	string(DMSAdapterPaperlessNGX): DMSAdapterPaperlessNGX,
}

func (p *DMSAdapter) Value() string {
	for value, adapter := range knownDMSAdapters {
		if *p == adapter {
			return value
		}
	}
	slog.Warn("unexpected DMS adapter", slog.Any("p", *p))
	return ""
}

func (p *DMSAdapter) MarshalTOML() ([]byte, error) {
	return []byte(`"` + p.Value() + `"`), nil
}

func (p *DMSAdapter) UnmarshalTOML(value any) error {
	adapterString, ok := value.(string)
	if !ok {
		return fmt.Errorf("unexpected DMS adapter type %v", value)
	}
	adapter, ok := knownDMSAdapters[adapterString]
	if !ok {
		return fmt.Errorf("unknown DMS adapter: '%s'", adapterString)
	}
	*p = adapter
	return nil
}

type DMSAccessMode string

const (
	DMSAccessModeRO DMSAccessMode = DMSAccessMode(domain.ReadOnly)
	DMSAccessModeRW DMSAccessMode = DMSAccessMode(domain.ReadWrite)
)

var knownDMSAccessModes map[string]DMSAccessMode = map[string]DMSAccessMode{
	string(DMSAccessModeRO): DMSAccessModeRO,
	string(DMSAccessModeRW): DMSAccessModeRW,
}

func (m *DMSAccessMode) Value() string {
	for value, mode := range knownDMSAccessModes {
		if *m == mode {
			return value
		}
	}
	slog.Warn("unexpected DMS access mode", slog.Any("m", *m))
	return ""
}

func (m *DMSAccessMode) MarshalTOML() ([]byte, error) {
	return []byte(`"` + m.Value() + `"`), nil
}

func (m *DMSAccessMode) UnmarshalTOML(value any) error {
	modeString, ok := value.(string)
	if !ok {
		return fmt.Errorf("unexpected DMS access mode type %v", value)
	}
	mode, ok := knownDMSAccessModes[modeString]
	if !ok {
		return fmt.Errorf("unknown DMS access mode: '%s'", modeString)
	}
	*m = mode
	return nil
}
