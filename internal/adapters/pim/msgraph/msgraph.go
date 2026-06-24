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

package msgraph

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/google/uuid"
	kiotaauth "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/tdrn-org/go-cache"
	"github.com/tdrn-org/go-cache/memory"
	"github.com/tdrn-org/go-httpserver"
	"github.com/tdrn-org/pim-mcp/config"
	"github.com/tdrn-org/pim-mcp/internal/adapters/middleware/auth"
	"github.com/tdrn-org/pim-mcp/internal/adapters/pim"
	"github.com/tdrn-org/pim-mcp/internal/domain"
	"github.com/tdrn-org/pim-mcp/internal/session/model"
	"github.com/thlib/go-timezone-local/tzlocal"
	"golang.org/x/oauth2"
)

const Name = "msgraph"

const DefaultSearchLimit int = 25

type Runtime interface {
	BaseURL() *url.URL
	Logger() *slog.Logger
	SessionCookie() *httpserver.CookieHandler
	LookupSession(ctx context.Context, id string) (*model.Session, error)
	LookupSessionByAPIKey(ctx context.Context, apiKey string) (*model.Session, error)
	UpdateSessionCredentials(ctx context.Context, id string, credentials string) error
}

const credentialCacheTTL time.Duration = time.Hour

type Provider struct {
	runtime         Runtime
	cfg             *config.ProviderConfig
	timeLocation    *time.Location
	runtimeTimezone string
	credentialCache cache.KeyValue[string, *credentialHolder]
	logger          *slog.Logger
}

type credentialHolder struct {
	UserToken       *oauth2.Token
	GraphCredential *azidentity.OnBehalfOfCredential
}

func NewProvider(runtime Runtime, cfg *config.ProviderConfig) (*Provider, error) {
	timezone, err := tzlocal.RuntimeTZ()
	if err != nil {
		slog.Warn("failed to detect runtime timezone, falling back to UTC", slog.Any("err", err))
		timezone = "UTC"
	}
	timeLocation, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, fmt.Errorf("failed to load runtime timezone '%s' (cause: %w)", timezone, err)
	}
	provider := &Provider{
		runtime:         runtime,
		cfg:             cfg,
		timeLocation:    timeLocation,
		runtimeTimezone: timezone,
		logger:          slog.With(slog.String("provider", Name)),
	}
	credentialCache, err := memory.NewKeyValue(0, credentialCacheTTL, provider.loadSessionCredential)
	if err != nil {
		return nil, err
	}
	provider.credentialCache = credentialCache
	return provider, nil
}

func (p *Provider) loadSessionCredential(ctx context.Context, sessionID string) (*credentialHolder, error) {
	sessionLogger := p.logger.With(slog.String("sessionID", sessionID))
	sessionLogger.Info("looking up session")
	session, err := p.runtime.LookupSession(ctx, sessionID)
	if err != nil {
		sessionLogger.Warn("failed to load session", slog.Any("err", err))
		return nil, cache.ErrNotFound
	}
	token, err := unmarshalToken(session.Credentials)
	if err != nil {
		sessionLogger.Warn("failed to unmarshal session credentials", slog.Any("err", err))
		return nil, cache.ErrNotFound
	}
	if !token.Valid() {
		sessionLogger.Info("refreshing session token...")
		refreshedToken, err := p.oauth2Config().TokenSource(ctx, &oauth2.Token{RefreshToken: token.RefreshToken}).Token()
		if err != nil {
			sessionLogger.Warn("failed to refresh session token", slog.Any("err", err))
			return nil, cache.ErrNotFound
		}
		token = refreshedToken
	}
	sessionLogger.Info("obtaining OBO credential...")
	credential, err := azidentity.NewOnBehalfOfCredentialWithSecret(p.cfg.MSGraph.TenantID, p.cfg.MSGraph.ClientID, token.AccessToken, p.cfg.MSGraph.ClientSecret, nil)
	if err != nil {
		sessionLogger.Warn("failed to create OBO credential", slog.Any("err", err))
		return nil, cache.ErrNotFound
	}
	cachedCredential := &credentialHolder{
		UserToken:       token,
		GraphCredential: credential,
	}
	return cachedCredential, nil
}

func (p *Provider) requestContextWithSession(r *http.Request) context.Context {
	ctx := r.Context()
	id, ok := p.runtime.SessionCookie().Get(r)
	if !ok {
		return ctx
	}
	session, _ := p.runtime.LookupSession(r.Context(), id)
	if session == nil {
		return ctx
	}
	return auth.ContextWithSession(ctx, session)
}

func (p *Provider) credentialFromContext(ctx context.Context) (*azidentity.OnBehalfOfCredential, error) {
	session := auth.SessionFromContext(ctx)
	if session == nil || session.Credentials == "" {
		return nil, domain.ErrAuthenticationRequired
	}
	cachedCredential, err := p.credentialCache.Get(ctx, session.ID)
	if err != nil {
		return nil, domain.ErrAuthenticationRequired
	}
	_, err = cachedCredential.GraphCredential.GetToken(ctx, graphTokenRequestOptions)
	if err != nil {
		p.logger.Warn("failed to get credential token", slog.Any("err", err))
		return nil, domain.ErrAuthenticationRequired
	}
	return cachedCredential.GraphCredential, nil
}

func (p *Provider) graphClient(ctx context.Context) (*msgraphsdk.GraphServiceClient, error) {
	credential, err := p.credentialFromContext(ctx)
	if err != nil {
		return nil, err
	}
	authProvider, err := kiotaauth.NewAzureIdentityAuthenticationProviderWithScopes(
		credential,
		[]string{
			graphScope,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Authentication Provider (cause: %w)", err)
	}
	adapter, err := msgraphsdk.NewGraphRequestAdapter(authProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create Graph request adapater (cause: %w)", err)
	}
	client := msgraphsdk.NewGraphServiceClient(adapter)
	return client, nil
}

func (p *Provider) handleLogin(w http.ResponseWriter, r *http.Request) {
	p.logger.Info("initiating OAuth2 authentication flow")
	state := uuid.NewString()
	redirect := p.oauth2Config().AuthCodeURL(state)
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (p *Provider) handleCallback(w http.ResponseWriter, r *http.Request) {
	p.logger.Info("completing OAuth2 authentication flow")
	code := r.URL.Query().Get("code")
	//TODO: Verify state
	//state := r.URL.Query().Get("state")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}
	token, err := p.oauth2Config().Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, "token exchange failed: "+err.Error(), http.StatusInternalServerError)
		return
	}
	credentials, err := marshalToken(token)
	if err != nil {
		http.Error(w, "failed to marshal credentials: "+err.Error(), http.StatusInternalServerError)
		return
	}
	id, _ := p.runtime.SessionCookie().Get(r)
	err = p.runtime.UpdateSessionCredentials(r.Context(), id, credentials)
	if err != nil {
		http.Error(w, "token exchange failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Credentials are now stored in the session DB.
	// MCP tools retrieve them per-call via SessionFromContext(ctx).

	http.Redirect(w, r, p.runtime.BaseURL().String(), http.StatusFound)
}

func (p *Provider) handleContacts(w http.ResponseWriter, r *http.Request) {
	ctx := p.requestContextWithSession(r)
	contacts, err := p.SearchContacts(ctx, domain.ContactFilter{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if len(contacts) > 0 {
		for _, contact := range contacts {
			fmt.Fprintf(w, "%s\n", contact)
		}
	} else {
		fmt.Fprintf(w, "no contacts found")
	}
}

func (p *Provider) handleEmails(w http.ResponseWriter, r *http.Request) {
	ctx := p.requestContextWithSession(r)
	emails, err := p.SearchEmails(ctx, domain.EmailFilter{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if len(emails) > 0 {
		for _, email := range emails {
			fmt.Fprintf(w, "%s\n", email)
		}
	} else {
		fmt.Fprintf(w, "no emails found")
	}
}

func (p *Provider) handleEvents(w http.ResponseWriter, r *http.Request) {
	ctx := p.requestContextWithSession(r)
	events, err := p.SearchEvents(ctx, domain.EventFilter{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if len(events) > 0 {
		for _, event := range events {
			fmt.Fprintf(w, "%s\n", event)
		}
	} else {
		fmt.Fprintf(w, "no events found")
	}
}

func (p *Provider) handleTasks(w http.ResponseWriter, r *http.Request) {
	ctx := p.requestContextWithSession(r)
	tasks, err := p.SearchTasks(ctx, domain.TaskFilter{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else if len(tasks) > 0 {
		for _, task := range tasks {
			fmt.Fprintf(w, "%s\n", task)
		}
	} else {
		fmt.Fprintf(w, "no tasks found")
	}
}

func (p *Provider) ID() string {
	return p.cfg.MSGraph.ClientID
}

func (*Provider) Name() string {
	return Name
}

func (p *Provider) Capabilities() domain.ProviderCapabilities {
	return domain.AllProviderCapabilities(domain.AccessMode(p.cfg.AccessMode))
}

func (p *Provider) Mount(server *httpserver.Instance) {
	server.HandleFunc("/msgraph/login", p.handleLogin)
	server.HandleFunc("/msgraph/callback", p.handleCallback)
	server.HandleFunc("/msgraph/contacts", p.handleContacts)
	server.HandleFunc("/msgraph/emails", p.handleEmails)
	server.HandleFunc("/msgraph/events", p.handleEvents)
	server.HandleFunc("/msgraph/tasks", p.handleTasks)
}

func (p *Provider) LoginURL() *url.URL {
	return p.runtime.BaseURL().JoinPath("/msgraph/login")
}

func (p *Provider) CheckCredentials(ctx context.Context, sessionID, credentials string) *pim.CredentialInfo {
	info := &pim.CredentialInfo{
		Valid: false,
	}
	if credentials == "" {
		return info
	}
	cachedCredential, err := p.credentialCache.Get(ctx, sessionID)
	if err != nil {
		return info
	}
	_, err = cachedCredential.GraphCredential.GetToken(ctx, graphTokenRequestOptions)
	if err != nil {
		p.logger.Warn("failed to get credential token", slog.Any("err", err))
		return info
	}
	info.Valid = true
	return info
}

func (p *Provider) RefreshCredentials(ctx context.Context, sessionID, credentials string, refreshInterval time.Duration) string {
	cachedCredential, err := p.credentialCache.Get(ctx, sessionID)
	if err != nil {
		p.logger.Warn("discarding outdated credentials", slog.Any("err", err))
		return ""
	}
	refreshedCredentials, err := marshalToken(cachedCredential.UserToken)
	if err != nil {
		p.logger.Warn("invalid cached credentials", slog.Any("err", err))
		return credentials
	}
	if refreshedCredentials != credentials {
		p.logger.Info("credentials refreshed")
	}
	return refreshedCredentials
}
