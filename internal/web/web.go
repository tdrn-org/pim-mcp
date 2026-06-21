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
	"sync"
)

//go:embed all:build/*
var buildFiles embed.FS

var (
	fileServerOnce sync.Once
	fileServer     http.Handler
	spaHTMLOnce    sync.Once
	spaHTML        []byte
)

// BuildFS returns the embedded SvelteKit build output as an http.FileSystem,
// with SPA fallback: unmatched paths serve spa.html for client-side routing.
func BuildFS() http.FileSystem {
	sub, err := fs.Sub(buildFiles, "build")
	if err != nil {
		panic("web: embedded build directory not found — run 'npm run build' first")
	}
	return &spaFS{inner: http.FS(sub)}
}

// spaFS wraps an http.FileSystem with SPA fallback behavior.
// If a requested file does not exist, spa.html is served instead,
// allowing SvelteKit client-side routing to handle the path.
type spaFS struct {
	inner http.FileSystem
}

func (s *spaFS) Open(name string) (http.File, error) {
	f, err := s.inner.Open(name)
	if err == nil {
		return f, nil
	}
	return s.inner.Open("spa.html")
}

// Handler returns an http.Handler that serves the embedded SvelteKit SPA.
// The http.FileServer is created once and cached for all subsequent calls.
func Handler() http.Handler {
	fileServerOnce.Do(func() {
		fileServer = http.FileServer(BuildFS())
	})
	return fileServer
}

// ServeSPA serves the SvelteKit SPA shell (spa.html) for any request.
// The HTML is read once from the embedded filesystem and cached.
// Use this as a HandleFunc for SPA routes like "/session/".
func ServeSPA(w http.ResponseWriter, r *http.Request) {
	spaHTMLOnce.Do(func() {
		var err error
		spaHTML, err = buildFiles.ReadFile("build/spa.html")
		if err != nil {
			panic("web: spa shell not found — run 'npm run build' first")
		}
	})
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(spaHTML)
}
