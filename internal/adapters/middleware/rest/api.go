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

package rest

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/tdrn-org/go-httpserver"
	"github.com/tdrn-org/pim-mcp/internal/adapters/pim"
	"github.com/tdrn-org/pim-mcp/internal/session/model"
)

type Runtime interface {
	Provider() pim.Provider
	BaseURL() *url.URL
	Logger() *slog.Logger
	Ping(ctx context.Context) error
	SessionCookie() *httpserver.CookieHandler
	GetSession(ctx context.Context, id string) (*model.Session, error)
	LookupSessionByAPIKey(ctx context.Context, apiKey string) (*model.Session, error)
	DeleteSession(ctx context.Context, id string) error
	LoginURL(ctx context.Context) (*url.URL, error)
}

type SessionInfo struct {
	ProviderName string         `json:"provider_name"`
	APIKey       string         `json:"api_key"`
	Credentials  CredentialInfo `json:"credentials"`
}

type CredentialInfo struct {
	Valid  bool      `json:"valid"`
	Expiry time.Time `json:"expiry"`
}

type loginRequest struct {
	APIKey string `json:"api_key"`
}

//	@title			PIM MCP Server REST API
//	@version		1.0
//	@description	MCP server providing Agent access to PIM services.

//	@contact.url	https://github.com/tdrn-org/pim-mcp

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:9123
//	@BasePath	/api/v1

type API struct {
	runtime Runtime
}

func NewAPI(runtime Runtime) *API {
	return &API{
		runtime: runtime,
	}
}

const basePath string = "/api/v1"
const PathPing string = basePath + "/ping"
const PathSession string = basePath + "/session"
const PathLogin string = basePath + "/login"

func (api *API) Mount(server *httpserver.Instance) {
	server.HandleFunc("GET "+PathPing, api.PingGet)
	server.HandleFunc("GET "+PathSession, api.SessionGet)
	server.HandleFunc("DELETE "+PathSession, api.SessionDelete)
	server.HandleFunc("POST "+PathLogin, api.LoginPost)
}

const responseOK string = "ok"
const responseServerError string = "server error"

// GET @BasePath/ping
//
//	@Summary		Ping the server
//	@Description	Ping the server to check general health
//	@Produce		text/plain
//	@Success		200	{string}	string	"ok"
//	@Failure		500	{string}	string	"server error"
//	@Router			/api/v1/ping [get]
func (api *API) PingGet(w http.ResponseWriter, r *http.Request) {
	err := api.runtime.Ping(r.Context())
	if err != nil {
		api.sendError(w, r, http.StatusInternalServerError, err)
		return
	}
	api.sendPlainTextResponse(w, r, http.StatusOK, responseOK)
}

// GET @BasePath/session
//
//	@Summary		Get the user session
//	@Description	Get the session for the current user
//	@Produce		json
//	@Success		200	{object}	SessionInfo
//	@Failure		500	{string}	string	"server error"
//	@Router			/api/v1/session [get]
func (api *API) SessionGet(w http.ResponseWriter, r *http.Request) {
	id, _ := api.runtime.SessionCookie().Get(r)
	session, err := api.runtime.GetSession(r.Context(), id)
	if err != nil {
		api.sendError(w, r, http.StatusInternalServerError, err)
		return
	}
	apiKey := ""
	if !session.APIKeyShown {
		apiKey = session.APIKey
	}
	provider := api.runtime.Provider()
	credenitalInfo, err := provider.CheckCredentials(session.Credentials)
	if err != nil {
		api.sendError(w, r, http.StatusInternalServerError, err)
		return
	}
	sessionInfo := &SessionInfo{
		ProviderName: api.runtime.Provider().Name(),
		APIKey:       apiKey,
		Credentials: CredentialInfo{
			Valid:  credenitalInfo.Valid,
			Expiry: credenitalInfo.Expiry,
		},
	}
	api.runtime.SessionCookie().Set(w, session.ID, false)
	api.sendApplicationJSONResponse(w, r, http.StatusOK, sessionInfo)
}

// DELETE @BasePath/session
//
//	@Summary		Delete the user session
//	@Description	Delete the session for the current user
//	@Produce		json
//	@Success		200	{string}	string	"ok"
//	@Failure		500	{string}	string	"server error"
//	@Router			/api/v1/session [delete]
func (api *API) SessionDelete(w http.ResponseWriter, r *http.Request) {
	sessionCookie := api.runtime.SessionCookie()
	id, ok := sessionCookie.Get(r)
	if ok {
		sessionCookie.Delete(w)
		err := api.runtime.DeleteSession(r.Context(), id)
		if err != nil {
			api.sendError(w, r, http.StatusInternalServerError, err)
			return
		}
	}
	api.sendPlainTextResponse(w, r, http.StatusOK, responseOK)
}

// POST @BasePath/login
//
//	@Summary		Initiate PIM provider login
//	@Description	Initiate PIM provider login for the current user. Optionally provide an api_key for session recovery.
//	@Accept			json
//	@Produce		json
//	@Param			body	body		loginRequest	false	"optional api_key for session recovery"
//	@Success		200		{object}	SessionInfo
//	@Success		302		{string}	string	""
//	@Failure		401		{string}	string	"invalid api_key"
//	@Failure		500		{string}	string	"server error"
//	@Router			/api/v1/login [post]
func (api *API) LoginPost(w http.ResponseWriter, r *http.Request) {
	// Try to parse optional api_key from JSON body
	var req loginRequest
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// If body can't be parsed, fall through to OAuth2 redirect
			req.APIKey = ""
		}
	}

	if req.APIKey != "" {
		// Session recovery via API key
		session, err := api.runtime.LookupSessionByAPIKey(r.Context(), req.APIKey)
		if err != nil {
			api.sendError(w, r, http.StatusInternalServerError, err)
			return
		}
		if session == nil {
			api.sendPlainTextResponse(w, r, http.StatusUnauthorized, "invalid api_key")
			return
		}
		provider := api.runtime.Provider()
		credenitalInfo, err := provider.CheckCredentials(session.Credentials)
		if err != nil {
			api.sendError(w, r, http.StatusInternalServerError, err)
			return
		}
		sessionInfo := &SessionInfo{
			ProviderName: provider.Name(),
			Credentials: CredentialInfo{
				Valid:  credenitalInfo.Valid,
				Expiry: credenitalInfo.Expiry,
			},
		}
		api.runtime.SessionCookie().Set(w, session.ID, false)
		api.sendApplicationJSONResponse(w, r, http.StatusOK, sessionInfo)
		return
	}

	// No API key — standard OAuth2 flow
	loginURL, err := api.runtime.LoginURL(r.Context())
	if err != nil {
		api.sendError(w, r, http.StatusInternalServerError, err)
		return
	}
	http.Redirect(w, r, loginURL.String(), http.StatusFound)
}

func (api *API) sendApplicationJSONResponse(w http.ResponseWriter, r *http.Request, status int, content any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	err := json.NewEncoder(w).Encode(content)
	if err != nil {
		api.runtime.Logger().Error("failed to send 'application/json' response", slog.String("path", r.URL.Path), slog.String("method", r.Method), slog.Any("err", err))
	}
}

func (api *API) sendPlainTextResponse(w http.ResponseWriter, r *http.Request, status int, content string) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(status)
	_, err := w.Write([]byte(content))
	if err != nil {
		api.runtime.Logger().Error("failed to send 'text/plain' response", slog.String("path", r.URL.Path), slog.String("method", r.Method), slog.Any("err", err))
	}
}

func (api *API) sendError(w http.ResponseWriter, r *http.Request, status int, cause error) {
	if cause != nil {
		api.runtime.Logger().Error("http handler failure", slog.String("path", r.URL.Path), slog.String("method", r.Method), slog.Any("err", cause))
	}
	http.Error(w, responseServerError, status)
}
