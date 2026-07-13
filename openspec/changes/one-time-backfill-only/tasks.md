## 1. Remove on-save auto-retag

- [ ] 1.1 `api_items_service.go`: remove the `maybeRetag` call from the save handler and the `contentChanged` computation that guarded it
- [ ] 1.2 Delete `maybeRetag`, `runRetag`, and the `retagInflight` map + `retagMu` mutex fields; remove now-unused imports/helpers (`enabledFamily` only if unused elsewhere)
- [ ] 1.3 Confirm the save handler still persists correctly and returns the same response shape (no AI side effect)

## 2. Family model: backfill-done flag

- [ ] 2.1 Add `AITaggingBackfillDone bool` (column `ai_tagging_backfill_done`, default false) to the family model; add the GORM migration/auto-migrate
- [ ] 2.2 `SetFamilyAISettings`: when `ai_tagging_backfill` transitions false→on, set `AITaggingBackfillDone = false`; expose the field on read where family AI settings are returned (no new API field required unless the UI needs it)

## 3. Backfill becomes a one-shot

- [ ] 3.1 `check_untagged.go`: extend the family gate to `AITaggingEnabled && AITaggingBackfill && !AITaggingBackfillDone`
- [ ] 3.2 `processItem`: when the model yields no suggestions, stamp `TagsSourceHash` (mark analyzed) instead of returning early unstamped
- [ ] 3.3 In `runForFamily`, distinguish "scan finished with no remaining un-analyzed candidate" from "stopped at the per-run cap"; when the corpus is exhausted, set `AITaggingBackfillDone = true` and persist it
- [ ] 3.4 Ensure a family with `AITaggingBackfillDone = true` runs no AI analysis (still fine to surface already-staged pending as review issues, without model calls)

## 4. Tests

- [ ] 4.1 `pkg/server/api`: saving a new entry triggers no AI suggestion/staging; editing an entry triggers no re-analysis (was previously handled by `maybeRetag`)
- [ ] 4.2 `pkg/checker`: a pre-existing day yielding no suggestions is marked analyzed and not re-queried on the next run
- [ ] 4.3 `pkg/checker`: backfill flips `AITaggingBackfillDone = true` when the corpus is exhausted, and does not flip when it stopped at the per-run cap
- [ ] 4.4 `pkg/checker`: a family with `AITaggingBackfillDone = true` makes no model calls
- [ ] 4.5 `pkg/checker` or family store test: toggling `ai_tagging_backfill` off→on resets `AITaggingBackfillDone = false`
- [ ] 4.6 Update/remove existing tests that assert edit/create-triggered retag behavior

## 5. UI copy (optional)

- [ ] 5.1 Reword the profile "Backfill" toggle to convey one-time semantics and "toggle off→on to re-run" (only if it currently implies ongoing behavior)

## 6. ADR

- [ ] 6.1 Add an ADR (Status: Proposed) recording the shift to one-time-backfill-only AI tagging (removal of on-save retag, per-family completion flag); flip to Accepted at merge

## 7. Verification

- [ ] 7.1 Run `make lint` and backend tests (`go test ./...`) and fix issues
- [ ] 7.2 Deploy to diary-wip; confirm: creating/editing an entry stages no suggestions; the manual "Suggest tags" button still works; a family's backfill completes and stops (check logs / DB flag)
