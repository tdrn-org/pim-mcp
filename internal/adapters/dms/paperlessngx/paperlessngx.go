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

package paperlessngx

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/url"
	"time"

	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/tdrn-org/go-httpserver"
	paperlessngx "github.com/tdrn-org/go-paperless-ngx"
	"github.com/tdrn-org/go-paperless-ngx/api"
	"github.com/tdrn-org/pim-mcp/config"
	"github.com/tdrn-org/pim-mcp/internal/adapters/dms"
	"github.com/tdrn-org/pim-mcp/internal/domain"
)

const Name = "paperlessngx"

// Runtime is the interface the adapter needs from the server.
type Runtime interface {
	BaseURL() *url.URL
	Logger() *slog.Logger
}

// Provider implements dms.Provider and domain.DocumentProvider/domain.DocumentWriteProvider.
type Provider struct {
	runtime Runtime
	cfg     *config.DMSConfig
	client  *paperlessngx.Client
	logger  *slog.Logger
}

// NewProvider creates a new PaperlessNGX DMS provider.
func NewProvider(runtime Runtime, cfg *config.DMSConfig) (*Provider, error) {
	apiURL := cfg.PaperlessNGX.APIURL.URL
	if apiURL == nil {
		return nil, fmt.Errorf("paperlessngx api_url is not configured")
	}
	client, err := paperlessngx.NewClient(apiURL, cfg.PaperlessNGX.APIKey,
		paperlessngx.WithLogger(slog.With(slog.String("adapter", Name))),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create paperlessngx client (cause: %w)", err)
	}
	return &Provider{
		runtime: runtime,
		cfg:     cfg,
		client:  client,
		logger:  slog.With(slog.String("provider", Name)),
	}, nil
}

// ID returns a unique provider instance ID.
func (p *Provider) ID() string {
	return uuid.NewString()
}

// Name returns the provider name.
func (*Provider) Name() string {
	return Name
}

// Capabilities reports DMS capabilities based on config.
func (p *Provider) Capabilities() domain.ProviderCapabilities {
	return domain.ProviderCapabilities{
		Documents:  true,
		AccessMode: domain.AccessMode(p.cfg.AccessMode),
	}
}

// Mount registers provider-specific HTTP handlers.
func (p *Provider) Mount(server *httpserver.Instance) {
	// PaperlessNGX uses API key auth — no OAuth flow needed.
	// Mount is a no-op but required by the interface.
}

// LoginURL returns the login URL (not used for API key auth).
func (p *Provider) LoginURL() *url.URL {
	return p.runtime.BaseURL()
}

// CheckCredentials verifies the API key is still valid.
func (p *Provider) CheckCredentials(ctx context.Context, sessionID, credentials string) *dms.CredentialInfo {
	// PaperlessNGX uses a static API key — always valid if configured.
	return &dms.CredentialInfo{Valid: p.cfg.PaperlessNGX.APIKey != ""}
}

// RefreshCredentials is a no-op for static API key auth.
func (p *Provider) RefreshCredentials(ctx context.Context, sessionID, credentials string, refreshInterval time.Duration) string {
	return credentials
}

// SearchDocuments searches for documents matching the given filter.
func (p *Provider) SearchDocuments(ctx context.Context, filter domain.DocumentFilter) ([]*domain.DocumentSummary, error) {
	params := &api.DocumentsListParams{}

	if filter.Query != nil {
		params.Query = filter.Query
	}
	if filter.Limit != nil {
		pageSize := *filter.Limit
		params.PageSize = &pageSize
	}
	if filter.Correspondent != nil {
		params.CorrespondentNameIcontains = filter.Correspondent
	}
	if filter.DocumentType != nil {
		params.DocumentTypeNameIcontains = filter.DocumentType
	}
	if len(filter.Tags) > 0 {
		// Use tag name icontains for the first tag
		params.TagsNameIcontains = &filter.Tags[0]
	}
	if filter.CreatedAfter != nil {
		params.CreatedGte = toDatePtr(*filter.CreatedAfter)
	}
	if filter.CreatedBefore != nil {
		params.CreatedLte = toDatePtr(*filter.CreatedBefore)
	}
	if filter.DateAddedAfter != nil {
		params.AddedGte = filter.DateAddedAfter
	}
	if filter.DateAddedBefore != nil {
		params.AddedLte = filter.DateAddedBefore
	}

	// Default ordering: newest first
	ordering := "-created"
	params.Ordering = &ordering

	response, err := p.client.DocumentsList(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to search documents (cause: %w)", err)
	}
	if response.JSON200 == nil {
		return nil, fmt.Errorf("unexpected search response")
	}

	// Resolve ID→name mappings
	tagNames, err := p.resolveTagNames(ctx)
	if err != nil {
		p.logger.Warn("failed to resolve tag names", slog.Any("err", err))
		tagNames = map[int]string{}
	}
	correspondentNames, err := p.resolveCorrespondentNames(ctx)
	if err != nil {
		p.logger.Warn("failed to resolve correspondent names", slog.Any("err", err))
		correspondentNames = map[int]string{}
	}
	documentTypeNames, err := p.resolveDocumentTypeNames(ctx)
	if err != nil {
		p.logger.Warn("failed to resolve document type names", slog.Any("err", err))
		documentTypeNames = map[int]string{}
	}

	summaries := make([]*domain.DocumentSummary, 0, len(response.JSON200.Results))
	for _, doc := range response.JSON200.Results {
		summary := convertToSummary(&doc, tagNames, correspondentNames, documentTypeNames)
		summaries = append(summaries, summary)
	}
	return summaries, nil
}

// GetDocument retrieves a single document by ID including OCR content.
func (p *Provider) GetDocument(ctx context.Context, id string) (*domain.Document, error) {
	intID, err := parseID(id)
	if err != nil {
		return nil, err
	}
	response, err := p.client.DocumentsRetrieve(ctx, intID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get document '%s' (cause: %w)", id, err)
	}
	if response.JSON200 == nil {
		return nil, fmt.Errorf("document '%s' not found", id)
	}

	tagNames, _ := p.resolveTagNames(ctx)
	correspondentNames, _ := p.resolveCorrespondentNames(ctx)
	documentTypeNames, _ := p.resolveDocumentTypeNames(ctx)

	doc := convertToDocument(response.JSON200, tagNames, correspondentNames, documentTypeNames)
	return doc, nil
}

// DownloadDocument downloads the original document file.
func (p *Provider) DownloadDocument(ctx context.Context, id string) ([]byte, string, error) {
	intID, err := parseID(id)
	if err != nil {
		return nil, "", err
	}
	response, err := p.client.DocumentsDownloadRetrieve(ctx, intID, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to download document '%s' (cause: %w)", id, err)
	}
	// The download response body contains the raw file bytes.
	// We need to get the filename from the document metadata first.
	docResponse, err := p.client.DocumentsRetrieve(ctx, intID, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get document metadata for download '%s' (cause: %w)", id, err)
	}
	filename := "document"
	if docResponse.JSON200 != nil && docResponse.JSON200.OriginalFileName != nil {
		filename = *docResponse.JSON200.OriginalFileName
	}
	return response.Body, filename, nil
}

// UploadDocument uploads a new document to PaperlessNGX.
func (p *Provider) UploadDocument(ctx context.Context, create domain.DocumentCreate) (*domain.DocumentSummary, error) {
	// Resolve tag names to IDs
	tagIDs, err := p.resolveTagIDs(ctx, create.Tags)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve tags (cause: %w)", err)
	}

	// Resolve correspondent name to ID
	var correspondentID *int
	if create.Correspondent != nil {
		id, err := p.resolveCorrespondentID(ctx, *create.Correspondent)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve correspondent '%s' (cause: %w)", *create.Correspondent, err)
		}
		correspondentID = id
	}

	// Resolve document type name to ID
	var documentTypeID *int
	if create.DocumentType != nil {
		id, err := p.resolveDocumentTypeID(ctx, *create.DocumentType)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve document type '%s' (cause: %w)", *create.DocumentType, err)
		}
		documentTypeID = id
	}

	// Build multipart form body
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add the document file
	part, err := writer.CreateFormFile("document", create.FileName)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file (cause: %w)", err)
	}
	if _, err := part.Write(create.Document); err != nil {
		return nil, fmt.Errorf("failed to write document data (cause: %w)", err)
	}

	// Add title
	if create.Title != "" {
		writer.WriteField("title", create.Title)
	}

	contentType := writer.FormDataContentType()
	writer.Close()

	response, err := p.client.DocumentsPostDocumentCreateWithBody(ctx, contentType, &buf)
	if err != nil {
		return nil, fmt.Errorf("failed to upload document (cause: %w)", err)
	}

	// The response is a task ID string — we need to wait and then look up the created document.
	// For now, return a summary with what we know.
	// TODO: implement task polling to get the actual document ID.
	_ = response
	_ = tagIDs
	_ = correspondentID
	_ = documentTypeID

	return nil, fmt.Errorf("upload not yet fully implemented — task polling needed")
}

// --- ID↔Name resolution helpers ---

func (p *Provider) resolveTagNames(ctx context.Context) (map[int]string, error) {
	response, err := p.client.TagsList(ctx, nil)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, fmt.Errorf("unexpected tags response")
	}
	names := make(map[int]string, len(response.JSON200.Results))
	for _, tag := range response.JSON200.Results {
		if tag.Id != nil {
			names[*tag.Id] = tag.Name
		}
	}
	return names, nil
}

func (p *Provider) resolveCorrespondentNames(ctx context.Context) (map[int]string, error) {
	response, err := p.client.CorrespondentsList(ctx, nil)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, fmt.Errorf("unexpected correspondents response")
	}
	names := make(map[int]string, len(response.JSON200.Results))
	for _, c := range response.JSON200.Results {
		if c.Id != nil {
			names[*c.Id] = c.Name
		}
	}
	return names, nil
}

func (p *Provider) resolveDocumentTypeNames(ctx context.Context) (map[int]string, error) {
	response, err := p.client.DocumentTypesList(ctx, nil)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, fmt.Errorf("unexpected document types response")
	}
	names := make(map[int]string, len(response.JSON200.Results))
	for _, dt := range response.JSON200.Results {
		if dt.Id != nil {
			names[*dt.Id] = dt.Name
		}
	}
	return names, nil
}

func (p *Provider) resolveTagIDs(ctx context.Context, tagNames []string) ([]int, error) {
	if len(tagNames) == 0 {
		return nil, nil
	}
	response, err := p.client.TagsList(ctx, nil)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, fmt.Errorf("unexpected tags response")
	}
	nameToID := make(map[string]int)
	for _, tag := range response.JSON200.Results {
		if tag.Id != nil {
			nameToID[tag.Name] = *tag.Id
		}
	}
	ids := make([]int, 0, len(tagNames))
	for _, name := range tagNames {
		id, ok := nameToID[name]
		if !ok {
			return nil, fmt.Errorf("unknown tag: '%s'", name)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (p *Provider) resolveCorrespondentID(ctx context.Context, name string) (*int, error) {
	response, err := p.client.CorrespondentsList(ctx, nil)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, fmt.Errorf("unexpected correspondents response")
	}
	for _, c := range response.JSON200.Results {
		if c.Name == name && c.Id != nil {
			return c.Id, nil
		}
	}
	return nil, fmt.Errorf("unknown correspondent: '%s'", name)
}

func (p *Provider) resolveDocumentTypeID(ctx context.Context, name string) (*int, error) {
	response, err := p.client.DocumentTypesList(ctx, nil)
	if err != nil {
		return nil, err
	}
	if response.JSON200 == nil {
		return nil, fmt.Errorf("unexpected document types response")
	}
	for _, dt := range response.JSON200.Results {
		if dt.Name == name && dt.Id != nil {
			return dt.Id, nil
		}
	}
	return nil, fmt.Errorf("unknown document type: '%s'", name)
}

// --- Conversion helpers ---

func convertToSummary(doc *api.Document, tagNames, correspondentNames, documentTypeNames map[int]string) *domain.DocumentSummary {
	s := &domain.DocumentSummary{}
	if doc.Id != nil {
		s.ID = fmt.Sprintf("%d", *doc.Id)
	}
	if doc.Title != nil {
		s.Title = *doc.Title
	}
	if doc.Created != nil {
		s.Created = doc.Created.Time
	}
	if doc.Added != nil {
		s.Added = *doc.Added
	}
	if doc.PageCount != nil {
		s.PageCount = *doc.PageCount
	}
	if doc.MimeType != nil {
		s.MimeType = *doc.MimeType
	}
	for _, tagID := range doc.Tags {
		if name, ok := tagNames[tagID]; ok {
			s.Tags = append(s.Tags, name)
		}
	}
	if doc.Correspondent != nil {
		if name, ok := correspondentNames[*doc.Correspondent]; ok {
			s.Correspondent = name
		}
	}
	if doc.DocumentType != nil {
		if name, ok := documentTypeNames[*doc.DocumentType]; ok {
			s.DocumentType = name
		}
	}
	return s
}

func convertToDocument(doc *api.Document, tagNames, correspondentNames, documentTypeNames map[int]string) *domain.Document {
	d := &domain.Document{
		DocumentSummary: *convertToSummary(doc, tagNames, correspondentNames, documentTypeNames),
	}
	if doc.Content != nil {
		d.Content = *doc.Content
	}
	if doc.OriginalFileName != nil {
		d.OriginalFileName = *doc.OriginalFileName
	}
	if doc.ArchivedFileName != nil {
		d.ArchivedFileName = *doc.ArchivedFileName
	}
	return d
}

// --- Helpers ---

func parseID(id string) (int, error) {
	var intID int
	_, err := fmt.Sscanf(id, "%d", &intID)
	if err != nil {
		return 0, fmt.Errorf("invalid document ID '%s': must be numeric", id)
	}
	return intID, nil
}

func toDatePtr(t time.Time) *openapi_types.Date {
	d := openapi_types.Date{Time: t}
	return &d
}
