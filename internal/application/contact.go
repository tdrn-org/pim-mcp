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

package application

import (
	"strings"

	"github.com/tdrn-org/pim-mcp/internal/domain"
)

func ContactFilterFunc(t *domain.Contact, query string) bool {
	return MatchString(t.DisplayName, query) ||
		MatchString(t.FirstName, query) ||
		MatchString(t.LastName, query) ||
		MatchStrings(t.Emails.Addresses(), query) ||
		MatchStrings(t.Phones.Numbers(), query) ||
		MatchStrings(t.Addresses.Addresses(), query)
}

func ContactSortFunc(c1, c2 *domain.Contact) int {
	displayName1 := c1.DisplayName
	if displayName1 == "" {
		displayName1 = c1.FirstName + " " + c1.LastName
	}
	displayName2 := c2.DisplayName
	if displayName2 == "" {
		displayName2 = c2.FirstName + " " + c2.LastName
	}
	return strings.Compare(strings.ToUpper(displayName1), strings.ToUpper(displayName2))
}
