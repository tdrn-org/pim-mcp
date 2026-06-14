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
	"context"
	"strings"
	"time"
)

type Contact struct {
	ID           string
	DisplayName  string
	FirstName    string
	LastName     string
	Emails       ContactEmailAddresses
	Phones       ContactPhoneNumbers
	Organization string
	JobTitle     string
	Addresses    ContactPostalAddresses
	UpdatedAt    time.Time
}

func (c *Contact) String() string {
	buffer := &strings.Builder{}
	if c.DisplayName != "" {
		buffer.WriteRune(' ')
		buffer.WriteString(c.DisplayName)
	} else {
		if c.FirstName != "" {
			buffer.WriteRune(' ')
			buffer.WriteString(c.FirstName)
		}
		if c.LastName != "" {
			buffer.WriteRune(' ')
			buffer.WriteString(c.LastName)
		}
	}
	for _, email := range c.Emails {
		buffer.WriteRune(' ')
		buffer.WriteString(email.Address.String())
	}
	for _, phone := range c.Phones {
		buffer.WriteRune(' ')
		buffer.WriteString(phone.Number.String())
	}
	return buffer.String()
}

func (c *Contact) Empty() bool {
	return c.ID == "" || (c.DisplayName == "" && c.FirstName == "" && c.LastName == "")
}

type ContactEmailAddress struct {
	Address EmailAddress
	Nature  string
}

func NewContactEmailAddress(address, nature string) ContactEmailAddress {
	return ContactEmailAddress{
		Address: EmailAddress(strings.TrimSpace(address)),
		Nature:  nature,
	}
}

func (e *ContactEmailAddress) Empty() bool {
	return e.Address.Empty()
}

type ContactEmailAddresses []ContactEmailAddress

func (emails ContactEmailAddresses) Addresses() []string {
	addresses := make([]string, len(emails))
	for _, email := range emails {
		addresses = append(addresses, email.Address.String())
	}
	return addresses
}

type ContactPhoneNumber struct {
	Number PhoneNumber
	Nature string
}

func NewContactPhoneNumber(number, nature string) ContactPhoneNumber {
	return ContactPhoneNumber{
		Number: PhoneNumber(strings.TrimSpace(number)),
		Nature: nature,
	}
}

func (p *ContactPhoneNumber) Empty() bool {
	return p.Number.Empty()
}

type ContactPhoneNumbers []ContactPhoneNumber

func (phoneNumbers ContactPhoneNumbers) Numbers() []string {
	numbers := make([]string, len(phoneNumbers))
	for _, phoneNumber := range phoneNumbers {
		numbers = append(numbers, phoneNumber.Number.String())
	}
	return numbers
}

type ContactPostalAddress struct {
	PostalAddress
	Nature string
}

func NewContactPostalAddress(street, city, postalCode, country, nature string) ContactPostalAddress {
	return ContactPostalAddress{
		PostalAddress: PostalAddress{
			Street:     strings.TrimSpace(street),
			City:       strings.TrimSpace(city),
			PostalCode: strings.TrimSpace(postalCode),
			Country:    strings.TrimSpace(country),
		},
		Nature: strings.TrimSpace(nature),
	}
}

func (p *ContactPostalAddress) Empty() bool {
	return p.PostalAddress.Empty()
}

type ContactPostalAddresses []ContactPostalAddress

func (postalAddresses ContactPostalAddresses) Addresses() []string {
	addresses := make([]string, len(postalAddresses))
	for _, postalAddress := range postalAddresses {
		addresses = append(addresses, postalAddress.Street+" "+postalAddress.City+" "+postalAddress.PostalCode+" "+postalAddress.Country)
	}
	return addresses
}

type ContactProvider interface {
	SearchContacts(ctx context.Context, filter ContactFilter) ([]*Contact, error)
	GetContact(ctx context.Context, id string) (*Contact, error)
}

type ContactFilter struct {
	StandardFilter
}
