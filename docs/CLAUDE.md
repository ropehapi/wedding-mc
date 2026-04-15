# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

wedding-mc is a Go REST API backend for a wedding-planning SaaS (MVP). Frontend lives in a separate repo. User communicates in Portuguese.

## Spec-Driven Workflow

This project is built top-down from `.specs/`. Before writing code for any feature, read the relevant spec. The specs are authoritative — code should conform to them, not the other way around.

- `.specs/project/PROJECT.md`, `ROADMAP.md`, `STATE.md` — product vision, stack decisions, backlog, deferred ideas
- `.specs/design/architecture.md` — layered architecture, folder layout, **exact DB schema** (copy SQL from here into migrations verbatim), API response envelope
- `.specs/design/tasks.md` — 42 numbered tasks (T1–T42) organized in 8 phases with explicit dependencies. Each task has `What`, `Where`, `Depends on`, and `Done when` criteria. **This is the implementation checklist** — work through it in order, respecting the phase dependency graph at the top of the file
- `.specs/features/{auth,wedding,guests,gifts,public-page}/spec.md` — per-module functional requirements with IDs (AUTH-01, WED-03, GUEST-05, etc.) referenced from tasks

When the user says "continue implementing" or "next task," check `.specs/design/tasks.md` and the current git state to find where to resume.

## Commands

```bash
# Build everything
go build ./...

# Vet
go vet ./...

# Tidy deps (run after adding imports)
go mod tidy

# Local Postgres (required for running the app and e2e tests)
docker compose up -d postgres

# Run the API (after T30 is implemented)
go run ./cmd/api

# Run all tests
go test ./...

# Run a single test by name
go test ./internal/service/ -run TestAuth_Login -v

# Coverage (target ≥80% per tasks.md)
go test ./... -cover

# Regenerate Swagger docs (after editing handler annotations)
swag init -g cmd/api/main.go
```

Config is read from env vars (with `.env` loaded via godotenv in dev). `DATABASE_URL` and `JWT_SECRET` are required; everything else has defaults in `internal/config/config.go`.

## Architecture

Classic Go layered architecture with **manual dependency injection in `cmd/api/main.go`** — no DI framework, no magic. Wire everything up explicitly.

```
HTTP → middleware → handler → service → repository → Postgres
                                    ↘ storage (local|s3)
```

**Layer rules:**
- `internal/handler/` — parses HTTP, validates input with `go-playground/validator`, calls service, writes response via `handler/response.go` helpers. No business logic.
- `internal/service/` — business rules. Orchestrates repositories and `StorageService`. Returns typed `domain` errors (`ErrNotFound`, `ErrConflict`, etc.).
- `internal/repository/` — sqlx queries only. No business logic. Implements interfaces declared in `internal/domain/`.
- `internal/domain/` — structs (with `db:""` and `json:""` tags), enums as string types, repository interfaces, typed errors. **Interfaces live next to the structs they operate on**, not in a separate `interfaces/` package.
- `internal/middleware/` — Chi middlewares: logger (zerolog), recover, CORS, JWT auth. JWT middleware injects `userID` into request context via a typed key; handlers pull it with `middleware.UserIDFromContext(ctx)`.
- `internal/config/` — `Load()` + `NewDB()` + `RunMigrations()`. Migrations run automatically on startup via golang-migrate from the `migrations/` directory.

**Response envelope** (enforced by `internal/handler/response.go`):
- Success: `{"data": ...}`
- Error: `{"error": "code", "message": "..."}`
- Validation: `{"error": "validation_error", "details": [{"field": ..., "rule": ...}]}`

**Route layout** (all under `/v1`):
- `/v1/auth/*` — public (register, login, refresh, logout)
- `/v1/wedding/*`, `/v1/guests/*`, `/v1/gifts/*` — require JWT auth middleware
- `/v1/public/:slug/*` — public, no auth; **must never expose `reserved_by_name`, user email, or password hash**

**Gift reservation is the one place atomicity matters:** `GiftRepository.Reserve` uses `BEGIN ... SELECT FOR UPDATE ... UPDATE ... COMMIT` to prevent double-reservation. See T26 in tasks.md for the exact SQL.

## Conventions

- Module path: `github.com/ropehapi/wedding-mc`
- Responses in Portuguese where user-facing (error messages, validation messages)
- Slug generation for weddings: `generateSlug("Ana", "João")` → `"ana-e-joao"`, collisions get `-2`, `-3` suffix
- Passwords: bcrypt cost 12, never stored/logged/returned in plain text
- Refresh tokens: random value, stored as SHA-256 hash in `refresh_tokens`
- Photo uploads: max 10MB, images only; keys formatted as `weddings/{id}/{uuid}.{ext}`
- Test files co-located with source (`foo_test.go` next to `foo.go`); e2e tests use testcontainers-go with a real Postgres
