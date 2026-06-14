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

package domain

import (
	"errors"
	"strings"
	"time"
)

var ErrEntityNotFound error = errors.New("entity not found")

const (
	NatureHome     string = "Home"
	NatureBusiness string = "Business"
	NatureMobile   string = "Mobile"
	NatureOther    string = "Other"
)

type EmailAddress string

func (ea EmailAddress) String() string {
	return string(ea)
}

func (ea EmailAddress) Empty() bool {
	return ea == ""
}

type NamedEmailAddress struct {
	Address EmailAddress
	Name    string
}

func NewNamedEmailAddress(address, name string) NamedEmailAddress {
	return NamedEmailAddress{
		Address: EmailAddress(strings.TrimSpace(address)),
		Name:    strings.TrimSpace(name),
	}
}

func (nea *NamedEmailAddress) String() string {
	buffer := &strings.Builder{}
	if nea.Name != "" {
		buffer.WriteString(nea.Name)
		buffer.WriteString(" (")
		buffer.WriteString(string(nea.Address))
		buffer.WriteString(")")
	} else {
		buffer.WriteString(string(nea.Address))
	}
	return buffer.String()
}

func (nea *NamedEmailAddress) Empty() bool {
	return nea.Address.Empty()
}

type PhoneNumber string

func (pn PhoneNumber) String() string {
	return string(pn)
}

func (pn PhoneNumber) Empty() bool {
	return pn == ""
}

type PostalAddress struct {
	Street     string
	City       string
	PostalCode string
	Country    string
}

func (pa *PostalAddress) String() string {
	buffer := &strings.Builder{}
	if pa.Street != "" {
		buffer.WriteString(pa.Street)
	}
	if pa.City != "" {
		if buffer.Len() > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(pa.City)
	}
	if pa.PostalCode != "" {
		if buffer.Len() > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(pa.PostalCode)
	}
	if pa.Country != "" {
		if buffer.Len() > 0 {
			buffer.WriteString(", ")
		}
		buffer.WriteString(pa.Country)
	}
	return buffer.String()
}

func (pa *PostalAddress) Empty() bool {
	return pa.Street == "" && pa.City == "" && pa.PostalCode == "" && pa.Country == ""
}

type TZTime struct {
	DateTime time.Time
	Timezone string
}

func NewTZTime(dateTime time.Time, timezone string) TZTime {
	return TZTime{
		DateTime: dateTime,
		Timezone: timezone,
	}
}

func (tzt *TZTime) String() string {
	return tzt.DateTime.Format(time.DateTime)
}

func (tzt *TZTime) Empty() bool {
	return tzt.DateTime.IsZero()
}

type StandardFilter struct {
	Query *string
	Limit *int
}
