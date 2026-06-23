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
	"time"

	kiota "github.com/microsoft/kiota-abstractions-go"
)

type headerBuilder struct {
	headers *kiota.RequestHeaders
}

func newHeaders() *headerBuilder {
	return &headerBuilder{headers: &kiota.RequestHeaders{}}
}

func (b *headerBuilder) WithDefaults() *headerBuilder {
	b.headers.Add("ConsistencyLevel", "eventual")
	return b
}

func (b *headerBuilder) WithPreferTextContentType() *headerBuilder {
	b.headers.Add("Prefer", "outlook.body-content-type=\"text\"")
	return b
}

func (b *headerBuilder) WithPreferTimezone(location *time.Location) *headerBuilder {
	timezone, mapped := mapLocationToWindowsTimezone(location)
	if mapped {
		b.headers.Add("Prefer", fmt.Sprintf("outlook.timezone=\"%s\"", timezone))
	}
	return b
}

func (b *headerBuilder) Headers() *kiota.RequestHeaders {
	return b.headers
}
