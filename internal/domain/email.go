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

type Email struct {
	ID         string
	Subject    string
	Body       string
	From       NamedEmailAddress
	To         []NamedEmailAddress
	CC         []NamedEmailAddress
	ReceivedAt time.Time
	SentAt     time.Time
	IsRead     bool
	ThreadID   string
}

func (e *Email) String() string {
	buffer := &strings.Builder{}
	if !e.IsRead {
		buffer.WriteString("* ")
	}
	buffer.WriteString(e.From.Address.String())
	buffer.WriteRune(' ')
	buffer.WriteString(e.Subject)
	buffer.WriteRune(' ')
	buffer.WriteString(e.ReceivedAt.Format(time.DateTime))
	return buffer.String()
}

func (e *Email) Empty() bool {
	return e.ID == ""
}

// MailUpdate beschreibt, welche Felder einer Mail geändert werden sollen.
// Nur non-nil Felder werden angewandt (PATCH-Semantik).
// Aktuell nur IsRead — erweiterbar wie TaskUpdate.
type MailUpdate struct {
	IsRead *bool // nil=don't touch; nur &true verwenden (Sicherheitsphilosophie)
}

type EmailProvider interface {
	SearchEmails(ctx context.Context, filter EmailFilter) ([]*Email, error)
	GetEmail(ctx context.Context, id string) (*Email, error)
}

type EmailWriteProvider interface {
	EmailProvider
	UpdateMail(ctx context.Context, id string, update MailUpdate) (*Email, error)
}

type EmailFilter struct {
	StandardFilter
	UnreadOnly bool
	Folder     *string
	Since      *time.Time
}
