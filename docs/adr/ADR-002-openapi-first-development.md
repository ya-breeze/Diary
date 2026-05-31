# ADR-002: OpenAPI-First Development with oapi-codegen

## Status
Accepted

## Context and Problem Statement

The backend exposes a REST API consumed by the Next.js frontend and a mobile sync client. We need a reliable contract between these components and want to avoid manually maintaining both the spec and the implementation in sync.

## Decision Drivers

- The API must be consistent between server, web client, and mobile client
- Documentation should stay accurate without extra effort
- Go server stubs and a typed Go client for E2E tests should be derivable from one source
- Code generation should be fast and not require Docker or external services

## Considered Options

- **Spec-first with oapi-codegen** — write `api/openapi.yaml`, generate server stubs and client
- **Code-first (swaggo/swag)** — annotate Go handlers, generate spec from annotations
- **No codegen** — write handlers and spec independently, keep them in sync manually

## Decision Outcome

Chosen: **Spec-first with oapi-codegen** (`github.com/oapi-codegen/oapi-codegen`).

`api/openapi.yaml` is the single source of truth. `make generate` runs `oapi-codegen` to produce:
- `backend/pkg/generated/goserver/` — Gorilla Mux server stubs and strict handler interfaces
- `backend/pkg/generated/goclient/` — typed Go HTTP client used in integration tests

Handlers implement the generated strict interfaces, so adding or changing an endpoint requires updating the spec first. This surfaces breaking changes at compile time.

### Pros

- API contract is explicit and version-controlled
- Compile-time enforcement: missing handler methods fail the build
- Generated client makes integration tests concise and type-safe
- `oapi-codegen` runs as a `go tool` — no Docker, no external binary required

### Cons

- Extra `make generate` step before editing handler code
- Generated files must not be hand-edited, which can be unintuitive
- `oapi-codegen`'s strict-server pattern requires adapting return types to the generated wrapper
