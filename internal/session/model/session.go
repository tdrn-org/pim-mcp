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
	_ "embed"

	"github.com/tdrn-org/go-database"
)

type Session struct {
	driver  *database.Driver
	ID      string
	Secrets string
}

func NewSession(driver *database.Driver, secrets string) *Session {
	return &Session{
		driver:  driver,
		ID:      database.NewID(),
		Secrets: secrets,
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
	err = row.Scan(&s.Secrets)
	if database.NoRows(err) {
		commitErr := tx.CommitTx(txCtx)
		if commitErr != nil {
			err = commitErr
		}
	}
	if err != nil {
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

	err = tx.ExecTx(txCtx, sessionInsertSQL, s.ID, s.Secrets)
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

	err = tx.ExecTx(txCtx, sessionUpdateSQL, s.Secrets, s.ID)
	if err != nil {
		return err
	}

	return tx.CommitTx(txCtx)
}
