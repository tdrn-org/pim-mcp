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

func addEmailTools(server *mcp.Server, provider domain.EmailProvider) {
	addSearchEmailsTool(server, provider)
	addGetEmailTool(server, provider)
}

func addSearchEmailsTool(server *mcp.Server, provider domain.EmailProvider) {
	tool := &mcp.Tool{
		Name:        "searchEmails",
		Description: "Searches for emails using the given search parameters. An email summary including the email ID is returned for every found email. The email ID can be used to get the full email details (getEmail).",
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, params *SearchEmailsParams) (*mcp.CallToolResult, any, error) {
		filter := domain.EmailFilter{
			Query:      params.Query,
			Limit:      params.Limit,
			UnreadOnly: params.UnreadOnly,
			Folder:     params.Folder,
			Since:      params.Since,
		}
		emails, err := provider.SearchEmails(ctx, filter)
		if err != nil {
			return nil, nil, err
		}
		return nil, toEmailSummaryOutputs(emails), nil
	}
	mcp.AddTool(server, tool, handler)
}

func addGetEmailTool(server *mcp.Server, provider domain.EmailProvider) {
	tool := &mcp.Tool{
		Name:        "getEmail",
		Description: "Gets the full email details for the given ID",
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, params *GetEmailParams) (*mcp.CallToolResult, any, error) {
		email, err := provider.GetEmail(ctx, params.ID)
		if err != nil {
			return nil, nil, err
		}
		return nil, toEmailOutput(email), nil
	}
	mcp.AddTool(server, tool, handler)
}

type SearchEmailsParams struct {
	Query      string     `json:"query,omitempty"       jsonschema:"Term to search for. All email attributes (Subject, Body, From Name, From Address, To Names, To Addresses) are matched against this term (substring match). As soon as one attribute matches, the email is included in the result."`
	Limit      int        `json:"limit,omitempty"       jsonschema:"The maximum number of emails to return. If no limit is given a provider specific one applies."`
	UnreadOnly bool       `json:"unread_only,omitempty" jsonschema:"If true only unread emails are returned. Defaults to false."`
	Folder     *string    `json:"folder,omitempty"      jsonschema:"The folder to search in. Defaults to 'inbox'."`
	Since      *time.Time `json:"since,omitempty"       jsonschema:"Only return emails received at or after this time. Use RFC3339 format (e.g. 2026-06-07T00:00:00Z). Defaults to 30 days ago."`
}

type GetEmailParams struct {
	ID string `json:"id" jsonschema:"ID of the email to return."`
}

type EmailSummaryOutput struct {
	ID         string          `json:"id" jsonschema:"ID of the email."`
	Subject    string          `json:"subject" jsonschema:"The subject of the email"`
	From       AddressOutput   `json:"from" jsonschema:"The sender address of the email"`
	To         []AddressOutput `json:"tos" jsonschema:"The TO: receiver addresses of the email"`
	CC         []AddressOutput `json:"ccs" jsonschema:"The CC: receiver addresses of the email"`
	ReceivedAt time.Time       `json:"received_at" jsonschema:"The receive time of the email"`
	SentAt     time.Time       `json:"sent_at" jsonschema:"The sent time of the email"`
	IsRead     bool            `json:"is_read" jsonschema:"The read status of the email"`
}

type EmailOutput struct {
	ID         string          `json:"id" jsonschema:"ID of the email."`
	Subject    string          `json:"subject" jsonschema:"The subject of the email"`
	Body       string          `json:"body" jsonschema:"The body of the email"`
	From       AddressOutput   `json:"from" jsonschema:"The sender address of the email"`
	To         []AddressOutput `json:"tos" jsonschema:"The TO: receiver addresses of the email"`
	CC         []AddressOutput `json:"ccs" jsonschema:"The CC: receiver addresses of the email"`
	ReceivedAt time.Time       `json:"received_at" jsonschema:"The receive time of the email"`
	SentAt     time.Time       `json:"sent_at" jsonschema:"The sent time of the email"`
	IsRead     bool            `json:"is_read" jsonschema:"The read status of the email"`
	ThreadID   string          `json:"thread_id" jsonschema:"The thread ID of the email"`
}

type AddressOutput struct {
	Name    string `json:"name" jsonschema:"The display name of this address."`
	Address string `json:"address" jsonschema:"The email address of this address."`
}

func toEmailSummaryOutputs(emails []*domain.Email) []*EmailSummaryOutput {
	outputs := make([]*EmailSummaryOutput, 0, len(emails))
	for _, email := range emails {
		output := &EmailSummaryOutput{
			ID:         email.ID,
			Subject:    email.Subject,
			From:       toAddressOutput(email.From),
			To:         toAddressOutputs(email.To),
			CC:         toAddressOutputs(email.CC),
			ReceivedAt: email.ReceivedAt,
			SentAt:     email.SentAt,
			IsRead:     email.IsRead,
		}
		outputs = append(outputs, output)
	}
	return outputs
}

func toEmailOutput(email *domain.Email) *EmailOutput {
	output := &EmailOutput{
		ID:         email.ID,
		Subject:    email.Subject,
		Body:       email.Body,
		From:       toAddressOutput(email.From),
		To:         toAddressOutputs(email.To),
		CC:         toAddressOutputs(email.CC),
		ReceivedAt: email.ReceivedAt,
		SentAt:     email.SentAt,
		IsRead:     email.IsRead,
		ThreadID:   email.ThreadID,
	}
	return output
}

func toAddressOutput(address domain.Address) AddressOutput {
	return AddressOutput{
		Name:    address.Name,
		Address: address.Address,
	}
}

func toAddressOutputs(addresses []domain.Address) []AddressOutput {
	outputs := make([]AddressOutput, 0, len(addresses))
	for _, address := range addresses {
		outputs = append(outputs, toAddressOutput(address))
	}
	return outputs
}
