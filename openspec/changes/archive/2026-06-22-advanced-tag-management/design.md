## Context

Tags today are stored per entry as a `models.StringList` (a JSON-encoded string array in SQLite, scoped by `family_id`). The backend exposes only `GET /v1/tags` (distinct names, via `storage.GetDistinctTags`) and an implicit `GET /v1/items?tags=` OR filter (via `JSON_EXTRACT(tags,'$') LIKE '%"tag"%'`). The API layer is **oapi-codegen generated**: handlers live in `pkg/server/api/api_items_service.go` and implement interfaces in `pkg/generated/goserver/`, regenerated from `api/openapi.yaml` via `make generate`. The frontend is Next.js 15; tags surface in `EntryEditor` (comma text field + autocomplete), `EntryViewer`/`EntryCard` (badges), the search page (text only), and the profile page (unique count + unranked top-5, both computed client-side by loading all entries).

The pain points this change targets: removing a tag in the editor means deleting characters; there is no way to browse entries by tag from a list; counts are absent; and a typo'd or unwanted tag can never be fixed or purged across entries.

## Goals / Non-Goals

**Goals:**
- Editor tags become chips with per-tag X removal and an inline add input that keeps autocomplete.
- A server-computed `GET /v1/tags/stats` returns `[{name, count}]` so counts don't require loading all entries.
- A dedicated Tags page lists tags with counts (sorted desc), browses entries by tag, and hosts rename/delete.
- Cross-entry rename (merge-on-collision) and delete via `PATCH`/`DELETE /v1/tags/{name}`.
- Search page exposes the existing `tags` filter as chips (OR among tags, AND with text).

**Non-Goals:**
- No change to tag storage format or any data migration.
- No change to AI suggestion / pending-tag behavior (`ai-tagging` capability untouched).
- No per-entry tag editing from the Tags page (only rename/delete across all entries).
- No AND semantics for the multi-tag filter — OR only (matches existing backend).
- `GET /v1/tags` (`string[]`) is **not** modified; counts go in a separate endpoint.

## Decisions

### 1. New `GET /v1/tags/stats` instead of changing `GET /v1/tags`
`GET /v1/tags` returns `string[]` and is consumed by editor autocomplete and AI suggestion context. Adding counts there would change the response shape and ripple into those consumers. A separate `GET /v1/tags/stats` returning `[{name, count}]` keeps the existing contract intact. *Alternative considered:* a `?withCounts=true` query flag — rejected as it overloads one endpoint with two response shapes.

### 2. Count aggregation in Go, not SQL
Tags are a JSON array column, so counts can't come from a clean `GROUP BY`. `GetTagStats` will select `tags` for the family (as `GetDistinctTags` already does), tally occurrences in Go (one entry contributes at most 1 to each distinct tag it carries), and return names with counts sorted by count desc then name asc. *Alternative considered:* SQLite `json_each` table-valued function — more efficient but couples us to a SQLite-specific extension and complicates GORM usage; the dataset (one entry per date per family) is small enough that in-Go tallying is fine.

### 3. Rename/delete as load-mutate-save per affected entry
SQLite can't bulk-rewrite a JSON array in a single `UPDATE`. `RenameTag(familyID, old, new)` and `DeleteTag(familyID, name)` will load the family's entries whose tags contain the target, mutate the `StringList` in Go, and re-save each — all inside one DB transaction so a partial failure rolls back. Rename de-duplicates within each entry (merge-on-collision) and preserves tag order, skipping entries already carrying `new`. *Alternative considered:* raw SQL `REPLACE()` on the JSON text — rejected as fragile (would corrupt substrings, e.g. renaming `work` would hit `homework`, and can't de-dupe).

### 4. Reuse existing change-tracking / save path
Entries already maintain `TagsSourceHash` and change-tracking for sync. Rename/delete only mutate the confirmed `tags` array, not title/body, so the content hash is unchanged; the save must still bump the entry's change-tracking timestamp so sync clients pick up the edit. The implementation will route through the existing item-update mechanism rather than a raw column write, to keep sync correct.

### 5. Chip input reuses the existing suggestion-chip pattern
`EntryEditor` already renders AI suggestions as pill chips with X/+ buttons and already has `knownTags` + autocomplete-dropdown logic. The confirmed-tags field becomes a chip row plus an inline `<input>`; the autocomplete dropdown re-anchors from "last comma token" to "inline input value". The form still serializes to the same `tags: string[]` on save, so `useSaveEntry` and the API are unchanged.

### 6. `{name}` path param encoding
Tag names can contain spaces and unicode. The rename/delete routes take the tag in the path (`/v1/tags/{name}`); the frontend URL-encodes the name and the backend decodes it. Empty/whitespace names are rejected (400). A rename whose `newName` is blank or equals the old name is a 400 / no-op respectively.

## Risks / Trade-offs

- **Bulk rename/delete touches many entries** → wrap in a single transaction so it's all-or-nothing; the operation is family-scoped and the per-family entry count is bounded by one-per-date.
- **Concurrent edit during rename** (a user saves an entry while a rename runs) → the transaction provides isolation; last-writer-wins is acceptable for a single-family diary and matches existing upsert semantics.
- **Destructive delete** → UI gates delete behind a confirmation showing the affected entry count ("remove from N entries?"); no undo, consistent with the app's existing delete flows.
- **Autocomplete regression risk** in the rewritten chip field → covered by updating the existing `tag-autocomplete` E2E spec to drive the inline input instead of the comma field.
- **Count drift** between the `stats` endpoint and the client's previous all-entries tally on profile → profile switches to the server count, removing the divergence.

## Migration Plan

No data migration. Deploy is a standard `make generate` + backend build + frontend build. Rollback is a redeploy of the prior image; since storage format is unchanged, no data cleanup is needed on rollback. New endpoints are additive; old `GET /v1/tags` keeps working throughout.

## Open Questions

None — the three contested design points (separate `stats` endpoint, rename-merges-on-collision, OR-only multi-tag filter) were resolved with the user before this proposal.
