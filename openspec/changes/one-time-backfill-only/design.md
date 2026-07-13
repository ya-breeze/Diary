## Context

Two code paths trigger automatic AI tag suggestion today:

1. **`maybeRetag`** (`pkg/server/api/api_items_service.go`) — called from the item-save handler when `contentChanged`. It fires on **create** (no existing item ⇒ `contentChanged = true`) and on **edit**, spawning `runRetag` in a detached goroutine (with `retagInflight`/`retagMu` coalescing) to stage/auto-apply suggestions.
2. **`UntaggedCheck`** (`pkg/checker/check_untagged.go`) — the 24h health sweep. Gated by `AITaggingEnabled && AITaggingBackfill`. A day is a candidate when `stale` (`TagsSourceHash != ComputeTagsSourceHash(title,body)`). Because `PutItem` re-stamps the hash on every save, normally-saved edits are **not** stale — so "stale" candidates are essentially pre-existing (empty-hash) items. But `processItem` returns early on `len(names)==0` **without** stamping the hash, so no-suggestion items stay empty-hash and are re-analyzed every run.

## Goals / Non-Goals

**Goals:**
- New and edited entries never trigger automatic AI (no `maybeRetag`).
- The manual "Suggest tags" endpoint is the only interactive AI path (unchanged).
- Backfill is a true one-shot per family: each pre-existing entry analyzed at most once, then the family is marked done and the sweep stops.
- Re-runnable by toggling the Backfill setting off→on.

**Non-Goals:**
- Changing the manual suggestion endpoint or its behavior.
- Changing request/response shapes.
- Re-analyzing edited content ever (that is the whole point).
- A UI redesign — only optional Backfill-toggle copy.

## Decisions

### Decision: Remove `maybeRetag` outright rather than gating it
Delete `maybeRetag`, `runRetag`, the `retagInflight`/`retagMu` fields, and the `contentChanged` computation in the save handler. Saving becomes a pure persist with no AI side effect.
- **Why:** The user wants *no* automatic analysis on save; a gate/flag would leave dead complexity. The manual button already covers on-demand tagging.
- **Alternative considered:** keep `maybeRetag` behind a family flag. Rejected — extra state and surface for a path we want gone.

### Decision: `ai_tagging_backfill_done` boolean on the family, default false
New column (GORM migration). Gate the sweep with `AITaggingEnabled && AITaggingBackfill && !AITaggingBackfillDone`.
- **Migration:** existing families default to `false`, so on first deploy the backfill runs to completion once (processing the leftover empty-hash / no-suggestion items), then flips true. This is safe: those items are exactly the ones that were being re-analyzed every run anyway.

### Decision: Stamp the hash even when a day yields no suggestions
In `processItem`, when the model returns no suggestions, stamp `TagsSourceHash` (mark analyzed) instead of returning early unstamped.
- **Why:** required for the one-shot to terminate — otherwise no-suggestion days remain candidates forever and `backfill_done` never flips.
- **Note:** stamping uses the same `ComputeTagsSourceHash(title,body)` so a later genuine content change would mismatch — but since the sweep is gated by `!backfill_done`, a completed family never re-analyzes regardless. The hash stamp is the per-item "analyzed" marker; `backfill_done` is the per-family "sweep exhausted" marker. Both are needed: the hash lets a single sweep skip done items and know when the corpus is empty; the flag lets the family stop sweeping entirely.

### Decision: Flip `backfill_done` only when the corpus is exhausted, not at the per-run cap
A run processes up to `maxBackfillPerRun` fresh analyses. Set `backfill_done = true` only when a full family scan completes with **no** remaining un-analyzed candidate (i.e., it did not stop because it hit the cap). If it hit the cap, more work remains for the next run.
- **Implementation shape:** track whether the scan stopped early (cap) vs. finished with zero fresh analyses needed. If it finished and every candidate is now analyzed, mark done.

### Decision: Reset `backfill_done` on Backfill off→on transition
In `SetFamilyAISettings`, when `ai_tagging_backfill` transitions from false to true, set `backfill_done = false`.
- **Why:** gives the user an explicit escape hatch to re-run over any still-un-analyzed entries (e.g., after a bulk import), without a separate control.

## Risks / Trade-offs

- **A family mid-backfill on deploy** → it simply continues from where the hashes left off and flips done when exhausted. No data migration of items needed.
- **`ai_tagging_auto` semantics narrow** (backfill-only) → documented in the spec; the edit-time auto-apply path is gone with `maybeRetag`. Existing behavior tests for edit-retag must be removed/replaced.
- **Detecting "corpus exhausted" vs "hit cap"** must be precise, or the flag never flips (stuck sweeping) or flips too early (leaves items un-analyzed). Covered by tests.

## Migration Plan

Additive column + code change; forward-only. On deploy, each enabled+backfill family finishes its remaining backfill once and flips `backfill_done`. Rollback = revert (the unused column is harmless).

## Open Questions

- Profile "Backfill" toggle copy: reword to convey "one-time" and "toggle off→on to re-run"? (Minor; can finalize during implementation.)
