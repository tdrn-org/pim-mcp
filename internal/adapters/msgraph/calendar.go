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
	"slices"
	"time"

	kiota "github.com/microsoft/kiota-abstractions-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/tdrn-org/pim-mcp/internal/application"
	"github.com/tdrn-org/pim-mcp/internal/domain"
)

func (p *Provider) SearchEvents(ctx context.Context, filter domain.EventFilter) ([]*domain.Event, error) {
	client, err := p.graphClient()
	if err != nil {
		return nil, err
	}
	requestConfig := p.eventFilterRequestConfig(filter)
	response, err := client.Me().CalendarView().Get(ctx, requestConfig)
	if err != nil {
		return nil, fmt.Errorf("search events Graph API failure (cause: %w)", err)
	}
	events := make([]*domain.Event, 0)
	for _, responseItem := range response.GetValue() {
		event := p.eventFromResponse(responseItem)
		if !event.Empty() {
			events = append(events, event)
		}
	}
	slices.SortFunc(events, application.EventSortFunc)
	return events, nil
}

func (p *Provider) GetEvent(ctx context.Context, id string) (*domain.Event, error) {
	client, err := p.graphClient()
	if err != nil {
		return nil, err
	}
	response, err := client.Me().Events().ByEventId(id).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("get event Graph API failure (cause: %w)", err)
	}
	event := p.eventFromResponse(response)
	return event, nil
}

func (p *Provider) eventFromResponse(model models.Eventable) *domain.Event {

	return &domain.Event{
		ID:          ptrString(model.GetId()),
		Title:       ptrString(model.GetSubject()),
		Description: ptrString(model.GetBody().GetContent()),
		Start:       p.timeRangeFromResponse(model.GetStart()),
		End:         p.timeRangeFromResponse(model.GetEnd()),
		Location:    ptrString(model.GetLocation().GetDisplayName()),
		Organizer:   p.attendeeFromResponse(model.GetOrganizer()),
		Attendees:   p.attendeesFromResponse(model.GetAttendees()),
		IsAllDay:    ptrBool(model.GetIsAllDay(), false),
		Status:      p.eventStatusFromResponse(model),
		UpdatedAt:   ptrTime(model.GetLastModifiedDateTime()),
	}
}

const eventDateTimeLayoutLong string = "2006-01-02T15:04:05.0000000"
const eventDateTimeLayoutShort string = "2006-01-02T15:04:05"

func (p *Provider) timeRangeFromResponse(model models.DateTimeTimeZoneable) domain.TimeRange {
	tz := model.GetTimeZone()
	location := time.UTC
	if tz != nil && *tz != "" {
		tzLocation, err := time.LoadLocation(*tz)
		if err == nil {
			location = tzLocation
		} else {
			p.logger.Info("unable to parse event time zone", slog.String("tz", *tz))
		}
	}
	dt := model.GetDateTime()
	dateTime := time.Time{}
	if dt != nil && *dt != "" {
		layout := eventDateTimeLayoutLong
		if len(*dt) <= len(eventDateTimeLayoutShort) {
			layout = eventDateTimeLayoutShort
		}
		parsedDT, err := time.ParseInLocation(layout, *dt, location)
		if err == nil {
			dateTime = parsedDT
		} else {
			p.logger.Info("unable to parse event date-time", slog.String("dt", *dt))
		}
	}
	return domain.NewTimeRange(dateTime, ptrString(tz))
}

func (p *Provider) attendeesFromResponse(models []models.Attendeeable) []domain.Attendee {
	attendees := make([]domain.Attendee, 0, len(models))
	for _, model := range models {
		attendees = append(attendees, p.attendeeFromResponse(model))
	}
	return attendees
}

func (p *Provider) attendeeFromResponse(model models.Recipientable) domain.Attendee {
	emailAddress := model.GetEmailAddress()
	return domain.NewAttendee(ptrString(emailAddress.GetName()), ptrString(emailAddress.GetAddress()))
}

func (p *Provider) eventStatusFromResponse(model models.Eventable) domain.EventStatus {
	canceled := ptrBool(model.GetIsCancelled(), false)
	if canceled {
		return domain.StatusCanceled
	}
	return domain.StatusConfirmed
}

func (p *Provider) eventFilterRequestConfig(filter domain.EventFilter) *users.ItemCalendarViewRequestBuilderGetRequestConfiguration {
	limit := int32(filter.Limit)
	if limit <= 0 {
		limit = int32(DefaultSearchLimit)
	}
	var search *string
	if filter.Query != "" {
		search = stringPtr(fmt.Sprintf("\"%s\"", filter.Query))
	}
	nowUTC := time.Now().UTC()
	var start string
	if filter.From != nil && !filter.From.IsZero() {
		start = filter.From.UTC().Format(time.RFC3339)
	} else {
		start = nowUTC.Format(time.RFC3339)
	}
	var end string
	if filter.To != nil && !filter.To.IsZero() {
		end = filter.To.UTC().Format(time.RFC3339)
	} else {
		end = nowUTC.Add(7 * 24 * time.Hour).Format(time.RFC3339)
	}
	headers := &kiota.RequestHeaders{}
	headers.Add("ConsistencyLevel", "eventual")
	requestConfig := &users.ItemCalendarViewRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemCalendarViewRequestBuilderGetQueryParameters{
			Search:        search,
			Top:           &limit,
			StartDateTime: &start,
			EndDateTime:   &end,
			Count:         boolPtr(true),
		},
		Headers: headers,
	}
	return requestConfig
}
