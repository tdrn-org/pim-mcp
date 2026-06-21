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

package demo

import (
	"context"
	"iter"
	"maps"
	"slices"
	"time"

	"github.com/tdrn-org/pim-mcp/internal/application"
	"github.com/tdrn-org/pim-mcp/internal/domain"
)

func (p *Provider) SearchEmails(ctx context.Context, filter domain.EmailFilter) ([]*domain.Email, error) {
	emails := slices.Collect(maps.Values(emailData))
	slices.SortFunc(emails, application.EmailSortFunc)
	emailList := slices.Values(emails)
	emailList = application.Match(emailList, filter.Query, application.EntityFilterFunc[*domain.Email](application.EmailFilterFunc))
	emailList = filterByUnreadOnly(emailList, filter.UnreadOnly)
	emailList = filterBySince(emailList, filter.Since)
	emailList = application.Limit(emailList, filter.Limit)
	return slices.Collect(emailList), nil
}

func (p *Provider) GetEmail(ctx context.Context, id string) (*domain.Email, error) {
	email, ok := emailData[id]
	if !ok {
		return nil, domain.ErrEntityNotFound
	}
	return email, nil
}

func (p *Provider) UpdateEmail(ctx context.Context, id string, update domain.EmailUpdate) (*domain.Email, error) {
	email, ok := emailData[id]
	if !ok {
		return nil, domain.ErrEntityNotFound
	}
	if update.IsRead != nil {
		email.IsRead = *update.IsRead
	}
	return email, nil
}

func filterByUnreadOnly(in iter.Seq[*domain.Email], unreadOnly bool) iter.Seq[*domain.Email] {
	if !unreadOnly {
		return in
	}
	return func(yield func(*domain.Email) bool) {
		for email := range in {
			if !email.IsRead {
				if !yield(email) {
					return
				}
			}
		}
	}
}

func filterBySince(in iter.Seq[*domain.Email], since *time.Time) iter.Seq[*domain.Email] {
	if since == nil || since.IsZero() {
		return in
	}
	return func(yield func(*domain.Email) bool) {
		for email := range in {
			if !email.ReceivedAt.Before(*since) {
				if !yield(email) {
					return
				}
			}
		}
	}
}

var emailData map[string]*domain.Email = map[string]*domain.Email{
	"e1": {
		ID:      "e1",
		Subject: "Weekly Team Sync — Agenda",
		Body:    "Hi team,\n\nHere is the agenda for our weekly sync on Friday:\n- Q2 roadmap review\n- Budget allocation update\n- Hiring pipeline status\n\nPlease add any additional topics by Thursday.\n\nBest,\nAlice",
		From:    domain.NewNamedEmailAddress("alice.chen@example.org", "Alice Chen"),
		To: []domain.NamedEmailAddress{
			domain.NewNamedEmailAddress("team@example.org", "Engineering Team"),
		},
		CC:         []domain.NamedEmailAddress{},
		ReceivedAt: time.Date(2026, 6, 15, 9, 30, 0, 0, time.UTC),
		SentAt:     time.Date(2026, 6, 15, 9, 28, 0, 0, time.UTC),
		IsRead:     false,
		Folder:     "Inbox",
		ThreadID:   "thread-w1",
	},
	"e2": {
		ID:      "e2",
		Subject: "Re: Q2 Budget Proposal",
		Body:    "Bob,\n\nThanks for the detailed breakdown. I have a few questions about the infrastructure line items — can we discuss in tomorrow's call?\n\nRegards,\nAlice",
		From:    domain.NewNamedEmailAddress("alice.chen@example.org", "Alice Chen"),
		To: []domain.NamedEmailAddress{
			domain.NewNamedEmailAddress("bob.martinez@example.org", "Bob Martinez"),
		},
		CC: []domain.NamedEmailAddress{
			domain.NewNamedEmailAddress("finance@example.org", "Finance Department"),
		},
		ReceivedAt: time.Date(2026, 6, 14, 16, 45, 0, 0, time.UTC),
		SentAt:     time.Date(2026, 6, 14, 16, 42, 0, 0, time.UTC),
		IsRead:     true,
		Folder:     "Inbox",
		ThreadID:   "thread-b1",
	},
	"e3": {
		ID:      "e3",
		Subject: "Invoice #2026-0642 attached",
		Body:    "Dear valued customer,\n\nPlease find your latest invoice attached to this email. Payment is due within 30 days.\n\nIf you have any questions, please contact our billing department.\n\nThank you for your business.\n\nSincerely,\nBilling Department\nCloudHost Solutions",
		From:    domain.NewNamedEmailAddress("billing@cloudhost.example.org", "CloudHost Billing"),
		To: []domain.NamedEmailAddress{
			domain.NewNamedEmailAddress("accounts@example.org", "Accounts Payable"),
		},
		CC:         []domain.NamedEmailAddress{},
		ReceivedAt: time.Date(2026, 6, 13, 8, 15, 0, 0, time.UTC),
		SentAt:     time.Date(2026, 6, 13, 8, 12, 0, 0, time.UTC),
		IsRead:     false,
		Folder:     "Inbox",
		ThreadID:   "thread-i1",
	},
	"e4": {
		ID:      "e4",
		Subject: "Conference travel confirmation",
		Body:    "Your travel arrangements for DevCon 2026 have been confirmed:\n\nFlight: LH402, July 12, 08:00 → 10:30\nHotel: Marriott Downtown, check-in July 12\nReturn: LH407, July 15, 17:00 → 19:30\n\nPlease save this email for your records.\n\nTravel Desk",
		From:    domain.NewNamedEmailAddress("travel@example.org", "Corporate Travel"),
		To: []domain.NamedEmailAddress{
			domain.NewNamedEmailAddress("employee@example.org", "Employee Name"),
		},
		CC:         []domain.NamedEmailAddress{},
		ReceivedAt: time.Date(2026, 6, 10, 14, 0, 0, 0, time.UTC),
		SentAt:     time.Date(2026, 6, 10, 13, 55, 0, 0, time.UTC),
		IsRead:     true,
		Folder:     "Inbox",
		ThreadID:   "thread-t1",
	},
	"e5": {
		ID:      "e5",
		Subject: "Security advisory: patch required",
		Body:    "This is an automated notification from the security team.\n\nA critical vulnerability (CVE-2026-12345) has been identified in the authentication module. All systems must be patched to version 4.2.1 or later within 48 hours.\n\nAffected systems: auth-service, api-gateway\nPatch available at: https://updates.example.org/patches/4.2.1\n\n— Security Operations Center",
		From:    domain.NewNamedEmailAddress("soc@example.org", "Security Operations Center"),
		To: []domain.NamedEmailAddress{
			domain.NewNamedEmailAddress("devops@example.org", "DevOps Team"),
		},
		CC: []domain.NamedEmailAddress{
			domain.NewNamedEmailAddress("cto@example.org", "Chief Technology Officer"),
		},
		ReceivedAt: time.Date(2026, 6, 15, 7, 0, 0, 0, time.UTC),
		SentAt:     time.Date(2026, 6, 15, 6, 58, 0, 0, time.UTC),
		IsRead:     false,
		ThreadID:   "thread-s1",
	},
}
