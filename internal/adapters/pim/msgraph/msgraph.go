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
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/google/uuid"
	kiotaauth "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/tdrn-org/go-httpserver"
	"github.com/tdrn-org/pim-mcp/config"
	"github.com/tdrn-org/pim-mcp/internal/domain"
	"golang.org/x/oauth2"
)

const Name = "msgraph"

const DefaultSearchLimit int = 25

type Runtime interface {
	BaseURL() *url.URL
	Logger() *slog.Logger
}

type Provider struct {
	runtime Runtime
	cfg     *config.MSGraphConfig
	//TODO: Store externally
	accessToken string
	logger      *slog.Logger
}

func NewProvider(runtime Runtime, cfg *config.MSGraphConfig) *Provider {
	return &Provider{
		runtime: runtime,
		cfg:     cfg,
		logger:  slog.With(slog.String("provider", Name)),
	}
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

func (p *Provider) graphClient() (*msgraphsdk.GraphServiceClient, error) {
	credential, err := azidentity.NewOnBehalfOfCredentialWithSecret(p.cfg.TenantID, p.cfg.ClientID, p.accessToken, p.cfg.ClientSecret, nil)
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

func (p *Provider) Mount(server *httpserver.Instance) {
	server.HandleFunc("/msgraph/login", p.handleLogin)
	server.HandleFunc("/msgraph/callback", p.handleCallback)
	server.HandleFunc("/msgraph/contacts", p.handleContacts)
	server.HandleFunc("/msgraph/emails", p.handleEmails)
	server.HandleFunc("/msgraph/events", p.handleEvents)
	server.HandleFunc("/msgraph/tasks", p.handleTasks)
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

	//TODO: Persistenz + Token-Management
	p.accessToken = token.AccessToken
	// refreshToken := token.RefreshToken

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
