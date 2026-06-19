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
	"strings"
	"time"

	kiota "github.com/microsoft/kiota-abstractions-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/tdrn-org/pim-mcp/internal/application"
	"github.com/tdrn-org/pim-mcp/internal/domain"
)

func (p *Provider) SearchEmails(ctx context.Context, filter domain.EmailFilter) ([]*domain.Email, error) {
	client, err := p.graphClient(ctx)
	if err != nil {
		return nil, err
	}
	var response models.MessageCollectionResponseable
	if filter.Folder != nil && *filter.Folder != "" {
		requestConfig := p.emailFolderFilterRequestConfig(filter)
		response, err = client.Me().MailFolders().ByMailFolderId(*filter.Folder).Messages().Get(ctx, requestConfig)
	} else {
		requestConfig := p.emailFilterRequestConfig(filter)
		response, err = client.Me().Messages().Get(ctx, requestConfig)
	}
	if err != nil {
		return nil, fmt.Errorf("search emails Graph API failure (cause: %w)", err)
	}
	emails := make([]*domain.Email, 0)
	for _, responseItem := range response.GetValue() {
		email := p.emailFromResponse(responseItem)
		if !email.Empty() {
			emails = append(emails, email)
		}
	}
	slices.SortFunc(emails, application.EmailSortFunc)
	return emails, nil
}

func (p *Provider) GetEmail(ctx context.Context, id string) (*domain.Email, error) {
	client, err := p.graphClient(ctx)
	if err != nil {
		return nil, err
	}
	response, err := client.Me().Messages().ByMessageId(id).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("get email Graph API failure (cause: %w)", err)
	}
	email := p.emailFromResponse(response)
	return email, nil
}

func (p *Provider) UpdateEmail(ctx context.Context, id string, update domain.EmailUpdate) (*domain.Email, error) {
	requestBuilder := mailRequestBuilder{Request: models.NewMessage()}
	request := requestBuilder.
		IsRead(update.IsRead).
		Request
	client, err := p.graphClient(ctx)
	if err != nil {
		return nil, err
	}
	response, err := client.Me().Messages().Post(ctx, request, nil)
	if err != nil {
		return nil, fmt.Errorf("update email Graph API failure (cause: %w)", err)
	}
	email := p.emailFromResponse(response)
	return email, nil
}

func (p *Provider) emailFromResponse(model models.Messageable) *domain.Email {
	body := model.GetUniqueBody()
	if body == nil {
		body = model.GetBody()
	}
	content := ""
	if body != nil {
		content = *body.GetContent()
	}
	return &domain.Email{
		ID:         ptrString(model.GetId()),
		Subject:    ptrString(model.GetSubject()),
		Body:       content,
		From:       p.addressFromResponse(model.GetFrom()),
		To:         p.addressesFromResponse(model.GetToRecipients()),
		CC:         p.addressesFromResponse(model.GetCcRecipients()),
		ReceivedAt: ptrTime(model.GetReceivedDateTime()),
		SentAt:     ptrTime(model.GetSentDateTime()),
		IsRead:     ptrBool(model.GetIsRead(), true),
		ThreadID:   ptrString(model.GetConversationId()),
	}
}

func (p *Provider) addressesFromResponse(models []models.Recipientable) []domain.NamedEmailAddress {
	addresses := make([]domain.NamedEmailAddress, 0, len(models))
	for _, model := range models {
		addresses = append(addresses, p.addressFromResponse(model))
	}
	return addresses
}

func (p *Provider) addressFromResponse(model models.Recipientable) domain.NamedEmailAddress {
	emailAddress := model.GetEmailAddress()
	return domain.NewNamedEmailAddress(ptrString(emailAddress.GetAddress()), ptrString(emailAddress.GetName()))
}

func (p *Provider) emailFolderFilterRequestConfig(filter domain.EmailFilter) *users.ItemMailFoldersItemMessagesRequestBuilderGetRequestConfiguration {
	search, limit := standardFilterPtr(filter.StandardFilter)
	filterParts := make([]string, 0)
	if filter.UnreadOnly {
		filterParts = append(filterParts, "(isRead eq false)")
	}
	if filter.Since != nil && !filter.Since.IsZero() {
		formattedDate := filter.Since.UTC().Format(time.RFC3339)
		filterParts = append(filterParts, fmt.Sprintf("(receivedDateTime ge %s)", formattedDate))
	}
	var filterParam *string
	if len(filterParts) > 0 {
		filterParam = stringPtr(strings.Join(filterParts, " and "))
	}
	headers := &kiota.RequestHeaders{}
	headers.Add("ConsistencyLevel", "eventual")
	requestConfig := &users.ItemMailFoldersItemMessagesRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemMailFoldersItemMessagesRequestBuilderGetQueryParameters{
			Search: search,
			Filter: filterParam,
			Top:    limit,
			Count:  boolPtr(true),
		},
		Headers: headers,
	}
	return requestConfig
}

func (p *Provider) emailFilterRequestConfig(filter domain.EmailFilter) *users.ItemMessagesRequestBuilderGetRequestConfiguration {
	search, limit := standardFilterPtr(filter.StandardFilter)
	filterParts := make([]string, 0)
	if filter.UnreadOnly {
		filterParts = append(filterParts, "(isRead eq false)")
	}
	if filter.Since != nil && !filter.Since.IsZero() {
		formattedDate := filter.Since.UTC().Format(time.RFC3339)
		filterParts = append(filterParts, fmt.Sprintf("(receivedDateTime ge %s)", formattedDate))
	}
	var filterParam *string
	if len(filterParts) > 0 {
		filterParam = stringPtr(strings.Join(filterParts, " and "))
	}
	headers := &kiota.RequestHeaders{}
	headers.Add("ConsistencyLevel", "eventual")
	headers.Add("Prefer", "outlook.body-content-type=\"text\"")
	requestConfig := &users.ItemMessagesRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemMessagesRequestBuilderGetQueryParameters{
			Search: search,
			Filter: filterParam,
			Top:    limit,
			Count:  boolPtr(true),
		},
		Headers: headers,
	}
	return requestConfig
}

type mailRequestBuilder struct {
	Request *models.Message
}

func (b *mailRequestBuilder) IsRead(isRead *bool) *mailRequestBuilder {
	if isRead != nil {
		b.Request.SetIsRead(isRead)
	}
	return b
}
