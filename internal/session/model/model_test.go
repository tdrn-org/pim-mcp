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

package model_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tdrn-org/go-database"
	"github.com/tdrn-org/go-database/memory"
	"github.com/tdrn-org/pim-mcp/internal/session/model"
)

func TestSession(t *testing.T) {
	driver := newDatabase(t)
	defer driver.Close()

	// New
	session1 := model.NewSession(driver)

	// Insert
	err := session1.Insert(t.Context())
	require.NoError(t, err)

	// Select (by ID)
	session2, err := model.SelectSession(t.Context(), driver, session1.ID)
	require.NoError(t, err)
	require.Equal(t, session1, session2)

	// Select (by APIKey)
	session3, err := model.SelectSessionByAPIKey(t.Context(), driver, session1.APIKey)
	require.NoError(t, err)
	require.Equal(t, session1, session3)

	// Update
	session2.Credentials = "new credentials"
	err = session2.Update(t.Context())
	require.NoError(t, err)

	// Select all
	sessions, err := model.SelectSessions(t.Context(), driver)
	require.NoError(t, err)
	require.Len(t, sessions, 1)
}

func newDatabase(t *testing.T) *database.Driver {
	driver, err := database.Open(memory.NewConfig(model.SqliteSchemaScriptOption))
	require.NoError(t, err)
	from, to, err := driver.UpdateSchema(t.Context())
	require.NoError(t, err)
	require.Equal(t, database.SchemaNone, from)
	require.Equal(t, 1, to)
	return driver
}
