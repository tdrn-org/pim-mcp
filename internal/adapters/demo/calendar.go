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
	"maps"
	"slices"

	"github.com/tdrn-org/pim-mcp/internal/application"
	"github.com/tdrn-org/pim-mcp/internal/domain"
)

func (p *Provider) SearchEvents(ctx context.Context, filter domain.EventFilter) ([]*domain.Event, error) {
	events := slices.Collect(maps.Values(eventData))
	slices.SortFunc(events, application.EventSortFunc)
	eventList := slices.Values(events)
	return slices.Collect(eventList), nil
}

func (p *Provider) GetEvent(ctx context.Context, id string) (*domain.Event, error) {
	event, ok := eventData[id]
	if !ok {
		return nil, domain.ErrEntityNotFound
	}
	return event, nil
}

var eventData map[string]*domain.Event = map[string]*domain.Event{}
