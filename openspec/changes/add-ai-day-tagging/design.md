## Context

Diary stores one entry per `(family, date)` with `Title`, `Body` (markdown that embeds asset references via `![alt](filename)`), and `Tags` (a string list). Assets live on disk under `<data_path>/assets/<familyID>/`; `utils.GetAssetsFromMarkdown(body)` already extracts the filenames a given entry references. A background health subsystem (`pkg/checker`) runs `mime`, `orphans`, and `refs` checks on a 24h schedule and exposes results via `GET /v1/health/issues` + `POST /v1/health/fix`; each `Issue` carries an optional `fix func() error`.

The sibling project KinCart already integrates Gemini in `internal/ai/gemini.go`: `google.golang.org/genai`, model `gemini-2.0-flash`, `GEMINI_API_KEY`, strict `ResponseSchema` structured output, multimodal via inline `Blob` parts, and a "known items" list injected into the prompt to keep the model on the user's existing vocabulary. This change ports that proven pattern. Diary has no AI dependency today.

The architectural decisions below are recorded as **ADR-011** (`docs/adr/ADR-011-ai-tag-suggestion-gemini.md`); this section is the working rationale, the ADR is the durable record.

## Goals / Non-Goals

**Goals:**
- Suggest tags for a day from its text (phase 1) and optionally its media (phases 2–3) using the family's existing vocabulary.
- Never silently overwrite confirmed tags by default; keep the user in control via a `pending_tags` staging field and per-tag accept.
- Make every trigger (explicit button, in-editor debounce, on-save, background backfill) converge on the same suggestion engine and the same `pending_tags`/confidence outcome.
- Reuse the existing health-issues flow for backfill rather than building a new review subsystem.
- Degrade gracefully: with no `GEMINI_API_KEY`, the app behaves exactly as today.

**Non-Goals:**
- Native video upload to Gemini (Files API). Out of scope — keyframe sampling only.
- Audio/speech transcription of videos.
- A managed tag taxonomy / tag-rename tooling (vocabulary stays free-form strings).
- Cross-family or global tag learning. All vocabulary context is per-family.
- Re-tagging on a fixed clock independent of content change.

## Decisions

### 1. `pending_tags` as a separate field, not a flag on `Tags`
Suggestions are stored in a new `pending_tags` string list, distinct from the confirmed `Tags`. Accept = move a name from `pending_tags` into `Tags`. **Why:** keeps "what the AI thinks" and "what the user endorsed" cleanly separable, so suggestions can be regenerated/discarded without ever corrupting confirmed tags. Alternative (a `source: ai|user` marker per tag) was rejected as more invasive to the existing tag storage and search.

### 2. `tags_source_hash` drives all staleness decisions
Store `tags_source_hash = hash(Title + "\n" + Body + "\n" + join(sorted(GetAssetsFromMarkdown(Body)), ","))` on the entry. A day is "stale" when its stored hash differs from a freshly computed one (or is empty). **Why:** one deterministic signal serves the on-save trigger *and* the backfill health check, and naturally ignores no-op saves. Asset list is sorted so reordering markdown doesn't trigger a needless retag. Alternative (compare `updated_at` vs a `tagged_at` timestamp) was rejected because saves that don't touch title/body/media would still look stale.

### 3. Confidence routing concentrated in one place
The suggester always returns `{tags: [{name, confidence 0..1}]}`. Where results go depends only on the trigger context and `ai_tagging_auto`:

```
attended (button, in-editor debounce):  always → pending_tags (suggest)
unattended (on-save-and-leave, backfill):
    ai_tagging_auto = false → all → pending_tags  (health issue: "N days suggested")
    ai_tagging_auto = true  → day has confirmed tags? → pending_tags (suggest only)
                              day untagged + conf ≥ τ → Tags (auto-apply)
                              day untagged + conf < τ → pending_tags (health issue: "solve manually")
```

τ is a single configurable threshold (default chosen during phase 4, e.g. 0.8). **Why:** keeps decision #2 ("suggest, not apply") the default while letting power users opt into automation, with the in-editor experience invariant.

**Auto-apply is scoped to untagged days only.** Once a day has any confirmed tag, unattended triggers only stage suggestions — they never auto-apply. This is the v1 guard against the "AI re-adds a tag the user removed" fight: a user who deletes an auto-applied tag leaves the day curated (still has ≥1 tag, or is intentionally emptied via the editor which the user owns), so subsequent retags won't silently re-add it. AI is additive-only and never deletes a confirmed tag; pending and confirmed lists are kept disjoint (a name confirmed by the user is pruned from pending). A per-entry "dismissed tags" memory is deferred as a later enhancement if this proves insufficient.

### 4. Backfill is a 4th `checker.Check`, not a new subsystem
Implement an `UntaggedCheck` satisfying the existing `Check` interface. It emits an `Issue` per stale/untagged day. Under `ai_tagging_auto`, confident days get a populated `fix func()` (applies tags) and are reported `Fixable: true`; uncertain days are reported `Fixable: false` with a "solve manually" message and stored `pending_tags`. **Why:** reuses the 24h runner, `GET /v1/health/issues`, `POST /v1/health/fix`, and the family-scoping already in `Runner.RunForFamily`. The check is a no-op unless `ai_tagging_backfill` is enabled for the family.

### 5. Video → keyframes → existing image path
With `ai_tagging_use_video`, sample ~3–5 keyframes per referenced video using ffmpeg (scene-change detection with a frame cap, falling back to fixed-interval), then pass those frames through the same inline-image mechanism as `ai_tagging_use_images`. **Why:** captures the dominant "what/where/who" signal of a diary clip at a fraction of native-video cost, adds no new AI plumbing, and keeps token usage predictable across background sweeps. Trade-off: motion and audio cues are lost — acceptable for tagging.

### 6. `pkg/ai/` mirrors KinCart's client shape
New package exposing a `TagSuggester` with text-only and multimodal entry points, strict `ResponseSchema`, and a `knownTags []string` context argument (the hybrid-vocabulary "prefer these, coin ≤2 new" instruction). Constructor returns an error / nil-capable client when `GEMINI_API_KEY` is unset; callers treat that as "feature disabled." **Why:** consistency with a working implementation lowers risk; same env var and graceful-degrade contract.

### 7. In-editor debounce lives in the frontend
The Next.js editor debounces ~4s after the last change and only calls `suggest-tags` if the content actually changed since the last call. **Why:** avoids per-keystroke token spend; keeps the backend endpoint stateless (it just suggests for given text/media).

## Risks / Trade-offs

- **Token cost of background sweeps over years of media** → backfill and media tagging are off by default and per-family opt-in; keyframe cap bounds per-video cost; backfill only processes stale/untagged days, not every day every run.
- **Privacy of sending personal photos/videos to a third party** → image/video tagging are explicit opt-in toggles, documented as such; phase 1 sends only text.
- **Vocabulary drift (model coins near-duplicate tags)** → hybrid vocabulary injects existing tags and caps new ones at ~2 per call; over time the known set stabilizes.
- **Silent writes feel invasive in a personal journal** → default `ai_tagging_auto = false` means nothing is ever written without a chip-click; auto-apply is strictly opt-in and still routes uncertain days to manual review.
- **ffmpeg as a new runtime dependency** → confined to phase 3 and only invoked when `ai_tagging_use_video` is on; absence degrades to "no video frames" rather than an error.
- **Gemini latency on the save path** → on-save retag runs asynchronously (does not block the save response); the editor debounce path is already async to the user.
- **Hash false-negatives** (e.g. asset content changes but filename stays) → accepted; an edited-in-place image with the same name won't retrigger. Mitigated by the explicit "suggest tags" button always being available.

## Migration Plan

- Additive schema: add `pending_tags` and `tags_source_hash` columns to the entries table via GORM auto-migration; existing rows get empty/null values and are treated as stale (eligible for suggestion) without backfilling at migration time.
- Ship phases behind config: phase 1 (text-only) is safe to enable broadly; phases 2–4 gated by per-family flags defaulting off.
- Rollback: unset `GEMINI_API_KEY` (disables all AI behavior) or turn off the family flags; the added columns are inert and can remain.

## Open Questions

- Final default value of the confidence threshold τ (tune during phase 4).
- Exact debounce interval and minimum content-delta before an in-editor auto-suggest fires (start ~4s; tune from usage).
- Keyframe sampling strategy specifics (scene-change sensitivity vs fixed interval) and the per-video frame cap (start 3–5).
- Whether accepting a suggestion should clear only that name from `pending_tags` or re-run the hash (decision: clear the name only; hash updates on next save).
