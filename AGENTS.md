# AGENTS.md

This file provides guidance to Kimi Code CLI when working with the `yarr` project.

## Project Overview

**yarr** (yet another rss reader) is a web-based RSS feed aggregator written in Go. It compiles to a single binary with an embedded SQLite database. Can run as both a desktop app (with system tray) and a self-hosted server.

- **Repository**: https://github.com/nkanaev/yarr
- **Go Module**: `github.com/nkanaev/yarr`
- **Go Version**: >= 1.23
- **Current Version**: 2.6

## Project Structure

| Path | Purpose |
|------|---------|
| `cmd/yarr/` | Main entry point |
| `src/assets/` | Frontend — HTML, JS, CSS (embedded via `//go:embed`) |
| `src/content/` | Content fetching/processing (sanitizer, readability, scraper) |
| `src/parser/` | Feed parsers (RSS, Atom, RDF, JSON Feed) |
| `src/platform/` | OS-specific code (system tray, browser open, version info) |
| `src/server/` | HTTP server, routing, auth, Fever API, OPML |
| `src/storage/` | SQLite storage layer (feeds, folders, items, migrations) |
| `src/systray/` | System tray GUI implementation |
| `src/worker/` | Background feed refresh worker |

## Build Requirements

- **Go**: >= 1.23
- **C Compiler**: GCC / Clang / MSVC (CGO_ENABLED=1 required for SQLite)
- **Zig**: >= 0.14.0 (optional, for cross-compilation)

## Essential Build Tags

All `go build` and `go test` commands **must** include:

```bash
-tags "sqlite_foreign_keys sqlite_json"
```

Additional tags:
- `debug` — Development mode: serves frontend assets from disk instead of embed
- `gui` — Enables desktop GUI (system tray) mode

## Common Commands

```bash
# Build for current platform
make host

# Development server with debug mode
make serve

# Run all tests
make test

# Run tests for a specific package
go test -tags "sqlite_foreign_keys sqlite_json" -v ./src/parser/

# Run a single test
go test -tags "sqlite_foreign_keys sqlite_json" -v ./src/parser/ -run TestRSSParse

# Cross-compilation examples (requires Zig)
make linux_amd64
make windows_amd64
```

## Coding Conventions

- **Go style**: Standard Go formatting (`gofmt`). Keep it simple and idiomatic.
- **Frontend**: Native JavaScript + Vue.js 2 + Bootstrap CSS. **No build tools** — edit files directly.
- **Templates**: Use `{% %}` as delimiters (Go template syntax) to avoid conflict with Vue.js `{{ }}`.
- **Database**: All schema migrations are managed in `src/storage/migration.go` in the `migrate()` function.
- **API routes**: RESTful style — GET (list/detail), POST (create), PUT (update), DELETE (remove).

## Key Architectural Notes

1. **Entry**: `cmd/yarr/main.go` parses CLI flags, initializes `storage.Storage` and `server.Server`
2. **Router**: Custom router in `src/server/router/` (not standard library); supports path params like `:id`
3. **Middleware**: gzip (`src/server/gzip/`), auth (`src/server/auth/`)
4. **Frontend assets**: Embedded via `//go:embed`. In `debug` mode, read directly from `src/assets/` directory
5. **Data flow**: Server starts `worker.Worker` which periodically refreshes feeds; frontend talks to `/api/*` endpoints
6. **Fever API**: Compatible endpoint implemented in `src/server/fever.go`

## Testing Guidelines

- Always use the required build tags when running tests.
- Tests should be simple and focused; table-driven tests preferred.
- The project uses standard `testing` package (no external test frameworks).

## Making Changes

- **Database changes**: Add new migration steps to `src/storage/migration.go`. Migrations run linearly on startup.
- **Frontend changes**: Edit files in `src/assets/` directly. Use `make serve` for development to see changes without rebuild.
- **API changes**: Update handlers in `src/server/` and corresponding frontend API calls in `src/assets/javascripts/api.js`.
- **New feed formats**: Add parser in `src/parser/` following existing patterns.

## Dependencies

Managed via Go modules. Key dependencies:
- `github.com/mattn/go-sqlite3` — SQLite driver
- `golang.org/x/net` — Extended networking libraries
- `golang.org/x/sys` — System calls
