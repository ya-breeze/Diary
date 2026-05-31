# ADR-004: Family-Based Multi-Tenancy via kin-core

## Status
Accepted

## Context and Problem Statement

The original Diary had a flat single-user model (`users` table with `id`, `login`, `hashed_password`). The requirement evolved to support _families_: groups of users whose diary data is isolated from other families but shared within the group. Auth and tenancy logic needed to be consistent with other apps in the same ecosystem.

## Decision Drivers

- Multiple users should be able to share diary entries under a single "family" tenant
- Auth and user management code should be reusable across sibling applications
- Migration from the old single-user schema must be non-destructive

## Considered Options

- **Custom family/auth implementation** — full control, no external dependency
- **Integrate `kin-core` shared library** — reuse proven auth + tenancy across apps in the ecosystem
- **Third-party auth service (Auth0, Keycloak)** — offload auth entirely, but adds infrastructure dependency

## Decision Outcome

Chosen: **`github.com/ya-breeze/kin-core`** — a shared internal library providing family-based multi-tenancy, password hashing, JWT issuance, and refresh-token management.

All diary data models carry a `FamilyID uuid.UUID` field. Middleware extracts the authenticated family from the JWT and injects it into the request context. Handlers retrieve it via `c.MustGet("family_id").(uuid.UUID)`.

A one-time migration converts old `user_id`-keyed rows to `family_id`-keyed rows by mapping each legacy user to a newly created family.

### Pros

- Auth and tenancy logic lives in one place and is tested once
- Family isolation is enforced at the data model level (`FamilyID` on every table)
- Sibling apps (KinCart, etc.) share the same auth surface

### Cons

- External dependency on an internal library — changes to `kin-core` can break Diary
- Required a complex data migration with legacy struct mappings
- Base64-encoded legacy password hashes needed special handling during migration
