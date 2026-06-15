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

func (p *Provider) SearchEmails(ctx context.Context, filter domain.EmailFilter) ([]*domain.Email, error) {
	emails := slices.Collect(maps.Values(emailData))
	slices.SortFunc(emails, application.EmailSortFunc)
	emailList := slices.Values(emails)
	return slices.Collect(emailList), nil
}

func (p *Provider) GetEmail(ctx context.Context, id string) (*domain.Email, error) {
	email, ok := emailData[id]
	if !ok {
		return nil, domain.ErrEntityNotFound
	}
	return email, nil
}

var emailData map[string]*domain.Email = map[string]*domain.Email{}
