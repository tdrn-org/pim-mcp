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
	"strings"
	"time"

	msgraphsdkgo "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/tdrn-org/pim-mcp/internal/domain"
)

// graphMailFolderClient implements mailFolderClient on top of the Graph SDK.
type graphMailFolderClient struct {
	graph *msgraphsdkgo.GraphServiceClient
}

// childFolders lists the direct child folders of parentID ("" = root folder).
func (c *graphMailFolderClient) childFolders(ctx context.Context, parentID string) ([]models.MailFolderable, error) {
	selectFields := []string{"id", "displayName", "childFolderCount"}
	var (
		response models.MailFolderCollectionResponseable
		err      error
	)
	if parentID == rootMailFolderID {
		requestConfig := &users.ItemMailFoldersRequestBuilderGetRequestConfiguration{
			QueryParameters: &users.ItemMailFoldersRequestBuilderGetQueryParameters{
				Select: selectFields,
				Top:    int32Ptr(mailFolderPageSize),
				Count:  boolPtr(true),
			},
		}
		response, err = c.graph.Me().MailFolders().Get(ctx, requestConfig)
	} else {
		requestConfig := &users.ItemMailFoldersItemChildFoldersRequestBuilderGetRequestConfiguration{
			QueryParameters: &users.ItemMailFoldersItemChildFoldersRequestBuilderGetQueryParameters{
				Select: selectFields,
				Top:    int32Ptr(mailFolderPageSize),
				Count:  boolPtr(true),
			},
		}
		response, err = c.graph.Me().MailFolders().ByMailFolderId(parentID).ChildFolders().Get(ctx, requestConfig)
	}
	if err != nil {
		return nil, fmt.Errorf("list child folders of mail folder '%s' Graph API failure (cause: %w)", parentID, err)
	}
	return response.GetValue(), nil
}

// messages lists the messages of the given folder matching the filter.
func (c *graphMailFolderClient) messages(ctx context.Context, folderID string, filter domain.EmailFilter) ([]models.Messageable, error) {
	requestConfig := mailFolderMessagesRequestConfig(filter)
	response, err := c.graph.Me().MailFolders().ByMailFolderId(folderID).Messages().Get(ctx, requestConfig)
	if err != nil {
		return nil, fmt.Errorf("list messages of mail folder '%s' Graph API failure (cause: %w)", folderID, err)
	}
	return response.GetValue(), nil
}

// mailFolderMessagesRequestConfig builds the message request configuration for
// a folder scoped message search.
func mailFolderMessagesRequestConfig(filter domain.EmailFilter) *users.ItemMailFoldersItemMessagesRequestBuilderGetRequestConfiguration {
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
	headers := newHeaders().WithDefaults().WithPreferTextContentType().Headers()
	return &users.ItemMailFoldersItemMessagesRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemMailFoldersItemMessagesRequestBuilderGetQueryParameters{
			Search: search,
			Filter: filterParam,
			Top:    limit,
			Count:  boolPtr(true),
		},
		Headers: headers,
	}
}
