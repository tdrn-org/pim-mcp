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

	"github.com/tdrn-org/go-httpserver"
)

type Runtime interface {
	BaseURL() *url.URL
	Logger() *slog.Logger
	Ping(ctx context.Context) error
	GetSession(ctx context.Context, id string) (string, *SessionInfo, error)
	DeleteSession(ctx context.Context, id string) error
	LoginURL(ctx context.Context) (*url.URL, error)
}

type SessionInfo struct {
	ProviderName string `json:"provider_name"`
	LoggedIn     bool   `json:"logged_in"`
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
	runtime       Runtime
	sessionCookie *httpserver.CookieHandler
}

func NewAPI(runtime Runtime) *API {
	sessionCookie := &httpserver.CookieHandler{
		Name:   "pim-mcp-session",
		Path:   "/",
		Secure: runtime.BaseURL().Scheme == "https",
	}
	return &API{
		runtime:       runtime,
		sessionCookie: sessionCookie,
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
	//TODO: Make it POST
	server.HandleFunc("GET "+PathLogin, api.LoginPost)
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
	sessionId, _ := api.sessionCookie.Get(r)
	sessionId, sessionInfo, err := api.runtime.GetSession(r.Context(), sessionId)
	if err != nil {
		api.sendError(w, r, http.StatusInternalServerError, err)
		return
	}
	api.sessionCookie.Set(w, sessionId, false)
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
	sessionId, ok := api.sessionCookie.Get(r)
	if ok {
		err := api.runtime.DeleteSession(r.Context(), sessionId)
		if err != nil {
			api.sendError(w, r, http.StatusInternalServerError, err)
			return
		}
		api.sessionCookie.Delete(w)
	}
	api.sendPlainTextResponse(w, r, http.StatusOK, responseOK)
}

// POST @BasePath/login
//
//	@Summary		Initiate PIM provider login
//	@Description	Initiate PIM provider login for the current user
//	@Accept			json
//	@Produce		text/plain
//	@Param			api_key	formData	string	false	"login using api_key"
//	@Success		302		{string}	string	""
//	@Failure		500		{string}	string	"server error"
//	@Router			/api/v1/login [post]
func (api *API) LoginPost(w http.ResponseWriter, r *http.Request) {
	loginURL, err := api.runtime.LoginURL(r.Context())
	if err != nil {
		//TODO
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
