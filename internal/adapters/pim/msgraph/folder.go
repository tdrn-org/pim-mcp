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
	"log/slog"
	"strings"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/tdrn-org/pim-mcp/internal/domain"
)

const (
	// maxMailFolderDepth limits the folder hierarchy traversal. It is a safety net
	// against cyclic or unexpectedly deep folder structures, not an expected limit.
	maxMailFolderDepth = 10
	// mailFolderPageSize is the number of folders requested per hierarchy level.
	// Mailboxes with more sibling folders than this are not fully traversed.
	mailFolderPageSize = 100
	// rootMailFolderID is the pseudo folder id addressing the root folder level.
	rootMailFolderID = ""
)

// mailFolderClient provides the mail folder access needed to search emails by
// folder name. It exists to keep the folder traversal free of the Graph SDK.
type mailFolderClient interface {
	// childFolders lists the direct child folders of parentID ("" = root folder).
	childFolders(ctx context.Context, parentID string) ([]models.MailFolderable, error)
	// messages lists the messages of the given folder matching the filter.
	messages(ctx context.Context, folderID string, filter domain.EmailFilter) ([]models.Messageable, error)
}

// searchEmailsInMatchingFolders searches all mail folders matching folderName.
//
// A folder name is not unique: a mailbox may hold several folders sharing the
// same display name (e.g. "Archiv" at the root and as a sub-folder of the inbox),
// and all of them are searched. The folder hierarchy has to be walked explicitly
// because the Graph mail folder collection only lists the root's direct children,
// so a server side display name filter can never match a sub-folder.
func searchEmailsInMatchingFolders(ctx context.Context, client mailFolderClient, folderName string, filter domain.EmailFilter, logger *slog.Logger) ([]models.Messageable, error) {
	folders, err := listAllFolders(ctx, client)
	if err != nil {
		return nil, err
	}
	folderIDs := matchingFolderIDs(folders, folderName)
	if len(folderIDs) == 0 {
		logger.Debug("no mail folder matches folder filter", slog.String("folder", folderName))
	}
	responseItems := make([]models.Messageable, 0)
	for _, folderID := range folderIDs {
		items, err := client.messages(ctx, folderID, filter)
		if err != nil {
			return nil, err
		}
		responseItems = append(responseItems, items...)
	}
	return responseItems, nil
}

// listAllFolders returns all mail folders of the mailbox, including all sub-folders.
func listAllFolders(ctx context.Context, client mailFolderClient) ([]models.MailFolderable, error) {
	folders := make([]models.MailFolderable, 0)
	if err := collectFolders(ctx, client, rootMailFolderID, 0, &folders); err != nil {
		return nil, err
	}
	return folders, nil
}

// collectFolders appends the child folders of parentID and recursively all their
// sub-folders to folders.
func collectFolders(ctx context.Context, client mailFolderClient, parentID string, depth int, folders *[]models.MailFolderable) error {
	if depth > maxMailFolderDepth {
		return fmt.Errorf("mail folder hierarchy exceeds maximum depth of %d (parent id: '%s')", maxMailFolderDepth, parentID)
	}
	children, err := client.childFolders(ctx, parentID)
	if err != nil {
		return err
	}
	*folders = append(*folders, children...)
	for _, child := range children {
		if ptrInt32(child.GetChildFolderCount()) == 0 {
			continue
		}
		if err := collectFolders(ctx, client, ptrString(child.GetId()), depth+1, folders); err != nil {
			return err
		}
	}
	return nil
}

// matchingFolderIDs returns the IDs of all folders whose display name matches
// displayName. Matching is case insensitive, mirroring the behavior of the
// formerly used Graph display name filter. An empty name matches nothing.
func matchingFolderIDs(folders []models.MailFolderable, displayName string) []string {
	ids := make([]string, 0)
	if displayName == "" {
		return ids
	}
	for _, folder := range folders {
		if strings.EqualFold(ptrString(folder.GetDisplayName()), displayName) {
			ids = append(ids, ptrString(folder.GetId()))
		}
	}
	return ids
}
