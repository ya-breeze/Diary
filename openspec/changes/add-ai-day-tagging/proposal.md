## Why

Tagging diary days by hand is tedious, so entries are often left untagged or tagged inconsistently (`travel` vs `traveling` vs `trip`), which undermines tag-based search and the profile's top-tags view. An AI assistant can read a day's text (and optionally its photos and videos) and propose relevant tags drawn from the family's existing vocabulary — turning tagging into a one-tap confirmation while never silently rewriting personal memories.

## What Changes

- Introduce a new `pkg/ai/` package wrapping Google Gemini (`google.golang.org/genai`, model `gemini-2.5-flash-lite`), gated on `GEMINI_API_KEY`. If the key is unset, all AI features degrade gracefully (the rest of the app is unaffected). This is the first AI dependency in the project.
- Add **tag suggestion** for a day from its `Title` + `Body` (phase 1, text-only), returning `{tags: [{name, confidence}]}` via strict structured output.
- Suggestions never overwrite confirmed tags. A new **`pending_tags`** field on an entry holds un-accepted suggestions; the user accepts them per-tag (chip click) to move them into the confirmed `tags`.
- **Hybrid vocabulary**: the family's existing distinct tags are passed as context so the model prefers them and may coin at most ~2 new tags per call. New tags join the known set on subsequent runs.
- **Explicit trigger**: a "suggest tags" action on the entry editor (`POST /v1/items/suggest-tags`, with the draft `date`/`title`/`body` in the request body so it works on unsaved content) returns suggestions without writing anything.
- **In-editor auto-suggest**: while editing, after a short debounce (~4s of inactivity, only if content changed), suggestions are fetched and shown as chips. The user is present, so this path always *suggests* — it never auto-applies.
- **Edit-triggered retagging**: a `tags_source_hash` (hash of `Title` + `Body` + sorted asset filenames) is stored per entry. When an entry is saved and the hash changed, the day is retagged.
- New per-family configuration: `ai_tagging_enabled`, `ai_tagging_use_images`, `ai_tagging_use_video`, `ai_tagging_backfill`, `ai_tagging_auto`.
- `ai_tagging_auto` governs **unattended** triggers (save-and-leave, backfill): when `false` (default) they produce `pending_tags` surfaced as a health issue; when `true`, suggestions with confidence ≥ τ are applied automatically and lower-confidence days are routed to a "solve manually" health issue. In-editor behavior is identical either way.
- **Media tagging** (`ai_tagging_use_images`): include referenced image assets in the suggestion request.
- **Video tagging** (`ai_tagging_use_video`): sample ~3–5 keyframes per video with **ffmpeg** (new runtime dependency) and feed them through the same image path. Native video upload to Gemini is explicitly out of scope.
- **Backfill** (`ai_tagging_backfill`): a new `untagged` health check (4th check alongside `mime`/`orphans`/`refs`) that finds untagged or stale days and surfaces them through the existing `GET /v1/health/issues` + `POST /v1/health/fix` flow.

## Capabilities

### New Capabilities
- `ai-tagging`: AI-assisted tag suggestion for a day from its text and optional media — the suggestion engine, `pending_tags` lifecycle, hybrid vocabulary, confidence-based apply/suggest routing, the explicit and debounced editor triggers, and the per-family `ai_tagging_*` configuration.

### Modified Capabilities
- `entries`: saving an entry computes/stores `tags_source_hash` and triggers retagging when content changed; entry read/edit exposes `pending_tags` and the accept-suggestion action distinct from confirmed `tags`.
- `health`: add a fourth `untagged` check that finds untagged/stale days; under `ai_tagging_auto` it auto-applies confident tags and routes uncertain days to the "solve manually" issue list via the existing issues/fix flow.

## Impact

- **New dependencies**: `google.golang.org/genai` (Go module); `ffmpeg` in the runtime image (phase 3 only).
- **New env/config**: `GEMINI_API_KEY` (server env); per-family `ai_tagging_*` settings.
- **Backend**: new `pkg/ai/` package; new entry model fields (`pending_tags`, `tags_source_hash`) + migration; new `untagged` health check; new/extended entry and config API endpoints.
- **Frontend (Next.js)**: editor "suggest tags" button, debounced auto-suggest, suggested-vs-confirmed tag chips with accept; settings UI for the `ai_tagging_*` toggles; surfacing of tagging health issues.
- **Cost/privacy**: image and video tagging send personal media to Gemini; both are off by default and per-family opt-in. Text-only phase 1 sends only entry text.
- Backward compatible: all features no-op without `GEMINI_API_KEY`; defaults keep behavior suggest-only (no silent writes).
