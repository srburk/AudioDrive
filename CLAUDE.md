# AudioDrive

A Go HTTP service that accepts URLs from users and persists them to PostgreSQL.

## Design Philosophy

- **Interface-isolated storage** — nothing outside `internal/store` ever touches SQL. Handlers receive a `store.URLStore` interface; `sql.ErrNoRows` is translated to `store.ErrNotFound` inside the implementation.
- **Standard library first** — no external router, no ORM. Go 1.22+ `ServeMux` handles method+path routing. Only external dependency: `github.com/lib/pq`.
- **TDD** — test files precede implementation. Unit tests require no database; integration tests are gated behind `//go:build integration`.
- **Minimal** — one concern per file, no premature abstraction.

## Project Structure

```
internal/model/   — URL struct + Validate() + ErrInvalidURL
internal/store/   — URLStore interface, InMemory impl, Postgres impl
internal/handler/ — HTTP handlers (CreateURL, GetURL, ListURLs)
internal/server/  — ServeMux wiring
main.go           — config → store → server
```

## API

| Method | Path       | Success |
|--------|------------|---------|
| POST   | /urls      | 201     |
| GET    | /urls/{id} | 200     |
| GET    | /urls      | 200     |

Errors always return `{"error": "..."}` JSON.

## Config

| Var            | Required | Default |
|----------------|----------|---------|
| `DATABASE_URL` | Yes      | —       |
| `PORT`         | No       | `8080`  |

## Commands

```bash
go test ./...                                                        # unit tests (no DB)
go test -tags=integration ./internal/store/...                       # integration tests
go build ./...                                                       # build check
DATABASE_URL=postgres://user:pass@localhost/db?sslmode=disable go run .
```
