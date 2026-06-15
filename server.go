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

package pimmcp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/tdrn-org/go-database"
	"github.com/tdrn-org/go-database/memory"
	"github.com/tdrn-org/go-database/sqlite"
	"github.com/tdrn-org/go-httpserver"
	"github.com/tdrn-org/go-tlsconf/tlsclient"
	"github.com/tdrn-org/pim-mcp/config"
	"github.com/tdrn-org/pim-mcp/internal/adapters/middleware/mcp"
	"github.com/tdrn-org/pim-mcp/internal/adapters/middleware/rest"
	"github.com/tdrn-org/pim-mcp/internal/adapters/pim"
	"github.com/tdrn-org/pim-mcp/internal/adapters/pim/demo"
	"github.com/tdrn-org/pim-mcp/internal/adapters/pim/msgraph"
	"github.com/tdrn-org/pim-mcp/internal/session"
	"github.com/tdrn-org/pim-mcp/internal/session/model"
)

type Server struct {
	cfg        *config.Config
	store      *session.Store
	httpServer *httpserver.Instance
	api        *rest.API
	baseURL    *url.URL
	logger     *slog.Logger
}

func StartServer(ctx context.Context, cfg *config.Config) (*Server, error) {
	// Setup early logger with configuration address (which may not be the final one).
	// We will reset the logger after listener has been created.
	earlyLogger := slog.With(slog.String("server", cfg.Server.Address))
	s := &Server{
		cfg:    cfg,
		logger: earlyLogger,
	}
	startFuncs := []func(context.Context, *config.Config) error{
		s.startStore,
		s.startHttpServer,
		s.startRestAPI,
		s.startMCPServer,
	}
	for _, startFunc := range startFuncs {
		err := startFunc(ctx, cfg)
		if err != nil {
			defer s.Close()
			return nil, err
		}
	}
	return s, nil
}

func (s *Server) Run(ctx context.Context) error {
	s.logger.Info("serving HTTP requests...")
	err := s.httpServer.Serve()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	shutdownFuncs := []func(context.Context) error{
		s.shutdownHttpServer,
	}
	shutdownErrs := make([]error, 0, len(shutdownFuncs))
	for _, shutdownFunc := range shutdownFuncs {
		shutdownErrs = append(shutdownErrs, shutdownFunc(ctx))
	}
	return errors.Join(shutdownErrs...)
}

func (s *Server) Close() error {
	closeFuncs := []func() error{
		s.closeHttpServer,
		s.closeStore,
	}
	closeErrs := make([]error, 0, len(closeFuncs))
	for _, closeFunc := range closeFuncs {
		closeErrs = append(closeErrs, closeFunc())
	}
	return errors.Join(closeErrs...)
}

func (s *Server) Ping(ctx context.Context) error {
	if s.httpServer == nil {
		return fmt.Errorf("server not started")
	}
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsclient.GetConfig(),
		},
	}
	pingURL := s.httpServer.BaseURL().JoinPath(rest.PathPing).String()
	rsp, err := client.Get(pingURL)
	if err != nil {
		return fmt.Errorf("failed to access URL: '%s' (cause: %w)", pingURL, err)
	}
	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to ping URL: '%s' (status: %s)", pingURL, rsp.Status)
	}
	return nil
}

func (s *Server) startStore(ctx context.Context, cfg *config.Config) error {
	var databaseConfig database.Config
	switch cfg.Store.DatabaseType {
	case config.DatabaseType(memory.Type):
		databaseConfig = memory.NewConfig(model.SqliteSchemaScriptOption)
	case config.DatabaseType(sqlite.Type):
		databaseConfig = sqlite.NewConfig(cfg.Store.SQLiteConfig.File, sqlite.ModeRWC, model.SqliteSchemaScriptOption)
	default:
		return fmt.Errorf("unrecognized store type '%s'", cfg.Store.DatabaseType)
	}
	driver, err := database.Open(databaseConfig)
	if err != nil {
		return err
	}
	_, _, err = driver.UpdateSchema(ctx)
	if err != nil {
		return errors.Join(err, driver.Close())
	}
	s.store = session.NewStore(driver)
	return nil
}

func (s *Server) closeStore() error {
	if s.store == nil {
		return nil
	}
	return s.store.Close()
}

func (s *Server) startHttpServer(ctx context.Context, cfg *config.Config) error {
	s.logger.Info("starting HTTP server...")
	httpServerOptions := httpServerOptions(&cfg.Server)
	httpServer, err := httpserver.Listen(ctx, "tcp", cfg.Server.Address, httpServerOptions...)
	if err != nil {
		return err
	}
	s.httpServer = httpServer
	if cfg.Server.PublicURL.URL != nil {
		s.baseURL = cfg.Server.PublicURL.URL
	} else {
		s.baseURL = httpServer.BaseURL()
	}
	// Replace early logger by one attributed with actual URL
	s.logger = slog.With(slog.String("baseURL", s.baseURL.String()))
	return nil
}

func (s *Server) shutdownHttpServer(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) closeHttpServer() error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Close()
}

func (s *Server) startRestAPI(_ context.Context, _ *config.Config) error {
	runtime := &serverRuntime{server: s}
	s.api = rest.NewAPI(runtime)
	s.api.Mount(s.httpServer)
	return nil
}

func (s *Server) startMCPServer(ctx context.Context, cfg *config.Config) error {
	runtime := &serverRuntime{server: s}
	var adapter pim.Adapter
	switch cfg.Provider.Adapter {
	case config.ProviderAdapterDemo:
		adapter = demo.NewProvider()
	case config.ProviderAdapterMSGraph:
		msgraphProvider := msgraph.NewProvider(runtime, &cfg.Provider.MSGraph)
		msgraphProvider.Mount(s.httpServer)
		adapter = msgraphProvider
	default:
		return fmt.Errorf("unrecognized provider adapter '%s'", cfg.Provider.Adapter)
	}
	handler := mcp.NewHandler(runtime, adapter)
	s.httpServer.Handle("/mcp", handler)
	return nil
}

func (s *Server) ping(ctx context.Context) error {
	err := s.store.Ping(ctx)
	if err != nil {
		s.logger.Warn("store ping failure", slog.Any("err", err))
		return err
	}
	return nil
}

type serverRuntime struct {
	server *Server
}

func (runtime *serverRuntime) BaseURL() *url.URL {
	return runtime.server.baseURL
}

func (runtime *serverRuntime) Logger() *slog.Logger {
	return runtime.server.logger
}

func (runtime *serverRuntime) Ping(ctx context.Context) error {
	return runtime.server.ping(ctx)
}

func (runtime *serverRuntime) GetSession(ctx context.Context) (*rest.SessionInfo, error) {
	return &rest.SessionInfo{
		ProviderName: string(runtime.server.cfg.Provider.Adapter),
		LoggedIn:     false,
	}, nil
}

func (runtime *serverRuntime) DeleteSession(ctx context.Context) error {
	return nil
}

func (runtime *serverRuntime) LoginURL(ctx context.Context) (*url.URL, error) {
	return nil, nil
}
