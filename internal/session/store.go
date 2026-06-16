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

package session

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/tdrn-org/go-database"
	"github.com/tdrn-org/pim-mcp/internal/session/model"
)

type Store struct {
	driver *database.Driver
}

func NewStore(driver *database.Driver) *Store {
	store := &Store{
		driver: driver,
	}
	return store
}

func (s *Store) Close() error {
	return s.driver.Close()
}

func (s *Store) Ping(ctx context.Context) error {
	return s.driver.Ping(ctx)
}

func (s *Store) GetSession(ctx context.Context, id string) (*model.Session, error) {
	txCtx, tx, err := s.driver.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	var session *model.Session
	if id != "" {
		session, err = model.SelectSession(txCtx, s.driver, id)
		if err != nil {
			return nil, err
		}
		if session != nil && !session.APIKeyShown {
			session.APIKeyShown = true
			err := session.Update(txCtx)
			if err != nil {
				return nil, err
			}
		}
	}
	if session == nil {
		session = model.NewSession(s.driver)
		err = session.Insert(txCtx)
		if err != nil {
			return nil, err
		}
	}

	err = tx.CommitTx(txCtx)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *Store) UpdateSessionCredentials(ctx context.Context, id string, credentials string) error {
	txCtx, tx, err := s.driver.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	session, err := model.SelectSession(txCtx, s.driver, id)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("unknown session id '%s'", id)
	}
	session.Credentials = credentials
	session.LastAccess = database.Now()
	err = session.Update(txCtx)
	if err != nil {
		return err
	}

	err = tx.CommitTx(txCtx)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) LookupSessionByAPIKey(ctx context.Context, apiKey string) (*model.Session, error) {
	txCtx, tx, err := s.driver.BeginTx(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	session, err := model.SelectSessionByAPIKey(txCtx, s.driver, apiKey)
	if err != nil {
		return nil, err
	}

	err = tx.CommitTx(txCtx)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *Store) DeleteSession(ctx context.Context, id string) error {
	txCtx, tx, err := s.driver.BeginTx(ctx)
	if err != nil {
		return err
	}
	defer tx.RollbackUncommitedTx(txCtx)

	session, err := model.SelectSession(txCtx, s.driver, id)
	if err != nil {
		return err
	}
	if session == nil {
		return fmt.Errorf("unknown session id '%s'", id)
	}
	err = session.Delete(txCtx)
	if err != nil {
		return err
	}

	return tx.CommitTx(txCtx)
}
