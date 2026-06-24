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
	"slices"
	"time"

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
	"github.com/tdrn-org/pim-mcp/internal/application"
	"github.com/tdrn-org/pim-mcp/internal/domain"
)

func (p *Provider) SearchEvents(ctx context.Context, filter domain.EventFilter) ([]*domain.Event, error) {
	client, err := p.graphClient(ctx)
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
	client, err := p.graphClient(ctx)
	if err != nil {
		return nil, err
	}
	headers := newHeaders().WithDefaults().WithPreferTextContentType().WithPreferTimezone(p.timeLocation).Headers()
	requestConfig := &users.ItemEventsEventItemRequestBuilderGetRequestConfiguration{
		Headers: headers,
	}
	response, err := client.Me().Events().ByEventId(id).Get(ctx, requestConfig)
	if err != nil {
		return nil, fmt.Errorf("get event Graph API failure (cause: %w)", err)
	}
	event := p.eventFromResponse(response)
	return event, nil
}

func (p *Provider) CreateEvent(ctx context.Context, create domain.EventCreate) (*domain.Event, error) {
	requestBuilder := eventRequestBuilder{Request: models.NewEvent(), runtimeTimezone: p.runtimeTimezone}
	request := requestBuilder.
		Title(&create.Title).
		Description(create.Description).
		Start(&create.Start).
		End(&create.End).
		Location(create.Location).
		IsAllDay(create.IsAllDay).
		Request
	client, err := p.graphClient(ctx)
	if err != nil {
		return nil, err
	}
	response, err := client.Me().Events().Post(ctx, request, nil)
	if err != nil {
		return nil, fmt.Errorf("create event Graph API failure (cause: %w)", err)
	}
	event := p.eventFromResponse(response)
	return event, nil
}

func (p *Provider) eventFromResponse(model models.Eventable) *domain.Event {
	sensitivity := model.GetSensitivity()
	if sensitivity != nil && *sensitivity > models.Sensitivity(p.cfg.MSGraph.SensitivityLimit) {
		return &domain.Event{}
	}
	body := model.GetBody()
	content := ""
	if body != nil {
		content = *body.GetContent()
	}
	return &domain.Event{
		ID:          ptrString(model.GetId()),
		Title:       ptrString(model.GetSubject()),
		Description: content,
		Start:       p.tzTimeFromResponse(model.GetStart()),
		End:         p.tzTimeFromResponse(model.GetEnd()),
		Location:    ptrString(model.GetLocation().GetDisplayName()),
		Organizer:   p.attendeeFromResponse(model.GetOrganizer()),
		Attendees:   p.attendeesFromResponse(model.GetAttendees()),
		IsAllDay:    ptrBool(model.GetIsAllDay(), false),
		Status:      p.eventStatusFromResponse(model),
		UpdatedAt:   ptrTime(model.GetLastModifiedDateTime()),
	}
}

func (p *Provider) tzTimeFromResponse(model models.DateTimeTimeZoneable) domain.TZTime {
	return unmarshalTZTime(model.GetDateTime(), model.GetTimeZone(), p.timeLocation)
}

func (p *Provider) attendeesFromResponse(models []models.Attendeeable) []domain.NamedEmailAddress {
	attendees := make([]domain.NamedEmailAddress, 0, len(models))
	for _, model := range models {
		attendees = append(attendees, p.attendeeFromResponse(model))
	}
	return attendees
}

func (p *Provider) attendeeFromResponse(model models.Recipientable) domain.NamedEmailAddress {
	emailAddress := model.GetEmailAddress()
	return domain.NewNamedEmailAddress(ptrString(emailAddress.GetAddress()), ptrString(emailAddress.GetName()))
}

func (p *Provider) eventStatusFromResponse(model models.Eventable) domain.EventStatus {
	canceled := ptrBool(model.GetIsCancelled(), false)
	if canceled {
		return domain.EventStatusCanceled
	}
	return domain.EventStatusConfirmed
}

func (p *Provider) eventFilterRequestConfig(filter domain.EventFilter) *users.ItemCalendarViewRequestBuilderGetRequestConfiguration {
	search, limit := standardFilterPtr(filter.StandardFilter)
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
	headers := newHeaders().WithDefaults().WithPreferTextContentType().WithPreferTimezone(p.timeLocation).Headers()
	requestConfig := &users.ItemCalendarViewRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.ItemCalendarViewRequestBuilderGetQueryParameters{
			Search:        search,
			Top:           limit,
			StartDateTime: &start,
			EndDateTime:   &end,
			Count:         boolPtr(true),
		},
		Headers: headers,
	}
	return requestConfig
}

type eventRequestBuilder struct {
	Request         *models.Event
	runtimeTimezone string
}

func (b *eventRequestBuilder) Title(title *string) *eventRequestBuilder {
	if title != nil {
		b.Request.SetSubject(title)
	}
	return b
}

func (b *eventRequestBuilder) Description(description *string) *eventRequestBuilder {
	if description != nil {
		body := models.NewItemBody()
		body.SetContentType(bodyTypePtr(models.TEXT_BODYTYPE))
		body.SetContent(description)
		b.Request.SetBody(body)
	}
	return b
}

func (b *eventRequestBuilder) Start(start *domain.TZTime) *eventRequestBuilder {
	if start != nil {
		if start.Empty() {
			b.Request.SetStart(nil)
		} else {
			startDateTime := models.NewDateTimeTimeZone()
			dateTime, timezone := marshalTZTime(*start, b.runtimeTimezone)
			startDateTime.SetDateTime(dateTime)
			startDateTime.SetTimeZone(timezone)
			b.Request.SetStart(startDateTime)
		}
	}
	return b
}

func (b *eventRequestBuilder) End(end *domain.TZTime) *eventRequestBuilder {
	if end != nil {
		if end.Empty() {
			b.Request.SetEnd(nil)
		} else {
			endDateTime := models.NewDateTimeTimeZone()
			dateTime, timezone := marshalTZTime(*end, b.runtimeTimezone)
			endDateTime.SetDateTime(dateTime)
			endDateTime.SetTimeZone(timezone)
			b.Request.SetEnd(endDateTime)
		}
	}
	return b
}

func (b *eventRequestBuilder) Location(location *string) *eventRequestBuilder {
	if location != nil {
		locationModel := models.NewLocation()
		locationModel.SetDisplayName(location)
		b.Request.SetLocation(locationModel)
	}
	return b
}

func (b *eventRequestBuilder) IsAllDay(isAllDay *bool) *eventRequestBuilder {
	if isAllDay != nil {
		b.Request.SetIsAllDay(isAllDay)
	}
	return b
}
