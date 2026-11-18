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
	"github.com/tdrn-org/go-config-toml"
	"github.com/tdrn-org/pim-mcp/internal/domain"
)

type DMSConfig struct {
	Adapter      DMSAdapter         `toml:"adapter"`
	AccessMode   DMSAccessMode      `toml:"access_mode"`
	PaperlessNGX PaperlessNGXConfig `toml:"paperlessngx"`
}

type PaperlessNGXConfig struct {
	APIURL config.URLSpec `toml:"api_url"`
	APIKey string         `toml:"api_key"`
}

type DMSAdapter string

const (
	DMSAdapterNone         DMSAdapter = "none"
	DMSAdapterPaperlessNGX DMSAdapter = "paperlessngx"
)

var dmsAdapterMarshalMap map[DMSAdapter]string = map[DMSAdapter]string{
	DMSAdapterNone:         string(DMSAdapterNone),
	DMSAdapterPaperlessNGX: string(DMSAdapterPaperlessNGX),
}

var dmsAdapterUnmarshalMap map[string]DMSAdapter = map[string]DMSAdapter{
	string(DMSAdapterNone):         DMSAdapterNone,
	string(DMSAdapterPaperlessNGX): DMSAdapterPaperlessNGX,
}

func (a DMSAdapter) String() string {
	return dmsAdapterMarshalMap[a]
}

func (a DMSAdapter) MarshalText() ([]byte, error) {
	return config.MarshalEnum(a, dmsAdapterMarshalMap)
}

func (a *DMSAdapter) UnmarshalText(text []byte) error {
	adapter, err := config.UnmarshalEnum(dmsAdapterUnmarshalMap, text)
	if err != nil {
		return err
	}
	*a = adapter
	return nil
}

type DMSAccessMode string

const (
	DMSAccessModeRO DMSAccessMode = DMSAccessMode(domain.ReadOnly)
	DMSAccessModeRW DMSAccessMode = DMSAccessMode(domain.ReadWrite)
)

var dmsAccessMarshalMap map[DMSAccessMode]string = map[DMSAccessMode]string{
	DMSAccessModeRO: string(DMSAccessModeRO),
	DMSAccessModeRW: string(DMSAccessModeRW),
}

var dmsAccessUnmarshalMap map[string]DMSAccessMode = map[string]DMSAccessMode{
	string(DMSAccessModeRO): DMSAccessModeRO,
	string(DMSAccessModeRW): DMSAccessModeRW,
}

func (m DMSAccessMode) String() string {
	return dmsAccessMarshalMap[m]
}

func (m DMSAccessMode) MarshalText() ([]byte, error) {
	return config.MarshalEnum(m, dmsAccessMarshalMap)
}

func (m *DMSAccessMode) UnmarshalText(text []byte) error {
	mode, err := config.UnmarshalEnum(dmsAccessUnmarshalMap, text)
	if err != nil {
		return err
	}
	*m = mode
	return nil
}
