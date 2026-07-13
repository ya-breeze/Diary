## Why

AI tag suggestion currently fires automatically in ways the user does not want:

1. **On save** — when an entry's tag-relevant content changes, the item service calls `maybeRetag` (async `runRetag`). This runs on **creating a new entry** (no existing item ⇒ content "changed") and on **editing** one, auto-staging (or auto-applying) suggestions. The user wants new and edited entries to be tagged only when they explicitly ask.
2. **Backfill never truly ends** — the background `UntaggedCheck` re-analyzes pre-existing items, but an item that yields no suggestions never gets its `TagsSourceHash` stamped, so it is re-analyzed on every 24h run indefinitely.

The desired model: automatic AI is a **one-time backfill of pre-existing entries only**; new and edited entries are never auto-analyzed; the manual "Suggest tags" button remains the only interactive AI path.

## What Changes

- **Remove `maybeRetag` entirely** — both the create and edit triggers — along with the now-unused `contentChanged` computation on save, `runRetag`, and the `retagInflight`/`retagMu` coalescing machinery. New/edited entries no longer trigger any automatic AI.
- **Keep the manual "Suggest tags" action unchanged** (`POST /v1/items/suggest-tags`) — it stays the only interactive AI path.
- **Make the background backfill a true one-shot per family**, gated by `ai_tagging_enabled && ai_tagging_backfill && !ai_tagging_backfill_done`:
  - It processes up to the per-run cap of pre-existing (empty/mismatched-hash) items per run, across successive runs, until the corpus is exhausted.
  - Stamp `TagsSourceHash` even when an item yields **no** suggestions, so the one-shot can terminate.
  - When a scan finds no remaining items needing a fresh AI call (corpus exhausted, not merely the per-run cap), set `ai_tagging_backfill_done = true` and stop running AI for that family.
  - **Re-trigger:** toggling `ai_tagging_backfill` OFF then ON resets `ai_tagging_backfill_done = false`, starting a fresh one-time pass.
- **`ai_tagging_auto`** now only auto-applies confident tags **during the backfill**; its save/edit role disappears with `maybeRetag`.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `ai-tagging`: retire the save-and-leave (on-save/edit) unattended trigger; the only unattended trigger becomes the one-time backfill. Add the one-shot backfill lifecycle with a per-family completion flag and re-trigger semantics; clarify `ai_tagging_auto` applies to backfill only.
- `health`: the untagged-days check no longer produces new untagged issues from edits; it still surfaces pending suggestions staged by the (one-time) backfill.

## Impact

- **Data model:** new `ai_tagging_backfill_done` boolean on the family (default `false`; DB migration).
- **Backend code:** `pkg/server/api/api_items_service.go` (remove `maybeRetag`/`runRetag`/`contentChanged`/retag mutex), `pkg/checker/check_untagged.go` (backfill-done gate, stamp-on-empty, flip-when-exhausted), family model + `SetFamilyAISettings` (new field + reset-on-toggle), `pkg/database/storage.go`.
- **Tests:** `pkg/checker` (one-shot termination, stamp-on-empty, flip, re-trigger) and `pkg/server/api` (save/create no longer retags).
- **UI:** the profile "Backfill" toggle wording may need a tweak to reflect one-time semantics.
- **Behavior:** no request/response shape change; auto-tagging on create/edit stops; the manual button is unaffected.
- **ADR:** warrants an ADR (Status: Proposed) — behavioral + data-model shift in the AI tagging model.
