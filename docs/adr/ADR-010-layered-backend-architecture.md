# ADR-010: Layered Backend Architecture (API → Storage Interface)

## Status
Accepted

## Context and Problem Statement

The Go backend must be organized in a way that keeps HTTP handling, business logic, and data access separable and testable. The choice of layering affects how easy it is to mock the database in tests and how clearly responsibilities are divided.

## Decision Drivers

- Database access must be mockable in unit tests without starting a real SQLite instance
- HTTP handlers should not embed SQL queries directly
- The generated OpenAPI server stubs define handler interfaces — the implementation layer must fit that contract
- The codebase should be navigable: a developer should be able to find where a given operation lives

## Considered Options

- **Flat handlers** — all logic in a single handler function; simple for small apps but hard to test
- **MVC (Model-View-Controller)** — classic pattern, but "View" is not meaningful for a JSON API
- **Two layers: API handlers + `database.Storage` interface** — handlers call storage directly via an interface

## Decision Outcome

Chosen: **Two explicit layers — API service handlers and a `database.Storage` interface** — with the generated OpenAPI strict-server as the entry point.

```
HTTP request
    │
    ▼
goserver (generated strict-server)
    │  calls
    ▼
server/api/*APIService   ← business logic lives here
    │  calls
    ▼
database.Storage         ← interface; real impl is database.storage (SQLite/GORM)
    │
    ▼
SQLite via GORM
```

Each API domain has its own service file (`api_items_service.go`, `api_sync_service.go`, etc.). Tests inject a mock `Storage` implementation. Background tasks (`CheckerTask`, `BackupTask`) also receive `database.Storage`, keeping them testable independently of HTTP.

### Pros

- `database.Storage` interface enables pure-Go unit tests with no SQLite process
- Business logic is in one place per domain — easy to locate
- Generated strict-server enforces the OpenAPI contract at compile time
- New endpoints require changes in one service file, not spread across a framework scaffold

### Cons

- For simple CRUD endpoints the service layer is thin (a direct pass-through to storage)
- The `database.Storage` interface must be kept in sync with the implementation manually
- No explicit "service" abstraction between handlers and storage — complex business logic lives in the handler service files
