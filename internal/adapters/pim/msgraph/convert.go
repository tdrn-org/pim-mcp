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

	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/tdrn-org/pim-mcp/internal/domain"
)

const dateTimeLayoutLong string = "2006-01-02T15:04:05.0000000"
const dateTimeLayoutShort string = "2006-01-02T15:04:05"

var windowsTimezoneMapping = map[string]string{
	"UTC":                          "UTC",
	"GMT Standard Time":            "Europe/London",
	"W. Europe Standard Time":      "Europe/Berlin",
	"Central Europe Standard Time": "Europe/Belgrade",
	"E. Europe Standard Time":      "Europe/Helsinki",
	"Romance Standard Time":        "Europe/Paris",
	"Russian Standard Time":        "Europe/Moscow",
	"Pacific Standard Time":        "America/Los_Angeles",
	"Mountain Standard Time":       "America/Denver",
	"Central Standard Time":        "America/Chicago",
	"Eastern Standard Time":        "America/New_York",
	"China Standard Time":          "Asia/Shanghai",
	"Tokyo Standard Time":          "Asia/Tokyo",
}

func mapWindowsTimezoneToLocation(timezone string) (*time.Location, bool) {
	mappedTimezone, exists := windowsTimezoneMapping[timezone]
	if !exists {
		return nil, false
	}
	location, err := time.LoadLocation(mappedTimezone)
	return location, err == nil
}

var ianaTimezoneMapping = map[string]string{
	"UTC":                 "UTC",
	"Europe/London":       "GMT Standard Time",
	"Europe/Berlin":       "W. Europe Standard Time",
	"Europe/Vienna":       "W. Europe Standard Time",
	"Europe/Zurich":       "W. Europe Standard Time",
	"Europe/Belgrade":     "Central Europe Standard Time",
	"Europe/Helsinki":     "E. Europe Standard Time",
	"Europe/Paris":        "Romance Standard Time",
	"Europe/Moscow":       "Russian Standard Time",
	"America/Los_Angeles": "Pacific Standard Time",
	"America/Denver":      "Mountain Standard Time",
	"America/Chicago":     "Central Standard Time",
	"America/New_York":    "Eastern Standard Time",
	"Asia/Shanghai":       "China Standard Time",
	"Asia/Tokyo":          "Tokyo Standard Time",
}

func mapLocationToWindowsTimezone(location *time.Location) (string, bool) {
	windowsTimezone, exists := windowsTimezoneMapping[location.String()]
	return windowsTimezone, exists
}

func marshalTZTime(tzTime domain.TZTime) (*string, *string) {
	dateTime := tzTime.DateTime.Format(dateTimeLayoutLong)
	timezone := tzTime.Timezone
	if timezone == "" {
		ianaTimezone, mapped := ianaTimezoneMapping[time.Local.String()]
		if mapped {
			timezone = ianaTimezone
		} else {
			timezone = "UTC"
		}
	}
	return &dateTime, &timezone
}

func unmarshalTZTime(dateTime, timezone *string) domain.TZTime {
	location := time.Local
	if timezone != nil && *timezone != "" {
		ianaTimezone, mapped := windowsTimezoneMapping[*timezone]
		if !mapped {
			ianaTimezone = *timezone
		}
		timezoneLocation, err := time.LoadLocation(ianaTimezone)
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

func bodyTypePtr(bt models.BodyType) *models.BodyType {
	value := bt
	return &value
}

func taskStatusPtr(ts models.TaskStatus) *models.TaskStatus {
	value := ts
	return &value
}

func importancePtr(i models.Importance) *models.Importance {
	value := i
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
