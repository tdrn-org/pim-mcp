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

package domain

import (
	"context"
	"strings"
	"time"
)

// DocumentSummary is a lightweight representation returned by search operations.
// It omits the OCR content to keep search results compact.
type DocumentSummary struct {
	ID            string
	Title         string
	Created       time.Time // Document date (e.g. invoice date)
	Added         time.Time // When the document was imported into the DMS
	PageCount     int
	MimeType      string
	Tags          []string
	Correspondent string
	DocumentType  string
}

func (d *DocumentSummary) String() string {
	buffer := &strings.Builder{}
	buffer.WriteString(d.Title)
	if d.Correspondent != "" {
		buffer.WriteString(" (")
		buffer.WriteString(d.Correspondent)
		buffer.WriteString(")")
	}
	buffer.WriteString(" [")
	buffer.WriteString(d.Created.Format(time.DateOnly))
	buffer.WriteString("]")
	return buffer.String()
}

func (d *DocumentSummary) Empty() bool {
	return d.ID == ""
}

// Document is the full representation including OCR content.
// Returned by GetDocument — use DocumentSummary for search results.
type Document struct {
	DocumentSummary
	Content          string // Full OCR text
	OriginalFileName string
	ArchivedFileName string
}

// DocumentFilter defines search criteria for documents.
// All fields are optional — nil means "no filter".
type DocumentFilter struct {
	StandardFilter
	Tags            []string
	Correspondent   *string
	DocumentType    *string
	DateAddedAfter  *time.Time
	DateAddedBefore *time.Time
	CreatedAfter    *time.Time
	CreatedBefore   *time.Time
}

// DocumentCreate describes the fields for uploading a new document.
// Title, Document, and FileName are required.
type DocumentCreate struct {
	Title         string   // Required — display title in the DMS
	Document      []byte   // Required — the document content
	FileName      string   // Required — original filename (e.g. "invoice.pdf")
	Tags          []string // Optional — tags to assign
	Correspondent *string  // Optional — correspondent name
	DocumentType  *string  // Optional — document type name
}

// DocumentProvider is the read-only interface — all adapters implement it.
type DocumentProvider interface {
	SearchDocuments(ctx context.Context, filter DocumentFilter) ([]*DocumentSummary, error)
	GetDocument(ctx context.Context, id string) (*Document, error)
	DownloadDocument(ctx context.Context, id string) ([]byte, string, error) // bytes, filename, error
}

// DocumentWriteProvider extends DocumentProvider with write operations.
// Only registered when ProviderCapabilities.AccessMode == ReadWrite.
type DocumentWriteProvider interface {
	DocumentProvider
	UploadDocument(ctx context.Context, create DocumentCreate) (*DocumentSummary, error)
}
