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

func (p *Provider) SearchEvents(ctx context.Context, filter domain.EventFilter) ([]*domain.Event, error) {
	events := slices.Collect(maps.Values(eventData))
	slices.SortFunc(events, application.EventSortFunc)
	eventList := slices.Values(events)
	eventList = application.Match(eventList, filter.Query, application.EntityFilterFunc[*domain.Event](application.EventFilterFunc))
	eventList = filterByFrom(eventList, filter.From)
	eventList = filterByTo(eventList, filter.To)
	eventList = application.Limit(eventList, filter.Limit)
	return slices.Collect(eventList), nil
}

func (p *Provider) GetEvent(ctx context.Context, id string) (*domain.Event, error) {
	event, ok := eventData[id]
	if !ok {
		return nil, domain.ErrEntityNotFound
	}
	return event, nil
}

func filterByFrom(in iter.Seq[*domain.Event], from *time.Time) iter.Seq[*domain.Event] {
	if from == nil || from.IsZero() {
		return in
	}
	return func(yield func(*domain.Event) bool) {
		for event := range in {
			if !event.End.DateTime.Before(*from) {
				if !yield(event) {
					return
				}
			}
		}
	}
}

func filterByTo(in iter.Seq[*domain.Event], to *time.Time) iter.Seq[*domain.Event] {
	if to == nil || to.IsZero() {
		return in
	}
	return func(yield func(*domain.Event) bool) {
		for event := range in {
			if !event.Start.DateTime.After(*to) {
				if !yield(event) {
					return
				}
			}
		}
	}
}

var eventData map[string]*domain.Event = map[string]*domain.Event{
	"ev1": {
		ID:          "ev1",
		Title:       "Q2 Roadmap Review",
		Description: "Quarterly review of the product roadmap. Each team lead presents progress against Q2 milestones and proposes adjustments for Q3. Meeting room: Kepler (4th floor).",
		Start:       domain.NewTZTime(time.Date(2026, 6, 16, 10, 0, 0, 0, time.UTC), "Europe/Berlin"),
		End:         domain.NewTZTime(time.Date(2026, 6, 16, 11, 30, 0, 0, time.UTC), "Europe/Berlin"),
		Location:    "Kepler Conference Room, 4th Floor",
		Organizer:   domain.NewNamedEmailAddress("alice.chen@example.org", "Alice Chen"),
		Attendees: []domain.NamedEmailAddress{
			domain.NewNamedEmailAddress("bob.martinez@example.org", "Bob Martinez"),
			domain.NewNamedEmailAddress("carol.wang@example.org", "Carol Wang"),
			domain.NewNamedEmailAddress("dave.jones@example.org", "Dave Jones"),
		},
		IsAllDay:  false,
		Status:    domain.EventStatusConfirmed,
		UpdatedAt: time.Date(2026, 6, 14, 9, 0, 0, 0, time.UTC),
	},
	"ev2": {
		ID:          "ev2",
		Title:       "Dentist Appointment",
		Description: "Regular check-up and cleaning. Remember to bring insurance card.",
		Start:       domain.NewTZTime(time.Date(2026, 6, 17, 14, 0, 0, 0, time.UTC), "Europe/Berlin"),
		End:         domain.NewTZTime(time.Date(2026, 6, 17, 15, 0, 0, 0, time.UTC), "Europe/Berlin"),
		Location:    "Dr. Schmidt, Hauptstr. 42, 80331 München",
		Organizer:   domain.NewNamedEmailAddress("employee@example.org", "Employee Name"),
		Attendees:   []domain.NamedEmailAddress{},
		IsAllDay:    false,
		Status:      domain.EventStatusConfirmed,
		UpdatedAt:   time.Date(2026, 6, 8, 11, 30, 0, 0, time.UTC),
	},
	"ev3": {
		ID:          "ev3",
		Title:       "Company Summer Party",
		Description: "Annual summer celebration. BBQ, drinks, and live music. Families welcome. Please RSVP by June 10.",
		Start:       domain.NewTZTime(time.Date(2026, 6, 20, 17, 0, 0, 0, time.UTC), "Europe/Berlin"),
		End:         domain.NewTZTime(time.Date(2026, 6, 20, 23, 0, 0, 0, time.UTC), "Europe/Berlin"),
		Location:    "Rooftop Terrace, Main Building",
		Organizer:   domain.NewNamedEmailAddress("hr@example.org", "Human Resources"),
		Attendees: []domain.NamedEmailAddress{
			domain.NewNamedEmailAddress("all-staff@example.org", "All Staff"),
		},
		IsAllDay:  false,
		Status:    domain.EventStatusConfirmed,
		UpdatedAt: time.Date(2026, 6, 1, 8, 0, 0, 0, time.UTC),
	},
	"ev4": {
		ID:          "ev4",
		Title:       "DevCon 2026",
		Description: "Annual developer conference. Three days of talks, workshops, and networking. Keynote: 'The Future of Distributed Systems' by Dr. Sarah Klein.",
		Start:       domain.NewTZTime(time.Date(2026, 7, 12, 9, 0, 0, 0, time.UTC), "America/New_York"),
		End:         domain.NewTZTime(time.Date(2026, 7, 15, 17, 0, 0, 0, time.UTC), "America/New_York"),
		Location:    "Convention Center, 123 Main St, New York, NY",
		Organizer:   domain.NewNamedEmailAddress("events@devcon.example.org", "DevCon Organizing Committee"),
		Attendees: []domain.NamedEmailAddress{
			domain.NewNamedEmailAddress("employee@example.org", "Employee Name"),
		},
		IsAllDay:  true,
		Status:    domain.EventStatusConfirmed,
		UpdatedAt: time.Date(2026, 5, 20, 16, 0, 0, 0, time.UTC),
	},
	"ev5": {
		ID:          "ev5",
		Title:       "Architecture Review: Auth Service",
		Description: "Technical deep-dive into the new authentication service architecture. Reviewing the OAuth2 flow, token management, and security boundaries. Preparation for the upcoming security audit.",
		Start:       domain.NewTZTime(time.Date(2026, 6, 18, 14, 0, 0, 0, time.UTC), "Europe/Berlin"),
		End:         domain.NewTZTime(time.Date(2026, 6, 18, 16, 0, 0, 0, time.UTC), "Europe/Berlin"),
		Location:    "Virtual — Microsoft Teams",
		Organizer:   domain.NewNamedEmailAddress("dave.jones@example.org", "Dave Jones"),
		Attendees: []domain.NamedEmailAddress{
			domain.NewNamedEmailAddress("alice.chen@example.org", "Alice Chen"),
			domain.NewNamedEmailAddress("bob.martinez@example.org", "Bob Martinez"),
			domain.NewNamedEmailAddress("soc@example.org", "Security Operations Center"),
		},
		IsAllDay:  false,
		Status:    domain.EventStatusTentative,
		UpdatedAt: time.Date(2026, 6, 13, 10, 15, 0, 0, time.UTC),
	},
}
