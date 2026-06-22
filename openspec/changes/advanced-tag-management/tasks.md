## 1. API contract (OpenAPI + codegen)

- [x] 1.1 Add `GET /v1/tags/stats` to `api/openapi.yaml` returning a list of `TagStat` (`{ name: string, count: integer }`)
- [x] 1.2 Add `PATCH /v1/tags/{name}` to `api/openapi.yaml` with body `{ newName: string }`, returning 200 on success and 400 for blank/unchanged name
- [x] 1.3 Add `DELETE /v1/tags/{name}` to `api/openapi.yaml` returning 200/204 on success
- [x] 1.4 Run `make generate` and confirm the new operations/schemas appear in `pkg/generated/goserver` (also wired `interfaces.go` + `adapter.go`, which are hand-maintained)

## 2. Storage layer (Go)

- [x] 2.1 Add `GetTagStats(familyID) ([]TagStat, error)` to the storage interface and implement it: tally each distinct tag once per entry, sort by count desc then name asc
- [x] 2.2 Add `RenameTag(familyID, oldName, newName) error`: load entries carrying `oldName`, rewrite each `StringList` (replace, de-dup, preserve order, skip if `newName` already present), re-save inside one transaction via the existing item-update path so change-tracking is bumped
- [x] 2.3 Add `DeleteTag(familyID, name) error`: load entries carrying `name`, remove it from each `StringList`, re-save inside one transaction
- [x] 2.4 Unit tests in `storage_tags_test.go`: stats counting/sorting/family-scoping; rename incl. merge-on-collision, non-existent tag no-op, family scoping; delete incl. last-tag-becomes-empty, non-existent no-op (rollback is structurally guaranteed by the shared transaction helper `mutateFamilyTags`)
- [x] 2.5 Make the tag filter robust to legacy non-JSON `tags` columns (`tags LIKE ?` instead of `JSON_EXTRACT`, which 500s on malformed rows) + regression test; add a startup `normalizeTagColumns` migration that rewrites NULL/empty/non-JSON tag columns to `[]` + test

## 3. API handlers (Go)

- [x] 3.1 Implement `GetTagStats` handler in `pkg/server/api/api_items_service.go` (resolve `family_id`, call storage, map to `TagStat` response)
- [x] 3.2 Implement `RenameTag` handler: URL-decode `{name}`, validate `newName` (non-blank, not equal to old) → 400 otherwise, call `RenameTag`
- [x] 3.3 Implement `DeleteTag` handler: URL-decode `{name}`, call `DeleteTag`
- [x] 3.4 Handler-level tests covering 400 validation and family scoping

## 4. Frontend API client & types

- [x] 4.1 Add `TagStat` and any request/response types to `next-frontend/src/types/index.ts`
- [x] 4.2 Add `getTagStats()`, `renameTag(name, newName)`, `deleteTag(name)` to `next-frontend/src/lib/api/diary.ts`

## 5. Editor chip input (entries + tag-autocomplete)

- [x] 5.1 Replace the comma-separated tags `<Input>` in `EntryEditor.tsx` with a chip row (removable `[tag ×]` chips) plus an inline add-tag `<input>`; serialize chips to `tags: string[]` on save
- [x] 5.2 Re-anchor the existing autocomplete dropdown to the inline input (match against `knownTags`, exclude already-selected chips, commit on select/Enter/comma, clear input after)
- [x] 5.3 Preserve AI suggestion chips and accept/dismiss behavior unchanged alongside the new confirmed-tag chips
- [x] 5.4 Update `e2e` tag-autocomplete spec to drive the inline input and assert chip add/remove

## 6. Tags page + browse (tag-management)

- [x] 6.1 Create the Tags page route under `next-frontend/src/app/(dashboard)` listing tags with counts from `getTagStats()`, sorted by count desc, with an empty state
- [x] 6.2 Add browse-by-tag: selecting a tag shows entries via the existing `diaryApi.search`/`?tags=` filter; entry cards navigate to `/diary/[date]`
- [x] 6.3 Make the profile "Tags" stat card a link to the Tags page; switched its count to `getTagStats()` (server count)
- [x] 6.4 Make browse deep-linkable via `/tags?tag=<name>` (URL-driven, bookmarkable) and make the profile "Top tags" badges clickable → browse

## 7. Rename / delete UI (tag-management)

- [x] 7.1 Add per-tag rename (✎) action on the Tags page: inline edit → `renameTag`, then refresh the list (reflecting merged counts)
- [x] 7.2 Add per-tag delete (🗑) action behind a confirmation dialog that names the affected entry count ("remove from N entries?") → `deleteTag`, then refresh
- [x] 7.3 Surface API errors via an inline error message (Diary has no toast system; matches the editor/profile pattern)

## 8. Search tag filter (search)

- [x] 8.1 Add a tag filter chip row to the search page (add/remove chips; reuse `knownTags` for suggestions)
- [x] 8.2 Pass selected tags to `diaryApi.search` as comma-separated `tags`; run a tag-only search even when text is below the 2-char minimum
- [x] 8.3 Ensure text + tags combine (AND with text, OR among tags) per spec

## 9. Verification

- [x] 9.1 Run backend static checks/tests (`make lint` / `make test` or `go vet` + `go test ./...`) and fix issues
- [x] 9.2 Run frontend lint/build/typecheck and fix issues
- [x] 9.3 Deployed the branch to diary-wip and ran the Playwright E2E suite against it: 22 passed, 1 pre-existing flaky (ai-tagging backfill toggle, passes on retry, unrelated to this change)
- [x] 9.4 Added `tag-management.spec.ts` covering profile→Tags link, browse-by-tag, rename merge-on-collision (count merges to 3), and delete confirmation ("from N entries")
