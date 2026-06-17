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
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/google/uuid"
	kiotaauth "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/tdrn-org/go-httpserver"
	"github.com/tdrn-org/pim-mcp/config"
	"github.com/tdrn-org/pim-mcp/internal/adapters/middleware/mcp"
	"github.com/tdrn-org/pim-mcp/internal/adapters/pim"
	"github.com/tdrn-org/pim-mcp/internal/domain"
	"github.com/tdrn-org/pim-mcp/internal/session/model"
	"golang.org/x/oauth2"
)

const Name = "msgraph"

const DefaultSearchLimit int = 25

type Runtime interface {
	BaseURL() *url.URL
	Logger() *slog.Logger
	SessionCookie() *httpserver.CookieHandler
	UpdateSessionCredentials(ctx context.Context, id string, credentials string) error
	LookupSessionByAPIKey(ctx context.Context, apiKey string) (*model.Session, error)
}

func UnmarshalToken(s string) (*oauth2.Token, error) {
	token := &oauth2.Token{}
	err := json.NewDecoder(strings.NewReader(s)).Decode(token)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal OAuth2 token (cause: %w)", err)
	}
	return token, nil
}

func MarshalToken(token *oauth2.Token) (string, error) {
	buffer := &strings.Builder{}
	err := json.NewEncoder(buffer).Encode(token)
	if err != nil {
		return "", fmt.Errorf("failed to marshal OAuth2 token (cause: %w)", err)
	}
	return buffer.String(), nil
}

type Provider struct {
	runtime Runtime
	cfg     *config.MSGraphConfig
	logger  *slog.Logger
}

func NewProvider(runtime Runtime, cfg *config.MSGraphConfig) *Provider {
	return &Provider{
		runtime: runtime,
		cfg:     cfg,
		logger:  slog.With(slog.String("provider", Name)),
	}
}

func (p *Provider) accessTokenFromContext(ctx context.Context) (string, error) {
	session := mcp.SessionFromContext(ctx)
	if session == nil || session.Credentials == "" {
		return "", fmt.Errorf("no session credentials available")
	}
	token, err := UnmarshalToken(session.Credentials)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal session credentials: %w", err)
	}
	if !token.Valid() {
		return "", fmt.Errorf("session credentials expired")
	}
	return token.AccessToken, nil
}

func (p *Provider) oauth2Config() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     p.cfg.ClientID,
		ClientSecret: p.cfg.ClientSecret,
		RedirectURL:  p.runtime.BaseURL().JoinPath("/msgraph/callback").String(),
		Scopes: []string{
			"offline_access",
			fmt.Sprintf("api://%s/access_as_user", p.cfg.ClientID),
		},
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/authorize", p.cfg.TenantID),
			TokenURL: fmt.Sprintf("https://login.microsoftonline.com/%s/oauth2/v2.0/token", p.cfg.TenantID),
		},
	}
}

func (p *Provider) graphClient(ctx context.Context) (*msgraphsdk.GraphServiceClient, error) {
	accessToken, err := p.accessTokenFromContext(ctx)
	if err != nil {
		return nil, err
	}
	credential, err := azidentity.NewOnBehalfOfCredentialWithSecret(p.cfg.TenantID, p.cfg.ClientID, accessToken, p.cfg.ClientSecret, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create OBO credential (cause: %w)", err)
	}
	authProvider, err := kiotaauth.NewAzureIdentityAuthenticationProviderWithScopes(
		credential,
		[]string{
			"https://graph.microsoft.com/.default",
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
	credentials, err := MarshalToken(token)
	if err != nil {
		http.Error(w, "token exchange failed: "+err.Error(), http.StatusInternalServerError)
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
	contacts, err := p.SearchContacts(r.Context(), domain.ContactFilter{})
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
	emails, err := p.SearchEmails(r.Context(), domain.EmailFilter{})
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
	events, err := p.SearchEvents(r.Context(), domain.EventFilter{})
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
	tasks, err := p.SearchTasks(r.Context(), domain.TaskFilter{})
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
	return p.cfg.ClientID
}

func (*Provider) Name() string {
	return Name
}

func (p *Provider) Capabilities() domain.ProviderCapabilities {
	return domain.AllProviderCapabilities()
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

func (p *Provider) CheckCredentials(credentials string) (*pim.CredentialInfo, error) {
	info := &pim.CredentialInfo{
		Valid: false,
	}
	if credentials == "" {
		return info, nil
	}
	token, err := UnmarshalToken(credentials)
	if err != nil {
		return info, nil
	}
	info.Valid = token.Valid()
	info.Expiry = token.Expiry
	return info, nil
}

func (p *Provider) RefreshCredentials(credentials string) (string, error) {
	return credentials, nil
}
