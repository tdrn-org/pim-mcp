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
	"12": {
		ID:           "12",
		DisplayName:  "John Smith",
		FirstName:    "John",
		LastName:     "Smith",
		Emails:       []domain.ContactEmailAddress{domain.NewContactEmailAddress("john.smith@nowhere.org", "private")},
		Phones:       []domain.ContactPhoneNumber{domain.NewContactPhoneNumber("+0123456789", "private")},
		Organization: "",
		JobTitle:     "",
		Addresses:    []domain.ContactPostalAddress{},
		UpdatedAt:    time.Date(2026, 06, 8, 12, 34, 56, 0, time.UTC),
	},
	"34": {
		ID:           "34",
		DisplayName:  "Jane Miller",
		FirstName:    "Jane",
		LastName:     "Miller",
		Emails:       []domain.ContactEmailAddress{domain.NewContactEmailAddress("jane.miller@nowhere.org", "private")},
		Phones:       []domain.ContactPhoneNumber{domain.NewContactPhoneNumber("+0123456789", "private")},
		Organization: "",
		JobTitle:     "",
		Addresses:    []domain.ContactPostalAddress{},
		UpdatedAt:    time.Date(2026, 06, 9, 12, 34, 56, 0, time.UTC),
	},
	"56": {
		ID:           "56",
		DisplayName:  "The Other",
		FirstName:    "The",
		LastName:     "Other",
		Emails:       []domain.ContactEmailAddress{domain.NewContactEmailAddress("the.other@nowhere.org", "private")},
		Phones:       []domain.ContactPhoneNumber{domain.NewContactPhoneNumber("+0123456789", "private")},
		Organization: "",
		JobTitle:     "",
		Addresses:    []domain.ContactPostalAddress{},
		UpdatedAt:    time.Date(2026, 06, 8, 12, 34, 56, 0, time.UTC),
	},
}
