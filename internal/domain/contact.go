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
	Emails       EmailAddresses
	Phones       PhoneNumbers
	Organization string
	JobTitle     string
	Addresses    PostalAddresses
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
		buffer.WriteString(email.Address)
	}
	for _, phone := range c.Phones {
		buffer.WriteRune(' ')
		buffer.WriteString(phone.Number)
	}
	return buffer.String()
}

func (c *Contact) Empty() bool {
	return c.ID == "" || (c.DisplayName == "" && c.FirstName == "" && c.LastName == "")
}

type EmailAddress struct {
	Address string
	Nature  string
}

func NewEmailAddress(address, nature string) EmailAddress {
	return EmailAddress{
		Address: strings.TrimSpace(address),
		Nature:  nature,
	}
}

func (e *EmailAddress) Empty() bool {
	return e.Address == ""
}

type EmailAddresses []EmailAddress

func (emails EmailAddresses) Addresses() []string {
	addresses := make([]string, len(emails))
	for _, email := range emails {
		addresses = append(addresses, email.Address)
	}
	return addresses
}

type PhoneNumber struct {
	Number string
	Nature string
}

func NewPhoneNumber(number, nature string) PhoneNumber {
	return PhoneNumber{
		Number: strings.TrimSpace(number),
		Nature: nature,
	}
}

func (p *PhoneNumber) Empty() bool {
	return p.Number == ""
}

type PhoneNumbers []PhoneNumber

func (phoneNumbers PhoneNumbers) Numbers() []string {
	numbers := make([]string, len(phoneNumbers))
	for _, phoneNumber := range phoneNumbers {
		numbers = append(numbers, phoneNumber.Number)
	}
	return numbers
}

type PostalAddress struct {
	Street     string
	City       string
	PostalCode string
	Country    string
	Nature     string
}

func NewPostalAddress(street, city, postalCode, country, nature string) PostalAddress {
	return PostalAddress{
		Street:     strings.TrimSpace(street),
		City:       strings.TrimSpace(city),
		PostalCode: strings.TrimSpace(postalCode),
		Country:    strings.TrimSpace(country),
		Nature:     strings.TrimSpace(nature),
	}
}

func (p *PostalAddress) Empty() bool {
	return p.Street == "" && p.City == "" && p.PostalCode == "" && p.Country == ""
}

type PostalAddresses []PostalAddress

func (postalAddresses PostalAddresses) Addresses() []string {
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
	Query string
	Limit int
}
