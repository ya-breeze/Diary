## Approach

The fix operates at two layers — runtime output and stored data — so that both
new reads and the existing dirty rows are covered.

### 1. API output filter (`api_items_service.go`)

`newItemResponse` now calls `filterTags` on `item.Tags` and `item.PendingTags`
before constructing the response struct. `filterTags` (already used on the write
path) trims whitespace and drops empty strings, returning a non-nil slice.

This is the primary defence: even if blank tags survive in the DB, they will
never reach the frontend.

### 2. Startup scrub (`migration.go` / `storage.go`)

`scrubBlankTags` loads all items (selecting `id`, `tags`, `pending_tags` only)
and rewrites any row whose tag lists contain empty/whitespace entries, using
`filterStringList` which mirrors the `filterTags` logic. It logs how many rows
it fixed and returns a non-fatal error if the DB query fails.

Wired into `Open()` right after `autoMigrateModels`.

### No delta spec

The items API contract is unchanged — blank tags were never part of the
specified output. This fix enforces existing intent; no spec update needed.
