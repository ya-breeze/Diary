# ADR-012: Automatic AI Tagging Is a One-Time Backfill Only

## Status

Proposed

(Introduced by OpenSpec change `one-time-backfill-only`. Amends the trigger model of ADR-011 — the suggest-vs-write boundary and provider decisions there stand.)

## Context and Problem Statement

ADR-011 introduced AI tag suggestion with two unattended triggers: an on-save retag (`maybeRetag`, fired on both creating and editing an entry) and a recurring background backfill over untagged/stale days. In practice the on-save trigger proved unwanted: new entries were auto-analyzed even though the editor already offers an explicit "Suggest tags" button, and editing an entry re-staged suggestions the user had deliberately dismissed. The recurring backfill also never truly finished — days for which the model returned no suggestions were never marked analyzed, so they were re-queried on every 24h sweep, indefinitely.

The owner's intent: automatic AI should exist only to bootstrap tags on the pre-existing corpus, exactly once. Everything after that is user-initiated.

## Decision Drivers

- The user must stay in control: automatic analysis must never react to the user writing or editing entries
- The explicit "Suggest tags" action must remain the only interactive AI path
- The backfill must provably terminate (bounded total model calls per family)
- A future bulk import needs an escape hatch to re-run the backfill without new mechanisms
- Keep cost predictable: after completion, background AI cost for a family is zero

## Considered Options

- **On-save trigger**: keep behind a new per-family flag vs. remove entirely. A flag preserves optionality but leaves dead complexity (goroutine coalescing, inflight maps) for a path the owner wants gone.
- **Backfill lifecycle**: recurring-but-idempotent (stamp every analyzed day, scan forever, skip everything) vs. a true one-shot with a per-family completion flag. The recurring scan still iterates all items every run and its "doneness" is implicit.
- **Completion granularity**: per-item only (hash) vs. per-item + per-family flag. Hash alone cannot express "stop sweeping this family."

## Decision

1. **Remove the on-save trigger entirely.** `maybeRetag`/`runRetag` and their coalescing machinery are deleted; saving an entry (create or edit) has no AI side effect. The manual "Suggest tags" endpoint is unchanged.
2. **The backfill is a one-shot per family**, tracked by a new `ai_tagging_backfill_done` column (default false). It is gated by `ai_tagging_enabled && ai_tagging_backfill && !ai_tagging_backfill_done` and processes pre-existing (never-analyzed) days in bounded batches until none remain.
3. **Every analyzed day is marked** by stamping `tags_source_hash` — including days that yield no suggestions (previously the non-terminating case).
4. **Completion flips the flag** only when a sweep finishes with no remaining un-analyzed day and no failed analysis (stopping at the per-run cap or a model error leaves work for the next run).
5. **Completion stops new analysis, not review**: days with staged `pending_tags` keep appearing in the untagged-days review list until the user accepts or dismisses them.
6. **Re-trigger** by toggling `ai_tagging_backfill` off→on, which resets `ai_tagging_backfill_done`. Only days without a stamped hash are picked up (entries saved through the normal API are stamped on save; re-analyzing them is deliberately not supported).
7. `ai_tagging_auto` (confident auto-apply) now governs the backfill only.

## Consequences

- New and edited entries are never auto-tagged; users tag interactively via the editor button.
- Per-family background AI cost is bounded: at most one model call per pre-existing day, ever (until an explicit re-trigger).
- The `untagged` health check becomes a bootstrap mechanism rather than an ongoing reconciler; content edits no longer resurface days.
- `tags_source_hash` semantics shift from "staleness detector" to "analyzed marker" — a mismatch after an edit no longer causes re-analysis because the family gate (`backfill_done`) short-circuits first.
- Bulk imports that go through the normal save path are stamped on save and will not be picked up by a re-triggered backfill; importing with re-analysis in mind requires inserting without a hash (accepted limitation).
