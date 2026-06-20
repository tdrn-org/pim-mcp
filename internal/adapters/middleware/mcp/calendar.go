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
	"fmt"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tdrn-org/pim-mcp/internal/domain"
)

func addCalendarTools(server *mcp.Server, caps domain.ProviderCapabilities, provider domain.CalendarProvider) {
	addSearchEventsTool(server, provider)
	addGetEventTool(server, provider)
	if caps.AccessMode == domain.ReadWrite {
		if writeProvider, ok := provider.(domain.CalendarWriteProvider); ok {
			addCreateEventTool(server, writeProvider)
		}
	}
}

func addSearchEventsTool(server *mcp.Server, provider domain.CalendarProvider) {
	tool := &mcp.Tool{
		Name:        "searchEvents",
		Description: "Searches for events using the given search parameters. An event summary including the event ID is returned for every found event. The event ID can be used to get the full event details (getEvent).",
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, params *SearchEventsParams) (*mcp.CallToolResult, any, error) {
		filter := domain.EventFilter{
			StandardFilter: domain.StandardFilter{
				Query: params.Query,
				Limit: params.Limit,
			},
			From: params.From,
			To:   params.To,
		}
		events, err := provider.SearchEvents(ctx, filter)
		if err != nil {
			return nil, nil, err
		}
		return nil, toEventSummaryOutputs(events), nil
	}
	mcp.AddTool(server, tool, handler)
}

func addGetEventTool(server *mcp.Server, provider domain.CalendarProvider) {
	tool := &mcp.Tool{
		Name:        "getEvent",
		Description: "Gets the full event details for the given ID",
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, params *GetEventParams) (*mcp.CallToolResult, any, error) {
		event, err := provider.GetEvent(ctx, params.ID)
		if err != nil {
			return nil, nil, err
		}
		return nil, toEventOutput(event), nil
	}
	mcp.AddTool(server, tool, handler)
}

type SearchEventsParams struct {
	Query *string    `json:"query,omitempty" jsonschema:"Term to search for. All event attributes (Title, Description, Location, Organizer Name) are matched against this term (substring match). As soon as one attribute matches, the event is included in the result."`
	Limit *int       `json:"limit,omitempty" jsonschema:"The maximum number of events to return. If no limit is given a provider specific one applies."`
	From  *time.Time `json:"from,omitempty"  jsonschema:"Only return events scheduled at or after this time. Use RFC3339 format (e.g. 2026-06-14T00:00:00Z)."`
	To    *time.Time `json:"to,omitempty"    jsonschema:"Only return events scheduled before this time. Use RFC3339 format (e.g. 2026-06-21T00:00:00Z)."`
}

type GetEventParams struct {
	ID string `json:"id" jsonschema:"ID of the event to return."`
}

type SearchEventsOutput struct {
	Events []*EventSummaryOutput `json:"events"`
}

type EventSummaryOutput struct {
	ID       string       `json:"id" jsonschema:"ID of the event."`
	Title    string       `json:"title" jsonschema:"The title of the event"`
	Start    TZTimeOutput `json:"start" jsonschema:"The start time of the event"`
	End      TZTimeOutput `json:"end" jsonschema:"The end time of the event"`
	Location string       `json:"location" jsonschema:"The location of the event"`
	Status   string       `json:"status" jsonschema:"The status of the event (confirmed, tentative, canceled)"`
}

type EventOutput struct {
	ID          string                    `json:"id" jsonschema:"ID of the event."`
	Title       string                    `json:"title" jsonschema:"The title of the event"`
	Description string                    `json:"description" jsonschema:"The description of the event"`
	Start       TZTimeOutput              `json:"start" jsonschema:"The start time of the event"`
	End         TZTimeOutput              `json:"end" jsonschema:"The end time of the event"`
	Location    string                    `json:"location" jsonschema:"The location of the event"`
	Organizer   NamedEmailAddressOutput   `json:"organizer" jsonschema:"The organizer of the event"`
	Attendees   []NamedEmailAddressOutput `json:"attendees" jsonschema:"The attendees of the event"`
	IsAllDay    bool                      `json:"is_all_day" jsonschema:"Whether the event is an all day event or not"`
	Status      string                    `json:"status" jsonschema:"The status of the event (confirmed, tentative, canceled)"`
	UpdatedAt   time.Time                 `json:"updated_at" jsonschema:"The last time the event was updated (RFC3339 format)."`
}

func toEventSummaryOutputs(events []*domain.Event) *SearchEventsOutput {
	outputs := make([]*EventSummaryOutput, 0, len(events))
	for _, event := range events {
		output := &EventSummaryOutput{
			ID:       event.ID,
			Title:    event.Title,
			Start:    toTZTimeOutput(event.Start),
			End:      toTZTimeOutput(event.End),
			Location: event.Location,
			Status:   string(event.Status),
		}
		outputs = append(outputs, output)
	}
	return &SearchEventsOutput{Events: outputs}
}

func toEventOutput(event *domain.Event) *EventOutput {
	output := &EventOutput{
		ID:          event.ID,
		Title:       event.Title,
		Description: event.Description,
		Start:       toTZTimeOutput(event.Start),
		End:         toTZTimeOutput(event.End),
		Location:    event.Location,
		Organizer:   toNamedEmailAddressOutput(event.Organizer),
		Attendees:   toNamedEmailAddressOutputs(event.Attendees),
		IsAllDay:    event.IsAllDay,
		Status:      string(event.Status),
		UpdatedAt:   event.UpdatedAt,
	}
	return output
}

func addCreateEventTool(server *mcp.Server, provider domain.CalendarWriteProvider) {
	tool := &mcp.Tool{
		Name:        "createEvent",
		Description: "Creates a new calendar event. The created event details are returned. No attendees or recurrence are supported per safety policy. Requires write access (access_mode = read_write).",
	}
	handler := func(ctx context.Context, req *mcp.CallToolRequest, params *CreateEventParams) (*mcp.CallToolResult, any, error) {
		start, err := time.Parse(time.RFC3339, params.Start)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid start time: %w", err)
		}
		end, err := time.Parse(time.RFC3339, params.End)
		if err != nil {
			return nil, nil, fmt.Errorf("invalid end time: %w", err)
		}
		create := domain.EventCreate{
			Title:       params.Title,
			Description: params.Description,
			Start:       domain.NewTZTime(start, ""),
			End:         domain.NewTZTime(end, ""),
			Location:    params.Location,
			IsAllDay:    params.IsAllDay,
		}
		event, err := provider.CreateEvent(ctx, create)
		if err != nil {
			return nil, nil, err
		}
		return nil, toEventOutput(event), nil
	}
	mcp.AddTool(server, tool, handler)
}

type CreateEventParams struct {
	Title       string  `json:"title"                 jsonschema:"The title of the event."`
	Description *string `json:"description,omitempty"  jsonschema:"An optional description for the event."`
	Start       string  `json:"start"                  jsonschema:"The start time of the event (RFC3339 format, e.g. 2026-06-21T14:00:00Z)."`
	End         string  `json:"end"                    jsonschema:"The end time of the event (RFC3339 format, e.g. 2026-06-21T15:00:00Z)."`
	Location    *string `json:"location,omitempty"     jsonschema:"An optional location for the event."`
	IsAllDay    *bool   `json:"is_all_day,omitempty"   jsonschema:"Whether the event is an all day event. Defaults to false."`
}
