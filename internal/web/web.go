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

package web

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:build/*
var buildFiles embed.FS

// BuildFS returns the embedded SvelteKit build output as an http.FileSystem,
// with SPA fallback: unmatched paths serve index.html for client-side routing.
func BuildFS() http.FileSystem {
	sub, err := fs.Sub(buildFiles, "build")
	if err != nil {
		panic("web: embedded build directory not found — run 'npm run build' first")
	}
	return &spaFS{inner: http.FS(sub)}
}

// spaFS wraps an http.FileSystem with SPA fallback behavior.
// If a requested file does not exist, index.html is served instead,
// allowing SvelteKit client-side routing to handle the path.
type spaFS struct {
	inner http.FileSystem
}

func (s *spaFS) Open(name string) (http.File, error) {
	// Try the exact path first
	f, err := s.inner.Open(name)
	if err == nil {
		return f, nil
	}
	// Fall back to index.html for SPA client-side routing
	return s.inner.Open("index.html")
}

// Handler returns an http.Handler that serves the embedded SvelteKit SPA.
func Handler() http.Handler {
	return http.FileServer(BuildFS())
}

// Mount registers the SPA handler on the given serve mux at the root path.
// Call this after all API routes are registered so the SPA catch-all
// only handles unmatched paths.
func Mount(mux *http.ServeMux) {
	// Strip "build/" prefix from the embedded filesystem
	sub, err := fs.Sub(buildFiles, "build")
	if err != nil {
		panic("web: embedded build directory not found — run 'npm run build' first")
	}
	fileServer := http.FileServer(http.FS(sub))
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only handle GET/HEAD for SPA fallback; let API routes pass through
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}
		// Try serving the exact file first
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		f, err := sub.Open(path)
		if err != nil {
			// File not found — serve index.html for SPA client-side routing
			r.URL.Path = "/index.html"
		} else {
			f.Close()
		}
		fileServer.ServeHTTP(w, r)
	}))
}
