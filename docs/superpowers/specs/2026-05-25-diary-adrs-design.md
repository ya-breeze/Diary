# Design: Architecture Decision Records for Diary

## Status
Approved

## Overview

Create 10 ADRs documenting the key architectural decisions already made and reflected in the Diary codebase.

## Format

MADR (lightweight Markdown Architectural Decision Records):
- Status
- Context and Problem Statement
- Decision Drivers
- Considered Options
- Decision Outcome (with Pros/Cons)

## Location

`docs/adr/ADR-NNN-<slug>.md` — numbered, kebab-case filenames.

## ADRs to Create

| File | Decision |
|------|----------|
| ADR-001-sqlite-database.md | Use SQLite as the database |
| ADR-002-openapi-first-development.md | OpenAPI-first development with oapi-codegen |
| ADR-003-jwt-http-only-cookies-silent-refresh.md | JWT auth with HTTP-only cookies and silent refresh |
| ADR-004-kin-core-family-multitenancy.md | Family-based multi-tenancy via kin-core |
| ADR-005-append-only-change-log-for-sync.md | Append-only change log for mobile sync |
| ADR-006-filesystem-asset-storage.md | Filesystem storage for assets |
| ADR-007-background-task-architecture.md | Background task architecture (health checker + backup) |
| ADR-008-rate-limiting-middleware.md | Rate limiting in middleware |
| ADR-009-date-keyed-diary-entries.md | Date-keyed diary entries |
| ADR-010-layered-backend-architecture.md | Layered backend architecture (API → Service → Data) |

## Evidence Sources

- `backend/pkg/database/sqlite.go` — SQLite/GORM setup
- `api/openapi.yaml` + `api/oapi-codegen-*.yaml` — OpenAPI-first with codegen
- `backend/pkg/server/middlewares.go` — JWT auth + rate limit middleware
- `backend/pkg/server/server.go` — kin-core integration, background tasks
- `backend/pkg/database/models/` — ItemChange (sync), Item (date-keyed), filesystem assets
- `backend/pkg/server/api/` — layered service handlers
