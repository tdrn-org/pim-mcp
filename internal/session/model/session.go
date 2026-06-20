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

package model

import (
	"context"
	"crypto/rand"
	_ "embed"
	"encoding/base64"

	"github.com/tdrn-org/go-database"
)

type Session struct {
	driver      *database.Driver
	ID          string `db:"id"`
	APIKey      string `db:"api_key"`
	APIKeyShown bool   `db:"api_key_shown"`
	Credentials string `db:"credentials"`
	LastUpdate  int64  `db:"last_update"`
}

const sessionAPIKeyLenght int = 48
const sessionAPIKeyPrefix string = "pim_mcp_"

func NewSession(driver *database.Driver) *Session {
	apiKeyBytes := make([]byte, sessionAPIKeyLenght)
	rand.Read(apiKeyBytes)
	apiKey := sessionAPIKeyPrefix + base64.RawURLEncoding.EncodeToString(apiKeyBytes)
	return &Session{
		driver:      driver,
		ID:          database.NewID(),
		APIKey:      apiKey,
		APIKeyShown: false,
		Credentials: "",
	}
}

//go:embed session.select.sql
var sessionSelectSQL string

func SelectSession(ctx context.Context, driver *database.Driver, id string) (*Session, error) {
	txCtx, tx, err := driver.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	row, err := tx.QueryRowTx(txCtx, sessionSelectSQL, id)
	if err != nil {
		return nil, err
	}
	s := &Session{
		driver: driver,
		ID:     id,
	}
	err = database.ScanRow(row, s, "api_key", "api_key_shown", "credentials", "last_update")
	if database.NoRows(err) {
		s = nil
		err = nil
	} else if err != nil {
		return nil, err
	}

	err = tx.CommitTx(txCtx)
	if err != nil {
		return nil, err
	}
	return s, nil
}

//go:embed session.select_all.sql
var sessionSelectAllSQL string

func SelectSessions(ctx context.Context, driver *database.Driver) ([]*Session, error) {
	txCtx, tx, err := driver.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	rows, err := tx.QueryTx(txCtx, sessionSelectAllSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	sessions := make([]*Session, 0)
	for rows.Next() {
		s := &Session{
			driver: driver,
		}
		err = database.Scan(rows, s)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}

	err = tx.CommitTx(txCtx)
	if err != nil {
		return nil, err
	}
	return sessions, nil
}

//go:embed session.select_by_api_key.sql
var sessionSelectByAPIKeySQL string

func SelectSessionByAPIKey(ctx context.Context, driver *database.Driver, apiKey string) (*Session, error) {
	txCtx, tx, err := driver.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	row, err := tx.QueryRowTx(txCtx, sessionSelectByAPIKeySQL, apiKey)
	if err != nil {
		return nil, err
	}
	s := &Session{
		driver: driver,
		APIKey: apiKey,
	}
	err = database.ScanRow(row, s, "id", "api_key_shown", "credentials", "last_update")
	if database.NoRows(err) {
		s = nil
		err = nil
	} else if err != nil {
		return nil, err
	}

	err = tx.CommitTx(txCtx)
	if err != nil {
		return nil, err
	}
	return s, nil
}

//go:embed session.insert.sql
var sessionInsertSQL string

func (s *Session) Insert(ctx context.Context) error {
	txCtx, tx, err := s.driver.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	s.LastUpdate = database.Now()
	err = tx.ExecTx(txCtx, sessionInsertSQL, s.ID, s.APIKey, s.APIKeyShown, s.Credentials, s.LastUpdate)
	if err != nil {
		return err
	}

	return tx.CommitTx(txCtx)
}

//go:embed session.update.sql
var sessionUpdateSQL string

func (s *Session) Update(ctx context.Context) error {
	txCtx, tx, err := s.driver.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	s.LastUpdate = database.Now()
	err = tx.ExecTx(txCtx, sessionUpdateSQL, s.APIKeyShown, s.Credentials, s.LastUpdate, s.ID)
	if err != nil {
		return err
	}

	return tx.CommitTx(txCtx)
}

//go:embed session.delete.sql
var sessionDeleteSQL string

func (s *Session) Delete(ctx context.Context) error {
	txCtx, tx, err := s.driver.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	err = tx.ExecTx(txCtx, sessionDeleteSQL, s.ID)
	if err != nil {
		return err
	}

	return tx.CommitTx(txCtx)
}

func DeleteSession(ctx context.Context, driver *database.Driver, id string) error {
	txCtx, tx, err := driver.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	err = tx.ExecTx(txCtx, sessionDeleteSQL, id)
	if err != nil {
		return err
	}

	return tx.CommitTx(txCtx)
}
