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

type Event struct {
	ID          string
	Title       string
	Description string
	Start       TimeRange
	End         TimeRange
	Location    string
	Organizer   Attendee
	Attendees   []Attendee
	IsAllDay    bool
	Status      EventStatus
	UpdatedAt   time.Time
}

func (e *Event) String() string {
	buffer := strings.Builder{}
	buffer.WriteString(e.Start.DateTime.Format(time.DateTime))
	buffer.WriteRune(' ')
	buffer.WriteString(e.End.DateTime.Sub(e.Start.DateTime).String())
	buffer.WriteRune(' ')
	buffer.WriteString(e.Title)
	return buffer.String()
}

func (e *Event) Empty() bool {
	return e.ID == ""
}

type TimeRange struct {
	DateTime time.Time
	Timezone string
}

func NewTimeRange(dateTime time.Time, timezone string) TimeRange {
	return TimeRange{
		DateTime: dateTime,
		Timezone: timezone,
	}
}

type Attendee struct {
	Name  string
	Email string
}

func NewAttendee(name, email string) Attendee {
	return Attendee{
		Name:  name,
		Email: email,
	}
}

type EventStatus string

const (
	StatusConfirmed EventStatus = "confirmed"
	StatusTentative EventStatus = "tentative"
	StatusCanceled  EventStatus = "canceled"
)

type CalendarProvider interface {
	SearchEvents(ctx context.Context, filter EventFilter) ([]*Event, error)
	GetEvent(ctx context.Context, id string) (*Event, error)
}

type EventFilter struct {
	Query string
	Limit int
	From  *time.Time
	To    *time.Time
}
