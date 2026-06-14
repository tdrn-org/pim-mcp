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
	"fmt"
	"log/slog"
	"time"

	"github.com/tdrn-org/pim-mcp/internal/domain"
)

const dateTimeLayoutLong string = "2006-01-02T15:04:05.0000000"
const dateTimeLayoutShort string = "2006-01-02T15:04:05"

func ParseTZtime(dateTime, timezone *string, defaultLocation *time.Location) domain.TZTime {
	location := defaultLocation
	if timezone != nil && *timezone != "" {
		timezoneLocation, err := time.LoadLocation(*timezone)
		if err == nil {
			location = timezoneLocation
		}
	}
	parsedDateTime := time.Time{}
	if dateTime != nil && *dateTime != "" {
		layout := dateTimeLayoutLong
		if len(*dateTime) <= len(dateTimeLayoutShort) {
			layout = dateTimeLayoutShort
		}
		parsed, err := time.ParseInLocation(layout, *dateTime, location)
		if err == nil {
			parsedDateTime = parsed
		} else {
			slog.Warn("unable to parse Graph API date-time", slog.String("dateTime", *dateTime))
		}
	}
	return domain.NewTZTime(parsedDateTime, ptrString(timezone))
}

func int32Ptr(i int) *int32 {
	value := int32(i)
	return &value
}

func stringPtr(s string) *string {
	value := s
	return &value
}

func boolPtr(b bool) *bool {
	value := b
	return &value
}

func standardFilterPtr(sf domain.StandardFilter) (*string, *int32) {
	var search *string
	if sf.Query != nil && *sf.Query != "" {
		search = stringPtr(fmt.Sprintf("\"%s\"", *sf.Query))
	}
	var limit *int32
	if sf.Limit != nil && *sf.Limit > 0 {
		limit = int32Ptr(*sf.Limit)
	} else {
		limit = int32Ptr(DefaultSearchLimit)
	}
	return search, limit
}

func ptrString(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

func ptrBool(p *bool, v bool) bool {
	if p == nil {
		return v
	}
	return *p
}

func ptrTime(p *time.Time) time.Time {
	if p == nil {
		return time.Time{}
	}
	return *p
}
