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

package mcp

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/tdrn-org/pim-mcp/internal/adapters/middleware/auth"
	"github.com/tdrn-org/pim-mcp/internal/buildinfo"
	"github.com/tdrn-org/pim-mcp/internal/domain"
	"github.com/tdrn-org/pim-mcp/internal/session/model"
)

const Path string = "/mcp"

type Runtime interface {
	BaseURL() *url.URL
	Logger() *slog.Logger
	LookupSessionByAPIKey(ctx context.Context, apiKey string) (*model.Session, error)
}

func NewHandler(runtime Runtime, provider domain.Provider) http.Handler {
	impl := &mcp.Implementation{
		Name:       buildinfo.Cmd(),
		Version:    buildinfo.Version(),
		WebsiteURL: runtime.BaseURL().String(),
	}
	options := &mcp.ServerOptions{
		Logger: runtime.Logger(),
	}
	server := mcp.NewServer(impl, options)

	// Receiving middleware: validate API key and inject session into context
	server.AddReceivingMiddleware(func(next mcp.MethodHandler) mcp.MethodHandler {
		return func(ctx context.Context, method string, req mcp.Request) (mcp.Result, error) {
			session := auth.SessionFromContext(ctx)
			if session != nil {
				runtime.Logger().Debug("mcp request with valid user session",
					slog.String("method", method),
					slog.String("session_id", session.ID))
			} else {
				runtime.Logger().Warn("mcp request without user session")
			}
			// Allow unauthenticated access for backward compatibility
			return next(ctx, method, req)
		}
	})

	capabilities := provider.Capabilities()
	if capabilities.Email {
		addEmailTools(server, capabilities, provider.(domain.EmailProvider))
	}
	if capabilities.Calendar {
		addCalendarTools(server, capabilities, provider.(domain.CalendarProvider))
	}
	if capabilities.Tasks {
		addTaskTools(server, capabilities, provider.(domain.TaskProvider))
	}
	if capabilities.Contacts {
		addContactTools(server, provider.(domain.ContactProvider))
	}

	getServerFromRequest := func(r *http.Request) *mcp.Server {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != "" {
			session, err := runtime.LookupSessionByAPIKey(r.Context(), apiKey)
			if err == nil && session != nil {
				// Inject session into the request context so middleware and tools can access it
				ctx := auth.ContextWithSession(r.Context(), session)
				*r = *r.WithContext(ctx)
			}
		}
		return server
	}
	return mcp.NewStreamableHTTPHandler(getServerFromRequest, nil)
}
