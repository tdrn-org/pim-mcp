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

func addDocumentTools(server *mcp.Server, caps domain.ProviderCapabilities, provider domain.DocumentProvider) {
	addSearchDocumentsTool(server, provider)
	addGetDocumentTool(server, provider)
	addDownloadDocumentTool(server, provider)
	if caps.AccessMode == domain.ReadWrite {
		if writeProvider, ok := provider.(domain.DocumentWriteProvider); ok {
			addUploadDocumentTool(server, writeProvider)
		}
	}
}

func addSearchDocumentsTool(server *mcp.Server, provider domain.DocumentProvider) {
	tool := &mcp.Tool{
		Name:        "searchDocuments",
		Description: "Searches for documents using the given search parameters. A document summary including the document ID is returned for every found document. The document ID can be used to get the full document details (getDocument) or download the original file (downloadDocument).",
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, params *SearchDocumentsParams) (*mcp.CallToolResult, any, error) {
		filter := domain.DocumentFilter{
			StandardFilter: domain.StandardFilter{
				Query: params.Query,
				Limit: params.Limit,
			},
			Tags:            params.Tags,
			Correspondent:   params.Correspondent,
			DocumentType:    params.DocumentType,
			DateAddedAfter:  params.DateAddedAfter,
			DateAddedBefore: params.DateAddedBefore,
			CreatedAfter:    params.CreatedAfter,
			CreatedBefore:   params.CreatedBefore,
		}
		documents, err := provider.SearchDocuments(ctx, filter)
		if err != nil {
			return nil, nil, err
		}
		return nil, toDocumentSummaryOutputs(documents), nil
	}
	mcp.AddTool(server, tool, handler)
}

func addGetDocumentTool(server *mcp.Server, provider domain.DocumentProvider) {
	tool := &mcp.Tool{
		Name:        "getDocument",
		Description: "Gets the full document details for the given ID including OCR content.",
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, params *GetDocumentParams) (*mcp.CallToolResult, any, error) {
		document, err := provider.GetDocument(ctx, params.ID)
		if err != nil {
			return nil, nil, err
		}
		return nil, toDocumentOutput(document), nil
	}
	mcp.AddTool(server, tool, handler)
}

func addDownloadDocumentTool(server *mcp.Server, provider domain.DocumentProvider) {
	tool := &mcp.Tool{
		Name:        "downloadDocument",
		Description: "Downloads the original document file for the given ID. Returns the file content and filename.",
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, params *DownloadDocumentParams) (*mcp.CallToolResult, any, error) {
		data, filename, err := provider.DownloadDocument(ctx, params.ID)
		if err != nil {
			return nil, nil, err
		}
		return nil, &DownloadDocumentOutput{
			Filename: filename,
			Size:     len(data),
		}, nil
	}
	mcp.AddTool(server, tool, handler)
}

func addUploadDocumentTool(server *mcp.Server, provider domain.DocumentWriteProvider) {
	tool := &mcp.Tool{
		Name:        "uploadDocument",
		Description: "Uploads a new document to the DMS. The document content is provided as base64-encoded bytes. Requires write access (access_mode = read_write).",
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, params *UploadDocumentParams) (*mcp.CallToolResult, any, error) {
		create := domain.DocumentCreate{
			Title:         params.Title,
			Document:      []byte(params.Document),
			FileName:      params.FileName,
			Tags:          params.Tags,
			Correspondent: params.Correspondent,
			DocumentType:  params.DocumentType,
		}
		summary, err := provider.UploadDocument(ctx, create)
		if err != nil {
			return nil, nil, err
		}
		return nil, toDocumentSummaryOutput(summary), nil
	}
	mcp.AddTool(server, tool, handler)
}

// --- Params types ---

type SearchDocumentsParams struct {
	Query           *string    `json:"query,omitempty"             jsonschema:"Term to search for. All document attributes (Title, Content, Correspondent, Document Type, Tags) are matched against this term (substring match). As soon as one attribute matches, the document is included in the result."`
	Limit           *int       `json:"limit,omitempty"             jsonschema:"The maximum number of documents to return. If no limit is given a provider specific one applies."`
	Tags            []string   `json:"tags,omitempty"              jsonschema:"Filter by tag names. Only documents with all given tags are returned."`
	Correspondent   *string    `json:"correspondent,omitempty"     jsonschema:"Filter by correspondent name (substring match)."`
	DocumentType    *string    `json:"document_type,omitempty"     jsonschema:"Filter by document type name (substring match)."`
	DateAddedAfter  *time.Time `json:"date_added_after,omitempty"  jsonschema:"Only return documents added at or after this time. Use RFC3339 format (e.g. 2026-06-07T00:00:00Z)."`
	DateAddedBefore *time.Time `json:"date_added_before,omitempty" jsonschema:"Only return documents added at or before this time. Use RFC3339 format (e.g. 2026-06-14T00:00:00Z)."`
	CreatedAfter    *time.Time `json:"created_after,omitempty"     jsonschema:"Only return documents with a document date at or after this time. Use RFC3339 format."`
	CreatedBefore   *time.Time `json:"created_before,omitempty"    jsonschema:"Only return documents with a document date at or before this time. Use RFC3339 format."`
}

type GetDocumentParams struct {
	ID string `json:"id" jsonschema:"ID of the document to return."`
}

type DownloadDocumentParams struct {
	ID string `json:"id" jsonschema:"ID of the document to download."`
}

type UploadDocumentParams struct {
	Title         string   `json:"title"                    jsonschema:"The title of the document."`
	Document      string   `json:"document"                jsonschema:"The document content as base64-encoded bytes."`
	FileName      string   `json:"file_name"               jsonschema:"The original filename (e.g. 'invoice.pdf')."`
	Tags          []string `json:"tags,omitempty"          jsonschema:"Optional tags to assign to the document."`
	Correspondent *string  `json:"correspondent,omitempty" jsonschema:"Optional correspondent name."`
	DocumentType  *string  `json:"document_type,omitempty" jsonschema:"Optional document type name."`
}

// --- Output types ---

type SearchDocumentsOutput struct {
	Documents []*DocumentSummaryOutput `json:"documents"`
}

type DocumentSummaryOutput struct {
	ID            string    `json:"id" jsonschema:"ID of the document."`
	Title         string    `json:"title" jsonschema:"The title of the document."`
	Created       time.Time `json:"created" jsonschema:"The document date (e.g. invoice date, RFC3339 format)."`
	Added         time.Time `json:"added" jsonschema:"The date the document was added to the DMS (RFC3339 format)."`
	PageCount     int       `json:"page_count" jsonschema:"The number of pages of the document."`
	MimeType      string    `json:"mime_type" jsonschema:"The MIME type of the document (e.g. 'application/pdf')."`
	Tags          []string  `json:"tags" jsonschema:"The tags assigned to the document."`
	Correspondent string    `json:"correspondent" jsonschema:"The correspondent of the document."`
	DocumentType  string    `json:"document_type" jsonschema:"The document type."`
}

type DocumentOutput struct {
	ID               string    `json:"id" jsonschema:"ID of the document."`
	Title            string    `json:"title" jsonschema:"The title of the document."`
	Content          string    `json:"content" jsonschema:"The OCR content of the document."`
	Created          time.Time `json:"created" jsonschema:"The document date (e.g. invoice date, RFC3339 format)."`
	Added            time.Time `json:"added" jsonschema:"The date the document was added to the DMS (RFC3339 format)."`
	PageCount        int       `json:"page_count" jsonschema:"The number of pages of the document."`
	MimeType         string    `json:"mime_type" jsonschema:"The MIME type of the document."`
	Tags             []string  `json:"tags" jsonschema:"The tags assigned to the document."`
	Correspondent    string    `json:"correspondent" jsonschema:"The correspondent of the document."`
	DocumentType     string    `json:"document_type" jsonschema:"The document type."`
	OriginalFileName string    `json:"original_file_name" jsonschema:"The original filename of the document."`
	ArchivedFileName string    `json:"archived_file_name" jsonschema:"The archived filename of the document."`
}

type DownloadDocumentOutput struct {
	Filename string `json:"filename" jsonschema:"The original filename of the downloaded document."`
	Size     int    `json:"size" jsonschema:"The size of the downloaded document in bytes."`
}

// --- Conversion helpers ---

func toDocumentSummaryOutputs(docs []*domain.DocumentSummary) *SearchDocumentsOutput {
	outputs := make([]*DocumentSummaryOutput, 0, len(docs))
	for _, doc := range docs {
		outputs = append(outputs, toDocumentSummaryOutput(doc))
	}
	return &SearchDocumentsOutput{Documents: outputs}
}

func toDocumentSummaryOutput(doc *domain.DocumentSummary) *DocumentSummaryOutput {
	return &DocumentSummaryOutput{
		ID:            doc.ID,
		Title:         doc.Title,
		Created:       doc.Created,
		Added:         doc.Added,
		PageCount:     doc.PageCount,
		MimeType:      doc.MimeType,
		Tags:          doc.Tags,
		Correspondent: doc.Correspondent,
		DocumentType:  doc.DocumentType,
	}
}

func toDocumentOutput(doc *domain.Document) *DocumentOutput {
	return &DocumentOutput{
		ID:               doc.ID,
		Title:            doc.Title,
		Content:          doc.Content,
		Created:          doc.Created,
		Added:            doc.Added,
		PageCount:        doc.PageCount,
		MimeType:         doc.MimeType,
		Tags:             doc.Tags,
		Correspondent:    doc.Correspondent,
		DocumentType:     doc.DocumentType,
		OriginalFileName: doc.OriginalFileName,
		ArchivedFileName: doc.ArchivedFileName,
	}
}
