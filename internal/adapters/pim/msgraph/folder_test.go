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
	"log/slog"
	"slices"
	"strings"
	"testing"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/tdrn-org/pim-mcp/internal/domain"
)

func testMailFolder(id, displayName string, childFolderCount int) models.MailFolderable {
	folder := models.NewMailFolder()
	folder.SetId(stringPtr(id))
	folder.SetDisplayName(stringPtr(displayName))
	folder.SetChildFolderCount(int32Ptr(childFolderCount))
	return folder
}

func testMessage(id string) models.Messageable {
	message := models.NewMessage()
	message.SetId(stringPtr(id))
	return message
}

func testLogger() *slog.Logger {
	return slog.New(slog.DiscardHandler)
}

// fakeMailFolderClient serves a static folder hierarchy and message set.
type fakeMailFolderClient struct {
	children           map[string][]models.MailFolderable
	messagesByFolder   map[string][]models.Messageable
	childFoldersErr    error
	messagesErr        error
	childFolderQueries []string
	messageQueries     []string
}

func (f *fakeMailFolderClient) childFolders(_ context.Context, parentID string) ([]models.MailFolderable, error) {
	f.childFolderQueries = append(f.childFolderQueries, parentID)
	if f.childFoldersErr != nil {
		return nil, f.childFoldersErr
	}
	return f.children[parentID], nil
}

func (f *fakeMailFolderClient) messages(_ context.Context, folderID string, _ domain.EmailFilter) ([]models.Messageable, error) {
	f.messageQueries = append(f.messageQueries, folderID)
	if f.messagesErr != nil {
		return nil, f.messagesErr
	}
	return f.messagesByFolder[folderID], nil
}

// newFakeMailFolderClient serves the folder hierarchy of a real mailbox:
//
//	Posteingang -> Admin, Info-Mails, Projekte
//	Ablage      -> Belege
//	Archiv (root), Trading
func newFakeMailFolderClient() *fakeMailFolderClient {
	return &fakeMailFolderClient{
		children: map[string][]models.MailFolderable{
			rootMailFolderID: {
				testMailFolder("inbox", "Posteingang", 3),
				testMailFolder("archive-root", "Archiv", 0),
				testMailFolder("trading", "Trading", 0),
				testMailFolder("ablage", "Ablage", 1),
			},
			"inbox": {
				testMailFolder("inbox-admin", "Admin", 0),
				testMailFolder("inbox-info", "Info-Mails", 0),
				testMailFolder("inbox-projects", "Projekte", 0),
				testMailFolder("inbox-archive", "Archiv", 0),
			},
			"ablage": {
				testMailFolder("ablage-belege", "Belege", 0),
			},
		},
		messagesByFolder: map[string][]models.Messageable{
			"inbox-admin":   {testMessage("admin-1"), testMessage("admin-2")},
			"archive-root":  {testMessage("archive-1")},
			"inbox-archive": {testMessage("archive-2")},
			"trading":       {testMessage("trading-1")},
		},
	}
}

func TestMatchingFolderIDs(t *testing.T) {
	folders := []models.MailFolderable{
		testMailFolder("root-archive", "Archiv", 0),
		testMailFolder("inbox", "Posteingang", 4),
		testMailFolder("inbox-admin", "Admin", 0),
		testMailFolder("inbox-archive", "Archiv", 0),
		testMailFolder("sent-archive", "archiv", 0),
		testMailFolder("no-name", "", 0),
	}
	unnamedFolder := models.NewMailFolder()
	unnamedFolder.SetId(stringPtr("nil-name"))
	folders = append(folders, unnamedFolder)
	tests := []struct {
		name        string
		displayName string
		expectedIDs []string
	}{
		{"unique folder", "Admin", []string{"inbox-admin"}},
		{"case insensitive", "admin", []string{"inbox-admin"}},
		{"all duplicates at any hierarchy level", "Archiv", []string{"root-archive", "inbox-archive", "sent-archive"}},
		{"no match", "Agents", []string{}},
		{"empty name matches nothing", "", []string{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actualIDs := matchingFolderIDs(folders, test.displayName)
			if !slices.Equal(actualIDs, test.expectedIDs) {
				t.Fatalf("matchingFolderIDs(%q) = %v, expected %v", test.displayName, actualIDs, test.expectedIDs)
			}
		})
	}
}

func TestListAllFolders(t *testing.T) {
	client := newFakeMailFolderClient()
	folders, err := listAllFolders(context.Background(), client)
	if err != nil {
		t.Fatalf("listAllFolders() failed: %v", err)
	}
	expectedIDs := []string{
		"inbox", "archive-root", "trading", "ablage",
		"inbox-admin", "inbox-info", "inbox-projects", "inbox-archive",
		"ablage-belege",
	}
	actualIDs := make([]string, 0, len(folders))
	for _, folder := range folders {
		actualIDs = append(actualIDs, ptrString(folder.GetId()))
	}
	if !slices.Equal(actualIDs, expectedIDs) {
		t.Fatalf("listAllFolders() = %v, expected %v", actualIDs, expectedIDs)
	}
	if !slices.Contains(client.childFolderQueries, "ablage") {
		t.Fatalf("listAllFolders() did not descend into 'ablage': %v", client.childFolderQueries)
	}
}

func TestListAllFoldersChildFoldersFailure(t *testing.T) {
	client := newFakeMailFolderClient()
	client.childFoldersErr = errors.New("graph is unhappy")
	folders, err := listAllFolders(context.Background(), client)
	if err == nil {
		t.Fatalf("listAllFolders() = %v, expected an error", folders)
	}
}

func TestSearchEmailsInMatchingFoldersSubFolder(t *testing.T) {
	client := newFakeMailFolderClient()
	items, err := searchEmailsInMatchingFolders(context.Background(), client, "Admin", domain.EmailFilter{}, testLogger())
	if err != nil {
		t.Fatalf("searchEmailsInMatchingFolders() failed: %v", err)
	}
	if !slices.Equal(client.messageQueries, []string{"inbox-admin"}) {
		t.Fatalf("message queries = %v, expected only the Admin sub-folder", client.messageQueries)
	}
	if len(items) != 2 {
		t.Fatalf("searchEmailsInMatchingFolders() returned %d messages, expected 2", len(items))
	}
}

func TestSearchEmailsInMatchingFoldersDuplicateNames(t *testing.T) {
	client := newFakeMailFolderClient()
	items, err := searchEmailsInMatchingFolders(context.Background(), client, "Archiv", domain.EmailFilter{}, testLogger())
	if err != nil {
		t.Fatalf("searchEmailsInMatchingFolders() failed: %v", err)
	}
	expectedQueries := []string{"archive-root", "inbox-archive"}
	if !slices.Equal(client.messageQueries, expectedQueries) {
		t.Fatalf("message queries = %v, expected %v", client.messageQueries, expectedQueries)
	}
	if len(items) != 2 {
		t.Fatalf("searchEmailsInMatchingFolders() returned %d messages, expected 2", len(items))
	}
}

func TestSearchEmailsInMatchingFoldersUnknownName(t *testing.T) {
	client := newFakeMailFolderClient()
	items, err := searchEmailsInMatchingFolders(context.Background(), client, "Agents", domain.EmailFilter{}, testLogger())
	if err != nil {
		t.Fatalf("searchEmailsInMatchingFolders() failed: %v", err)
	}
	if len(client.messageQueries) != 0 {
		t.Fatalf("message queries = %v, expected none for an unknown folder name", client.messageQueries)
	}
	if len(items) != 0 {
		t.Fatalf("searchEmailsInMatchingFolders() returned %d messages, expected none", len(items))
	}
}

func TestSearchEmailsInMatchingFoldersMessagesFailure(t *testing.T) {
	client := newFakeMailFolderClient()
	client.messagesErr = errors.New("graph is unhappy")
	items, err := searchEmailsInMatchingFolders(context.Background(), client, "Admin", domain.EmailFilter{}, testLogger())
	if err == nil {
		t.Fatalf("searchEmailsInMatchingFolders() = %v, expected an error", items)
	}
}

func TestCollectFoldersDepthLimit(t *testing.T) {
	// A folder reporting itself as its own child must not be followed forever.
	client := &fakeMailFolderClient{
		children: map[string][]models.MailFolderable{
			rootMailFolderID: {testMailFolder("loop", "Loop", 1)},
			"loop":           {testMailFolder("loop", "Loop", 1)},
		},
	}
	folders := make([]models.MailFolderable, 0)
	err := collectFolders(context.Background(), client, rootMailFolderID, 0, &folders)
	if err == nil {
		t.Fatalf("collectFolders() = %d folders, expected a depth error", len(folders))
	}
	if !strings.Contains(err.Error(), "maximum depth") {
		t.Fatalf("collectFolders() error = %v, expected a maximum depth error", err)
	}
}
