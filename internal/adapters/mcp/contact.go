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
	"context"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tdrn-org/pim-mcp/internal/domain"
)

func addContactTools(server *mcp.Server, provider domain.ContactProvider) {
	addSearchContactsTool(server, provider)
	addGetContactTool(server, provider)
}

func addSearchContactsTool(server *mcp.Server, provider domain.ContactProvider) {
	tool := &mcp.Tool{
		Name:        "searchContacts",
		Description: "Searches for contacts using the given search parameters. A contact summary including the contact ID is returned for every found contact. The contact ID can be used to get the full contact details (getContact).",
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, params *SearchContactsParams) (*mcp.CallToolResult, any, error) {
		filter := domain.ContactFilter{
			Query: params.Query,
			Limit: params.Limit,
		}
		contacts, err := provider.SearchContacts(ctx, filter)
		if err != nil {
			return nil, nil, err
		}
		return nil, toContactSummaryOutputs(contacts), nil
	}
	mcp.AddTool(server, tool, handler)

}

func addGetContactTool(server *mcp.Server, provider domain.ContactProvider) {
	tool := &mcp.Tool{
		Name:        "getContact",
		Description: "Gets the full contact details for the given ID",
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, params *GetContactParams) (*mcp.CallToolResult, any, error) {
		contact, err := provider.GetContact(ctx, params.ID)
		if err != nil {
			return nil, nil, err
		}
		return nil, toContactOutput(contact), nil
	}
	mcp.AddTool(server, tool, handler)

}

type SearchContactsParams struct {
	Query string `json:"query,omitempty" jsonschema:"Term to search for. All contact attributes (Display Name, First Name, Last Name, Email Address, Phone Number, Address) are matched against this term (substring match). As soon as one attribute matches, the contact is included in the result. Leave empty to list all contacts."`
	Limit int    `json:"limit,omitempty" jsonschema:"The maximum number of contacts to return. If no limit is given a provider specific one applies."`
}

type GetContactParams struct {
	ID string `json:"id" jsonschema:"ID of the contact to return."`
}

type ContactSummaryOutput struct {
	ID          string               `json:"id" jsonschema:"ID of the contact."`
	DisplayName string               `json:"display_name" jsonschema:"The display name of the contact."`
	FirstName   string               `json:"first_name" jsonschema:"The first name of the contact."`
	LastName    string               `json:"last_name" jsonschema:"The last name of the contact."`
	Emails      []EmailAddressOutput `json:"emails" jsonschema:"The email addresses of the contact."`
	Phones      []PhoneNumberOutput  `json:"phones" jsonschema:"The phone numbers of the contact."`
}

type ContactOutput struct {
	ID           string                `json:"id" jsonschema:"ID of the contact."`
	DisplayName  string                `json:"display_name" jsonschema:"The display name of the contact."`
	FirstName    string                `json:"first_name" jsonschema:"The first name of the contact."`
	LastName     string                `json:"last_name" jsonschema:"The last name of the contact."`
	Emails       []EmailAddressOutput  `json:"emails" jsonschema:"The email addresses of the contact."`
	Phones       []PhoneNumberOutput   `json:"phones" jsonschema:"The phone numbers of the contact."`
	Organization string                `json:"organization" jsonschema:"The organization unit of the contact"`
	JobTitle     string                `json:"job_title" jsonschema:"The job title of the contact"`
	Addresses    []PostalAddressOutput `json:"addresses" jsonschema:"The postal addresses of the contact"`
	UpdatedAt    time.Time             `json:"updated_at" jsonschema:"The last time the contact was updated."`
}

type EmailAddressOutput struct {
	Address string `json:"address" jsonschema:"The email address."`
	Nature  string `json:"nature" jsonschema:"The nature of the email address (private, work, ...)."`
}

type PhoneNumberOutput struct {
	Number string `json:"number" jsonschema:"The phone number."`
	Nature string `json:"nature" jsonschema:"The nature of the phone number (private, work, ...)."`
}

type PostalAddressOutput struct {
	Street     string `json:"street" jsonschema:"The address' street."`
	City       string `json:"city" jsonschema:"The address' city."`
	PostalCode string `json:"postal_code" jsonschema:"The address' postal code."`
	Country    string `json:"country" jsonschema:"The address' country."`
	Nature     string `json:"nature" jsonschema:"The nature of the postal address (private, work, ...)."`
}

func toContactSummaryOutputs(contacts []*domain.Contact) []*ContactSummaryOutput {
	outputs := make([]*ContactSummaryOutput, 0, len(contacts))
	for _, contact := range contacts {
		output := &ContactSummaryOutput{
			ID:          contact.ID,
			DisplayName: contact.DisplayName,
			FirstName:   contact.FirstName,
			LastName:    contact.LastName,
			Emails:      toEmailAddressOutputs(contact.Emails),
			Phones:      toPhoneNumberOutputs(contact.Phones),
		}
		outputs = append(outputs, output)
	}
	return outputs
}

func toContactOutput(contact *domain.Contact) *ContactOutput {
	output := &ContactOutput{
		ID:           contact.ID,
		DisplayName:  contact.DisplayName,
		FirstName:    contact.FirstName,
		LastName:     contact.LastName,
		Emails:       toEmailAddressOutputs(contact.Emails),
		Phones:       toPhoneNumberOutputs(contact.Phones),
		Organization: contact.Organization,
		JobTitle:     contact.JobTitle,
		Addresses:    toPostalAddressOutputs(contact.Addresses),
		UpdatedAt:    contact.UpdatedAt,
	}
	return output
}

func toEmailAddressOutputs(emails domain.EmailAddresses) []EmailAddressOutput {
	outputs := make([]EmailAddressOutput, 0, len(emails))
	for _, email := range emails {
		output := EmailAddressOutput{
			Address: email.Address,
			Nature:  email.Nature,
		}
		outputs = append(outputs, output)
	}
	return outputs
}

func toPhoneNumberOutputs(phones domain.PhoneNumbers) []PhoneNumberOutput {
	outputs := make([]PhoneNumberOutput, 0, len(phones))
	for _, phone := range phones {
		output := PhoneNumberOutput{
			Number: phone.Number,
			Nature: phone.Nature,
		}
		outputs = append(outputs, output)
	}
	return outputs
}

func toPostalAddressOutputs(addresses domain.PostalAddresses) []PostalAddressOutput {
	outputs := make([]PostalAddressOutput, 0, len(addresses))
	for _, address := range addresses {
		output := PostalAddressOutput{
			Street:     address.Street,
			City:       address.City,
			PostalCode: address.PostalCode,
			Country:    address.Country,
			Nature:     address.Nature,
		}
		outputs = append(outputs, output)
	}
	return outputs
}
