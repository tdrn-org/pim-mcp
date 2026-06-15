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

package mcp

import (
	"time"

	"github.com/tdrn-org/pim-mcp/internal/domain"
)

type TZTimeOutput struct {
	DateTime time.Time `json:"date_time" jsonschema:"The date-time value of this point in time (RFC3339 format)."`
	Timezone string    `json:"timezone" jsonschema:"The timezone this point in time has been created in."`
}

func toTZTimeOutput(tzt domain.TZTime) TZTimeOutput {
	return TZTimeOutput{
		DateTime: tzt.DateTime,
		Timezone: tzt.Timezone,
	}
}

func toTZTimeOutputPtr(tzt *domain.TZTime) *TZTimeOutput {
	if tzt == nil {
		return nil
	}
	output := toTZTimeOutput(*tzt)
	return &output
}

type NamedEmailAddressOutput struct {
	Address string `json:"email" jsonschema:"The email address."`
	Name    string `json:"name" jsonschema:"The display name associated with this email address."`
}

func toNamedEmailAddressOutput(nea domain.NamedEmailAddress) NamedEmailAddressOutput {
	return NamedEmailAddressOutput{
		Address: string(nea.Address),
		Name:    nea.Name,
	}
}

func toNamedEmailAddressOutputs(neas []domain.NamedEmailAddress) []NamedEmailAddressOutput {
	outputs := make([]NamedEmailAddressOutput, 0, len(neas))
	for _, nea := range neas {
		outputs = append(outputs, toNamedEmailAddressOutput(nea))
	}
	return outputs
}
