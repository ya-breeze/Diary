## 1. Remove on-save auto-retag

- [x] 1.1 `api_items_service.go`: remove the `maybeRetag` call from the save handler and the `contentChanged` computation that guarded it
- [x] 1.2 Delete `maybeRetag`, `runRetag`, and the `retagInflight` map + `retagMu` mutex fields; remove now-unused imports/helpers (`enabledFamily` only if unused elsewhere)
- [x] 1.3 Confirm the save handler still persists correctly and returns the same response shape (no AI side effect)

## 2. Family model: backfill-done flag

- [x] 2.1 Add `AITaggingBackfillDone bool` (column `ai_tagging_backfill_done`, default false) to the family model; add the GORM migration/auto-migrate
- [x] 2.2 Reset-on-toggle: `UpdateFamilySettings` already loads `current` before writing, so detect `!current.AITaggingBackfill && newBackfill` there (or read-current inside `SetFamilyAISettings`) and set `AITaggingBackfillDone = false` on that transition; adjust the `SetFamilyAISettings` signature/storage write accordingly
- [x] 2.3 Update the `AITaggingAuto` model doc comment (family.go) — it currently says "unattended triggers (on-save, backfill)"; on-save is gone, so it's backfill-only

## 3. Backfill becomes a one-shot

- [x] 3.1 `check_untagged.go`: keep the family gate `AITaggingEnabled && AITaggingBackfill`; gate only the fresh-analysis branch (`processItem`/model call) on `!AITaggingBackfillDone`. The "surface existing pending" branch must still run so already-staged pending days keep appearing as review issues
- [x] 3.2 `processItem`: when the model yields no suggestions, stamp `TagsSourceHash` (mark analyzed) instead of returning early unstamped
- [x] 3.3 In `runForFamily`, distinguish "scan finished with no remaining un-analyzed candidate" from "stopped at the per-run cap"; when the corpus is exhausted, set `AITaggingBackfillDone = true` and persist it (skip this flip entirely when already done)
- [x] 3.4 Verify: a family with `AITaggingBackfillDone = true` makes zero model calls, yet days with existing `PendingTags` are still surfaced as non-fixable review issues; never-analyzed days are left alone

## 4. Tests

- [x] 4.1 `pkg/server/api`: saving a new entry triggers no AI suggestion/staging; editing an entry triggers no re-analysis (was previously handled by `maybeRetag`)
- [x] 4.2 `pkg/checker`: a pre-existing day yielding no suggestions is marked analyzed and not re-queried on the next run
- [x] 4.3 `pkg/checker`: backfill flips `AITaggingBackfillDone = true` when the corpus is exhausted, and does not flip when it stopped at the per-run cap
- [x] 4.4 `pkg/checker`: a family with `AITaggingBackfillDone = true` makes no model calls, AND a day that already has `PendingTags` is still surfaced as a review issue (results not hidden on completion)
- [x] 4.5 `pkg/checker` or family store test: toggling `ai_tagging_backfill` off→on resets `AITaggingBackfillDone = false`
- [x] 4.6 Update/remove existing tests that assert edit/create-triggered retag behavior

## 5. UI copy (optional)

- [x] 5.1 Reword the profile "Backfill" toggle to convey one-time semantics and "toggle off→on to re-run" (only if it currently implies ongoing behavior)

## 6. ADR

- [x] 6.1 Add an ADR (Status: Proposed) recording the shift to one-time-backfill-only AI tagging (removal of on-save retag, per-family completion flag); flip to Accepted at merge

## 7. Verification

- [x] 7.1 Run `make lint` and backend tests (`go test ./...`) and fix issues
- [ ] 7.2 Deploy to diary-wip; confirm: creating/editing an entry stages no suggestions; the manual "Suggest tags" button still works; a family's backfill completes and stops (check logs / DB flag)
