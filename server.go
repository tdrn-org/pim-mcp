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
	"sync"
	"time"

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
	"github.com/tdrn-org/pim-mcp/internal/web"
)

const serverJobTickerSchedule time.Duration = 5 * time.Minute

type Server struct {
	cfg                 *config.Config
	store               *session.Store
	httpServer          *httpserver.Instance
	sessionCookie       *httpserver.CookieHandler
	provider            pim.Provider
	api                 *rest.API
	baseURL             *url.URL
	jobTicker           *time.Ticker
	jobTickerShutdown   chan any
	jobTickerShutdownWG sync.WaitGroup

	logger *slog.Logger
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
		s.startMCPServer,
		s.startRestAPI,
		s.startUI,
		s.startJobTicker,
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
		s.shutdownJobTicker,
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
	s.sessionCookie = &httpserver.CookieHandler{
		Name:   "pim-mcp-session",
		Path:   "/",
		Secure: s.baseURL.Scheme == "https",
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

func (s *Server) startMCPServer(ctx context.Context, cfg *config.Config) error {
	runtime := &serverRuntime{server: s}
	switch cfg.Provider.Adapter {
	case config.ProviderAdapterDemo:
		s.provider = demo.NewProvider(runtime)
	case config.ProviderAdapterMSGraph:
		s.provider = msgraph.NewProvider(runtime, &cfg.Provider.MSGraph)
	default:
		return fmt.Errorf("unrecognized provider adapter '%s'", cfg.Provider.Adapter)
	}
	s.provider.Mount(s.httpServer)
	handler := mcp.NewHandler(runtime, s.provider)
	s.httpServer.Handle("/mcp", handler)
	return nil
}

func (s *Server) startRestAPI(_ context.Context, _ *config.Config) error {
	runtime := &serverRuntime{server: s}
	s.api = rest.NewAPI(runtime)
	s.api.Mount(s.httpServer)
	return nil
}

func (s *Server) startUI(_ context.Context, _ *config.Config) error {
	// Landing page at / — checks for existing session and redirects to /session
	s.httpServer.HandleFunc("/", s.handleLanding)
	// SPA at /session — SvelteKit UI, all paths served by single-page app
	s.httpServer.HandleFunc("/session/", web.ServeSPA)
	return nil
}

func (s *Server) startJobTicker(_ context.Context, _ *config.Config) error {
	schedule := serverJobTickerSchedule
	s.jobTicker = time.NewTicker(schedule)
	s.jobTickerShutdown = make(chan any)
	slog.Info("starting job ticker", slog.String("schedule", schedule.String()))
	s.jobTickerShutdownWG.Go(func() {
		for stopped := false; !stopped; {
			select {
			case <-s.jobTickerShutdown:
				stopped = true
			case <-s.jobTicker.C:
				s.runJobs()
			}
		}
		slog.Info("job ticker stopped")
	})
	return nil
}

func (s *Server) shutdownJobTicker(_ context.Context) error {
	s.jobTicker.Stop()
	s.jobTickerShutdown <- true
	s.jobTickerShutdownWG.Wait()
	return nil
}

func (s *Server) runJobs() {

}

type serverRuntime struct {
	server *Server
}

func (runtime *serverRuntime) Provider() pim.Provider {
	return runtime.server.provider
}

func (runtime *serverRuntime) BaseURL() *url.URL {
	return runtime.server.baseURL
}

func (runtime *serverRuntime) Logger() *slog.Logger {
	return runtime.server.logger
}

func (runtime *serverRuntime) Ping(ctx context.Context) error {
	err := runtime.server.store.Ping(ctx)
	if err != nil {
		runtime.server.logger.Warn("store ping failure", slog.Any("err", err))
		return err
	}
	return nil
}

func (runtime *serverRuntime) SessionCookie() *httpserver.CookieHandler {
	return runtime.server.sessionCookie
}

func (runtime *serverRuntime) GetSession(ctx context.Context, id string) (*model.Session, error) {
	return runtime.server.store.GetSession(ctx, id)
}

func (runtime *serverRuntime) UpdateSessionCredentials(ctx context.Context, id string, credentials string) error {
	return runtime.server.store.UpdateSessionCredentials(ctx, id, credentials)
}

func (runtime *serverRuntime) LookupSessionByAPIKey(ctx context.Context, apiKey string) (*model.Session, error) {
	return runtime.server.store.LookupSessionByAPIKey(ctx, apiKey)
}

func (runtime *serverRuntime) DeleteSession(ctx context.Context, id string) error {
	return runtime.server.store.DeleteSession(ctx, id)
}

func (runtime *serverRuntime) LoginURL(ctx context.Context) (*url.URL, error) {
	return runtime.Provider().LoginURL(), nil
}

// handleLanding serves the landing page at GET /.
// If a valid session cookie exists, redirects to /session.
// Otherwise, serves the prerendered landing page (pure HTML, no JS).
func (s *Server) handleLanding(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		web.Handler().ServeHTTP(w, r)
		return
	}
	// Check for existing session cookie
	id, ok := s.sessionCookie.Get(r)
	if ok {
		session, err := s.store.GetSession(r.Context(), id)
		if err == nil && session != nil {
			http.Redirect(w, r, "/session", http.StatusFound)
			return
		}
	}
	// No session — serve the prerendered landing page
	web.Handler().ServeHTTP(w, r)
}
