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

package msgraph

import (
	"context"
	"fmt"
	"slices"

	kiota "github.com/microsoft/kiota-abstractions-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/tdrn-org/pim-mcp/internal/application"
	"github.com/tdrn-org/pim-mcp/internal/domain"
)

func (p *Provider) SearchContacts(ctx context.Context, filter domain.ContactFilter) ([]*domain.Contact, error) {
	client, err := p.graphClient(ctx)
	if err != nil {
		return nil, err
	}
	requestConfig := p.contactFilterRequestConfig(filter)
	response, err := client.Me().Contacts().Get(ctx, requestConfig)
	if err != nil {
		return nil, fmt.Errorf("search contacts Graph API failure (cause: %w)", err)
	}
	contacts := make([]*domain.Contact, 0)
	for _, responseItem := range response.GetValue() {
		contact := p.contactFromResponse(responseItem)
		if !contact.Empty() {
			contacts = append(contacts, contact)
		}
	}
	slices.SortFunc(contacts, application.ContactSortFunc)
	return contacts, nil
}

func (p *Provider) GetContact(ctx context.Context, id string) (*domain.Contact, error) {
	client, err := p.graphClient(ctx)
	if err != nil {
		return nil, err
	}
	response, err := client.Me().Contacts().ByContactId(id).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("get contact Graph API failure (cause: %w)", err)
	}
	contact := p.contactFromResponse(response)
	return contact, nil
}

func (p *Provider) contactFromResponse(model models.Contactable) *domain.Contact {
	return &domain.Contact{
		ID:           ptrString(model.GetId()),
		DisplayName:  ptrString(model.GetDisplayName()),
		FirstName:    ptrString(model.GetGivenName()),
		LastName:     ptrString(model.GetSurname()),
		Emails:       p.emailAddressesFromResponse(model),
		Phones:       p.phoneNumbersFromResponse(model),
		Organization: ptrString(model.GetCompanyName()),
		JobTitle:     ptrString(model.GetJobTitle()),
		Addresses:    p.postalAddressesFromResponse(model),
		UpdatedAt:    ptrTime(model.GetLastModifiedDateTime()),
	}
}

func (p *Provider) emailAddressesFromResponse(model models.Contactable) []domain.ContactEmailAddress {
	models := model.GetEmailAddresses()
	emails := make([]domain.ContactEmailAddress, 0, len(models))
	for index, model := range models {
		nature := domain.NatureOther
		switch index {
		case 0:
			nature = domain.NatureBusiness
		case 1:
			nature = domain.NatureHome
		}
		email := domain.NewContactEmailAddress(ptrString(model.GetAddress()), nature)
		if !email.Empty() {
			emails = append(emails, email)
		}
	}
	return emails
}

func (p *Provider) phoneNumbersFromResponse(model models.Contactable) []domain.ContactPhoneNumber {
	homeNumbers := model.GetHomePhones()
	businessNumbers := model.GetBusinessPhones()
	mobileNumber := model.GetMobilePhone()
	phones := make([]domain.ContactPhoneNumber, len(homeNumbers)+len(businessNumbers)+1)
	for _, homeNumber := range homeNumbers {
		phone := domain.NewContactPhoneNumber(homeNumber, domain.NatureHome)
		if !phone.Empty() {
			phones = append(phones, phone)
		}
	}
	for _, businessNumber := range businessNumbers {
		phone := domain.NewContactPhoneNumber(businessNumber, domain.NatureBusiness)
		if !phone.Empty() {
			phones = append(phones, phone)
		}
	}
	if mobileNumber != nil && *mobileNumber != "" {
		phone := domain.NewContactPhoneNumber(*mobileNumber, domain.NatureMobile)
		if !phone.Empty() {
			phones = append(phones, phone)
		}
	}
	return phones
}

func (p *Provider) postalAddressesFromResponse(model models.Contactable) []domain.ContactPostalAddress {
	addresses := make([]domain.ContactPostalAddress, 0)
	address := p.postalAddressFromResponse(model.GetBusinessAddress(), domain.NatureBusiness)
	if !address.Empty() {
		addresses = append(addresses, address)
	}
	address = p.postalAddressFromResponse(model.GetHomeAddress(), domain.NatureHome)
	if !address.Empty() {
		addresses = append(addresses, address)
	}
	address = p.postalAddressFromResponse(model.GetOtherAddress(), domain.NatureOther)
	if !address.Empty() {
		addresses = append(addresses, address)
	}
	return addresses
}

func (p *Provider) postalAddressFromResponse(model models.PhysicalAddressable, nature string) domain.ContactPostalAddress {
	return domain.NewContactPostalAddress(ptrString(model.GetStreet()), ptrString(model.GetCity()),
		ptrString(model.GetPostalCode()), ptrString(model.GetCountryOrRegion()), nature)
}

func (p *Provider) contactFilterRequestConfig(filter domain.ContactFilter) *users.ItemContactsRequestBuilderGetRequestConfiguration {
	search, limit := standardFilterPtr(filter.StandardFilter)
	headers := &kiota.RequestHeaders{}
	headers.Add("ConsistencyLevel", "eventual")
	requestConfig := &users.ItemContactsRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemContactsRequestBuilderGetQueryParameters{
			Search: search,
			Top:    limit,
			Count:  boolPtr(true),
		},
		Headers: headers,
	}
	return requestConfig
}
