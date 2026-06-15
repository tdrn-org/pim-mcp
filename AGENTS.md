# Agent Guidelines — pim-mcp

## Projekt
- Go-Backend: PIM MCP Server (MCP + REST API)
- SvelteKit 2+ Frontend unter `internal/web/`
- REST API: `/api/v1/session`, `/api/v1/login`, `/api/v1/ping`
- MCP Endpoint: `/mcp`
- Frontend wird von Go als embedded fs ausgeliefert (`//go:embed all:build/*`)

## Frontend-Entwicklung
- SvelteKit 5, TypeScript strict, TailwindCSS v4
- Dark-Mode (slate-950 Hintergrund, slate-100 Text)
- Design: Indigo-Akzente (brand-400/500/600), minimal, zentriert
- Drei-Zustands-UI: Loading → Nicht angemeldet → Angemeldet
- Static Adapter (SPA), ssr=false, prerender=false
- Keine i18n (Englisch only für v1)
- API-Calls zentral in `$lib/api.ts`
- TypeScript-Interfaces in `$lib/types.ts`, spiegeln Go-Structs

## Konventionen
- Jede Svelte-Komponente in eigener Datei
- Kein CSS außer Tailwind-Klassen (kein `<style>` Block)
- Keine Erklärungen zu Tailwind-Klassen in Commit-Messages — nur Ergebnisse
- `npm install --legacy-peer-deps` wegen vite 6 + SvelteKit peer dep mismatch
