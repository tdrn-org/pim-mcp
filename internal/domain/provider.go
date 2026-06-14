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

type ProviderCapabilities struct {
	Email    bool
	Calendar bool
	Tasks    bool
	Contacts bool
}

func AllProviderCapabilities() ProviderCapabilities {
	return ProviderCapabilities{
		Email:    true,
		Calendar: true,
		Tasks:    true,
		Contacts: true,
	}
}

type Provider interface {
	ID() string
	Name() string
	Capabilities() ProviderCapabilities
}
