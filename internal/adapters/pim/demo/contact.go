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
	"maps"
	"slices"
	"time"

	"github.com/tdrn-org/pim-mcp/internal/application"
	"github.com/tdrn-org/pim-mcp/internal/domain"
)

func (p *Provider) SearchContacts(ctx context.Context, filter domain.ContactFilter) ([]*domain.Contact, error) {
	contacts := slices.Collect(maps.Values(contactData))
	slices.SortFunc(contacts, application.ContactSortFunc)
	contactList := slices.Values(contacts)
	contactList = application.Match(contactList, filter.Query, application.EntityFilterFunc[*domain.Contact](application.ContactFilterFunc))
	contactList = application.Limit(contactList, filter.Limit)
	return slices.Collect(contactList), nil
}

func (p *Provider) GetContact(ctx context.Context, id string) (*domain.Contact, error) {
	contact, ok := contactData[id]
	if !ok {
		return nil, domain.ErrEntityNotFound
	}
	return contact, nil
}

var contactData map[string]*domain.Contact = map[string]*domain.Contact{
	"c1": {
		ID:          "c1",
		DisplayName: "Alice Chen",
		FirstName:   "Alice",
		LastName:    "Chen",
		Emails: []domain.ContactEmailAddress{
			domain.NewContactEmailAddress("alice.chen@example.org", domain.NatureBusiness),
			domain.NewContactEmailAddress("alice.chen.personal@example.org", domain.NatureHome),
		},
		Phones: []domain.ContactPhoneNumber{
			domain.NewContactPhoneNumber("+1-555-0101", domain.NatureBusiness),
			domain.NewContactPhoneNumber("+1-555-0201", domain.NatureMobile),
		},
		Organization: "Example Corp",
		JobTitle:     "Engineering Manager",
		Addresses: []domain.ContactPostalAddress{
			domain.NewContactPostalAddress("123 Main Street", "San Francisco", "94105", "USA", domain.NatureBusiness),
		},
		UpdatedAt: time.Date(2026, 6, 8, 12, 34, 56, 0, time.UTC),
	},
	"c2": {
		ID:          "c2",
		DisplayName: "Bob Martinez",
		FirstName:   "Bob",
		LastName:    "Martinez",
		Emails: []domain.ContactEmailAddress{
			domain.NewContactEmailAddress("bob.martinez@example.org", domain.NatureBusiness),
		},
		Phones: []domain.ContactPhoneNumber{
			domain.NewContactPhoneNumber("+1-555-0102", domain.NatureBusiness),
		},
		Organization: "Example Corp",
		JobTitle:     "Senior Software Engineer",
		Addresses:    []domain.ContactPostalAddress{},
		UpdatedAt:    time.Date(2026, 6, 9, 12, 34, 56, 0, time.UTC),
	},
	"c3": {
		ID:          "c3",
		DisplayName: "Carol Wang",
		FirstName:   "Carol",
		LastName:    "Wang",
		Emails: []domain.ContactEmailAddress{
			domain.NewContactEmailAddress("carol.wang@example.org", domain.NatureBusiness),
		},
		Phones: []domain.ContactPhoneNumber{
			domain.NewContactPhoneNumber("+1-555-0103", domain.NatureBusiness),
			domain.NewContactPhoneNumber("+1-555-0203", domain.NatureHome),
		},
		Organization: "Example Corp",
		JobTitle:     "Product Designer",
		Addresses: []domain.ContactPostalAddress{
			domain.NewContactPostalAddress("456 Oak Avenue", "New York", "10001", "USA", domain.NatureBusiness),
		},
		UpdatedAt: time.Date(2026, 6, 10, 12, 34, 56, 0, time.UTC),
	},
	"c4": {
		ID:          "c4",
		DisplayName: "Dave Jones",
		FirstName:   "Dave",
		LastName:    "Jones",
		Emails: []domain.ContactEmailAddress{
			domain.NewContactEmailAddress("dave.jones@example.org", domain.NatureBusiness),
		},
		Phones: []domain.ContactPhoneNumber{
			domain.NewContactPhoneNumber("+1-555-0104", domain.NatureBusiness),
		},
		Organization: "Example Corp",
		JobTitle:     "DevOps Lead",
		Addresses:    []domain.ContactPostalAddress{},
		UpdatedAt:    time.Date(2026, 6, 11, 12, 34, 56, 0, time.UTC),
	},
	"c5": {
		ID:          "c5",
		DisplayName: "Dr. Sarah Klein",
		FirstName:   "Sarah",
		LastName:    "Klein",
		Emails: []domain.ContactEmailAddress{
			domain.NewContactEmailAddress("sarah.klein@university.example.org", domain.NatureBusiness),
			domain.NewContactEmailAddress("s.klein@personal.example.org", domain.NatureHome),
		},
		Phones: []domain.ContactPhoneNumber{
			domain.NewContactPhoneNumber("+49-30-5550105", domain.NatureBusiness),
		},
		Organization: "Technical University of Berlin",
		JobTitle:     "Professor of Distributed Systems",
		Addresses: []domain.ContactPostalAddress{
			domain.NewContactPostalAddress("Straße des 17. Juni 135", "Berlin", "10623", "Germany", domain.NatureBusiness),
		},
		UpdatedAt: time.Date(2026, 6, 12, 12, 34, 56, 0, time.UTC),
	},
}
