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
	"errors"
	"strings"
	"testing"
	"time"

	abstractions "github.com/microsoft/kiota-abstractions-go"
	"github.com/microsoft/kiota-abstractions-go/serialization"
	"github.com/microsoft/kiota-abstractions-go/store"
	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/tdrn-org/pim-mcp/internal/domain"
)

// fakeRequestAdapter answers the Graph requests issued by graphMailFolderClient
// with canned responses and records the requested url templates.
type fakeRequestAdapter struct {
	sendErr               error
	mailFolderResponse    models.MailFolderCollectionResponseable
	messageResponse       models.MessageCollectionResponseable
	requestedURLTemplates []string
}

func (a *fakeRequestAdapter) Send(_ context.Context, requestInfo *abstractions.RequestInformation, _ serialization.ParsableFactory, _ abstractions.ErrorMappings) (serialization.Parsable, error) {
	a.requestedURLTemplates = append(a.requestedURLTemplates, requestInfo.UrlTemplate)
	if a.sendErr != nil {
		return nil, a.sendErr
	}
	switch {
	case strings.Contains(requestInfo.UrlTemplate, "/childFolders"):
		return a.mailFolderResponse, nil
	case strings.Contains(requestInfo.UrlTemplate, "/messages"):
		return a.messageResponse, nil
	case strings.Contains(requestInfo.UrlTemplate, "/mailFolders"):
		return a.mailFolderResponse, nil
	}
	return nil, nil
}

func (a *fakeRequestAdapter) SendCollection(context.Context, *abstractions.RequestInformation, serialization.ParsableFactory, abstractions.ErrorMappings) ([]serialization.Parsable, error) {
	return nil, a.sendErr
}

func (a *fakeRequestAdapter) SendEnum(context.Context, *abstractions.RequestInformation, serialization.EnumFactory, abstractions.ErrorMappings) (any, error) {
	return nil, a.sendErr
}

func (a *fakeRequestAdapter) SendEnumCollection(context.Context, *abstractions.RequestInformation, serialization.EnumFactory, abstractions.ErrorMappings) ([]any, error) {
	return nil, a.sendErr
}

func (a *fakeRequestAdapter) SendPrimitive(context.Context, *abstractions.RequestInformation, string, abstractions.ErrorMappings) (any, error) {
	return nil, a.sendErr
}

func (a *fakeRequestAdapter) SendPrimitiveCollection(context.Context, *abstractions.RequestInformation, string, abstractions.ErrorMappings) ([]any, error) {
	return nil, a.sendErr
}

func (a *fakeRequestAdapter) SendNoContent(context.Context, *abstractions.RequestInformation, abstractions.ErrorMappings) error {
	return a.sendErr
}

func (a *fakeRequestAdapter) GetSerializationWriterFactory() serialization.SerializationWriterFactory {
	return nil
}

func (a *fakeRequestAdapter) EnableBackingStore(store.BackingStoreFactory) {}

func (a *fakeRequestAdapter) SetBaseUrl(string) {}

func (a *fakeRequestAdapter) GetBaseUrl() string {
	return "https://graph.microsoft.com/v1.0"
}

func (a *fakeRequestAdapter) ConvertToNativeRequest(context.Context, *abstractions.RequestInformation) (any, error) {
	return nil, a.sendErr
}

func newMailFolderCollection(folders ...models.MailFolderable) models.MailFolderCollectionResponseable {
	response := models.NewMailFolderCollectionResponse()
	response.SetValue(folders)
	return response
}

func newMessageCollection(messages ...models.Messageable) models.MessageCollectionResponseable {
	response := models.NewMessageCollectionResponse()
	response.SetValue(messages)
	return response
}

func newFakeGraphMailFolderClient(adapter *fakeRequestAdapter) *graphMailFolderClient {
	return &graphMailFolderClient{graph: msgraphsdkgo.NewGraphServiceClient(adapter)}
}

func TestGraphMailFolderClientChildFoldersRootLevel(t *testing.T) {
	adapter := &fakeRequestAdapter{
		mailFolderResponse: newMailFolderCollection(
			testMailFolder("inbox", "Posteingang", 4),
			testMailFolder("trading", "Trading", 0),
		),
	}
	client := newFakeGraphMailFolderClient(adapter)
	folders, err := client.childFolders(context.Background(), rootMailFolderID)
	if err != nil {
		t.Fatalf("childFolders() failed: %v", err)
	}
	if len(folders) != 2 {
		t.Fatalf("childFolders() returned %d folders, expected 2", len(folders))
	}
	if len(adapter.requestedURLTemplates) != 1 || strings.Contains(adapter.requestedURLTemplates[0], "/childFolders") {
		t.Fatalf("childFolders(%q) requested %v, expected the root folder level", rootMailFolderID, adapter.requestedURLTemplates)
	}
}

func TestGraphMailFolderClientChildFoldersSubFolder(t *testing.T) {
	adapter := &fakeRequestAdapter{
		mailFolderResponse: newMailFolderCollection(testMailFolder("inbox-admin", "Admin", 0)),
	}
	client := newFakeGraphMailFolderClient(adapter)
	folders, err := client.childFolders(context.Background(), "inbox")
	if err != nil {
		t.Fatalf("childFolders() failed: %v", err)
	}
	if len(folders) != 1 || ptrString(folders[0].GetId()) != "inbox-admin" {
		t.Fatalf("childFolders() = %v, expected the single Admin folder", folders)
	}
	if len(adapter.requestedURLTemplates) != 1 || !strings.Contains(adapter.requestedURLTemplates[0], "/childFolders") {
		t.Fatalf("childFolders(%q) requested %v, expected the childFolders endpoint", "inbox", adapter.requestedURLTemplates)
	}
}

func TestGraphMailFolderClientChildFoldersFailure(t *testing.T) {
	adapter := &fakeRequestAdapter{sendErr: errors.New("graph is unhappy")}
	client := newFakeGraphMailFolderClient(adapter)
	folders, err := client.childFolders(context.Background(), rootMailFolderID)
	if err == nil {
		t.Fatalf("childFolders() = %v, expected an error", folders)
	}
}

func TestGraphMailFolderClientMessages(t *testing.T) {
	adapter := &fakeRequestAdapter{
		messageResponse: newMessageCollection(testMessage("mail-1"), testMessage("mail-2")),
	}
	client := newFakeGraphMailFolderClient(adapter)
	messages, err := client.messages(context.Background(), "inbox-admin", domain.EmailFilter{})
	if err != nil {
		t.Fatalf("messages() failed: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("messages() returned %d messages, expected 2", len(messages))
	}
	if len(adapter.requestedURLTemplates) != 1 || !strings.Contains(adapter.requestedURLTemplates[0], "/messages") {
		t.Fatalf("messages() requested %v, expected the messages endpoint", adapter.requestedURLTemplates)
	}
}

func TestGraphMailFolderClientMessagesFailure(t *testing.T) {
	adapter := &fakeRequestAdapter{sendErr: errors.New("graph is unhappy")}
	client := newFakeGraphMailFolderClient(adapter)
	messages, err := client.messages(context.Background(), "inbox-admin", domain.EmailFilter{})
	if err == nil {
		t.Fatalf("messages() = %v, expected an error", messages)
	}
}

func TestMailFolderMessagesRequestConfig(t *testing.T) {
	since := time.Date(2026, 11, 7, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name           string
		filter         domain.EmailFilter
		expectedFilter string
	}{
		{"no filter", domain.EmailFilter{}, ""},
		{"unread only", domain.EmailFilter{UnreadOnly: true}, "(isRead eq false)"},
		{
			"unread only and since",
			domain.EmailFilter{UnreadOnly: true, Since: &since},
			"(isRead eq false) and (receivedDateTime ge 2026-11-07T00:00:00Z)",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			requestConfig := mailFolderMessagesRequestConfig(test.filter)
			actualFilter := ""
			if requestConfig.QueryParameters.Filter != nil {
				actualFilter = *requestConfig.QueryParameters.Filter
			}
			if actualFilter != test.expectedFilter {
				t.Fatalf("filter = %q, expected %q", actualFilter, test.expectedFilter)
			}
		})
	}
}
