## 1. API contract (OpenAPI + codegen)

- [ ] 1.1 Add `GET /v1/tags/stats` to `api/openapi.yaml` returning a list of `TagStat` (`{ name: string, count: integer }`)
- [ ] 1.2 Add `PATCH /v1/tags/{name}` to `api/openapi.yaml` with body `{ newName: string }`, returning 200 on success and 400 for blank/unchanged name
- [ ] 1.3 Add `DELETE /v1/tags/{name}` to `api/openapi.yaml` returning 200/204 on success
- [ ] 1.4 Run `make generate` and confirm the new operations/schemas appear in `pkg/generated/goserver`

## 2. Storage layer (Go)

- [ ] 2.1 Add `GetTagStats(familyID) ([]TagStat, error)` to the storage interface and implement it: tally each distinct tag once per entry, sort by count desc then name asc
- [ ] 2.2 Add `RenameTag(familyID, oldName, newName) error`: load entries carrying `oldName`, rewrite each `StringList` (replace, de-dup, preserve order, skip if `newName` already present), re-save inside one transaction via the existing item-update path so change-tracking is bumped
- [ ] 2.3 Add `DeleteTag(familyID, name) error`: load entries carrying `name`, remove it from each `StringList`, re-save inside one transaction
- [ ] 2.4 Unit tests in `storage_tags_test.go`: stats counting/sorting/family-scoping; rename incl. merge-on-collision, non-existent tag no-op, family scoping; delete incl. last-tag-becomes-empty, non-existent no-op; transaction rollback on failure

## 3. API handlers (Go)

- [ ] 3.1 Implement `GetTagStats` handler in `pkg/server/api/api_items_service.go` (resolve `family_id`, call storage, map to `TagStat` response)
- [ ] 3.2 Implement `RenameTag` handler: URL-decode `{name}`, validate `newName` (non-blank, not equal to old) → 400 otherwise, call `RenameTag`
- [ ] 3.3 Implement `DeleteTag` handler: URL-decode `{name}`, call `DeleteTag`
- [ ] 3.4 Handler-level tests covering 400 validation and family scoping

## 4. Frontend API client & types

- [ ] 4.1 Add `TagStat` and any request/response types to `next-frontend/src/types/index.ts`
- [ ] 4.2 Add `getTagStats()`, `renameTag(name, newName)`, `deleteTag(name)` to `next-frontend/src/lib/api/diary.ts`

## 5. Editor chip input (entries + tag-autocomplete)

- [ ] 5.1 Replace the comma-separated tags `<Input>` in `EntryEditor.tsx` with a chip row (removable `[tag ×]` chips) plus an inline add-tag `<input>`; serialize chips to `tags: string[]` on save
- [ ] 5.2 Re-anchor the existing autocomplete dropdown to the inline input (match against `knownTags`, exclude already-selected chips, commit on select/Enter/comma, clear input after)
- [ ] 5.3 Preserve AI suggestion chips and accept/dismiss behavior unchanged alongside the new confirmed-tag chips
- [ ] 5.4 Update `e2e` tag-autocomplete spec to drive the inline input and assert chip add/remove

## 6. Tags page + browse (tag-management)

- [ ] 6.1 Create the Tags page route under `next-frontend/src/app/(dashboard)` listing tags with counts from `getTagStats()`, sorted by count desc, with an empty state
- [ ] 6.2 Add browse-by-tag: selecting a tag shows entries via the existing `diaryApi.search`/`?tags=` filter; entry cards navigate to `/diary/[date]`
- [ ] 6.3 Make the profile "Tags" stat card a link to the Tags page; switch its count to `getTagStats()` (or keep client count) consistently

## 7. Rename / delete UI (tag-management)

- [ ] 7.1 Add per-tag rename (✎) action on the Tags page: inline edit → `renameTag`, then refresh the list (reflecting merged counts)
- [ ] 7.2 Add per-tag delete (🗑) action behind a confirmation dialog that names the affected entry count ("remove from N entries?") → `deleteTag`, then refresh
- [ ] 7.3 Surface API errors via the existing toast/error pattern

## 8. Search tag filter (search)

- [ ] 8.1 Add a tag filter chip row to the search page (add/remove chips; reuse `knownTags` for suggestions)
- [ ] 8.2 Pass selected tags to `diaryApi.search` as comma-separated `tags`; run a tag-only search even when text is below the 2-char minimum
- [ ] 8.3 Ensure text + tags combine (AND with text, OR among tags) per spec

## 9. Verification

- [ ] 9.1 Run backend static checks/tests (`make lint` / `make test` or `go vet` + `go test ./...`) and fix issues
- [ ] 9.2 Run frontend lint/build/typecheck and fix issues
- [ ] 9.3 Deploy the branch to the diary-wip stack and run the Playwright E2E suite against it (chip editor, Tags page, browse, search filter, rename/delete); fix failures
- [ ] 9.4 Add/confirm E2E coverage for rename merge-on-collision and delete confirmation
