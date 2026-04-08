# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build        # Build API binary (build/xkcd)
make web          # Build web server binary (build/web-server)
make test         # Run unit tests with coverage
make lint         # Run golangci-lint
make format       # Format code with gofumpt + wsl
make sec          # Security checks (govulncheck + trivy)
make bench        # Run benchmarks
make integration  # Run integration tests
make e2e          # Run Python-based e2e tests
make e2e-playwright # Run Playwright browser e2e tests
```

Run a single test:
```bash
go test ./internal/core/services/... -run TestSearchService
go test ./... -run TestFoo -v
```

## Architecture

Hexagonal (ports-and-adapters) pattern with two entry points:
- `cmd/xkcd/` — REST API server (port 8080)
- `cmd/web/` — Web frontend server (port 8090, proxies to API)

**Layer flow:**

```
REST handlers / Web handlers
        ↓↑
internal/core/services/     ← business logic
internal/core/ports/        ← interfaces (repos)
        ↓↑
internal/adapters/repos/    ← SQLite + in-memory search index
db/                         ← SQLite setup & migrations
```

**Key packages:**
- `internal/core/services/` — Search, Fetcher, User, Comic services
- `internal/core/ports/` — Repository interfaces (ComicsRepo, SearchComicsRepo, UserRepo, ComicFetcherRepo)
- `internal/core/models/` — Comic, User structs
- `internal/adapters/repos/search/` — In-memory inverted index using snowball stemmer; indexes title, transcription, alt text, news
- `internal/adapters/repos/comic/` and `user/` — SQLite-backed repositories
- `internal/adapters/rest/` — HTTP handlers, JWT auth middleware, rate limiting
- `web/` — HTML template handlers + REST client (separate binary)
- `pkg/` — Standalone utilities: xkcd API client, stemmer/stop-words, JSONL, rate limiter, HTTP helpers

## REST API

| Method | Path | Auth | Notes |
|--------|------|------|-------|
| POST | `/api/login` | — | Returns JWT |
| GET | `/api/pics?search=…` | — | Search comics |
| GET | `/api/comic/{id}` | — | Single comic |
| POST | `/api/update` | Admin Bearer | Trigger re-index |

## Auth & Config

- JWT Bearer tokens, 3h default TTL, admin role required for `/api/update`
- Config loaded from `config.yaml`; test config in `test/config/`
- Scheduled index updates via cron (`fetcher.update_spec` in config)

## Testing layout

- Unit tests: `*_test.go` co-located with source
- Integration tests: `test/integration/`
- E2e (curl-based Python): `test/e2e/`
- E2e (browser): `test/e2e/playwright/`
- Test DB: `test/db/test.db`

## Unit testing requirements (coursework)

**Target:** ≥80% code coverage, minimum 25 unit tests per team member. Tests run automatically via `make test`.

**Test design techniques to apply** (must use several):
- Equivalence partitioning
- Boundary value analysis
- Pairwise testing

**Coverage goals:**

| Coverage | Points |
|----------|--------|
| 60–70%   | 10     |
| 71–80%   | 15     |
| 81%+     | 25     |

**When writing new tests**, follow this pattern:
1. Identify the unit under test (function/method/service)
2. Define equivalence classes for inputs (valid, invalid, edge)
3. Add boundary cases (empty input, max values, zero, nil)
4. Add one test function per logical scenario, named `Test<Unit>_<Scenario>`
5. Place the `_test.go` file alongside the source file it tests

**To extend the test suite for a new function/algorithm:**
- For a new service method: add cases in the corresponding `*_test.go` next to the service file; mock the port interface using a local struct implementing the interface
- For a new repo: add tests that use `test/db/test.db` and the test config from `test/config/`
- For a new HTTP handler: use `net/http/httptest` with a real router wired to a mock service

**Coverage report:** `make test` generates `build/cover.out`; view HTML with `go tool cover -html=build/cover.out`
