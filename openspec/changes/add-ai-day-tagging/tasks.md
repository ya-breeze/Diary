# Tasks

Organized by the four phases from the proposal. Each phase is independently shippable.

## 1. Phase 1 — Data model & migration (text-only foundation)

- [ ] 1.1 Add `PendingTags StringList` and `TagsSourceHash string` fields to the `Item` model
- [ ] 1.2 Add GORM auto-migration for the two new columns; verify existing rows get empty/null values and are treated as stale
- [ ] 1.3 Implement `ComputeTagsSourceHash(title, body, assets)` helper: hash of `title + body + join(sorted(GetAssetsFromMarkdown(body)), ",")`
- [ ] 1.4 Add `pendingTags` to the entry read/edit schemas in `api/openapi.yaml`; run `make generate`

## 2. Phase 1 — `pkg/ai/` suggester (text-only)

- [ ] 2.1 Add `google.golang.org/genai` to `go.mod`; run `make` to confirm it builds
- [ ] 2.2 Create `pkg/ai/` with a `TagSuggester` constructor reading `GEMINI_API_KEY`; return a disabled/nil-capable client when unset (graceful degrade)
- [ ] 2.3 Implement text-only `SuggestTags(ctx, title, body, knownTags []string)` using model `gemini-2.0-flash`, strict `ResponseSchema` for `{tags:[{name,confidence}]}`
- [ ] 2.4 Implement hybrid-vocabulary prompt: inject `knownTags`, instruct "prefer these, coin at most ~2 new"; exclude already-confirmed tags from results
- [ ] 2.5 Unit tests for schema decoding, empty-text short-circuit, and disabled-client behavior

## 3. Phase 1 — Suggestion wiring & API

- [ ] 3.1 Add per-family `ai_tagging_enabled` config (storage + read/update); default off
- [ ] 3.2 Implement `POST /v1/items/suggest-tags` (draft `date`/`title`/`body` in body, matching the existing `/v1/items` resource style) returning `{tags:[{name,confidence}]}` without writing; 0 results / unavailable when key or flag missing
- [ ] 3.3 Implement "accept suggestion" path: move a name from `pending_tags` into confirmed `tags`
- [ ] 3.4 On entry save: recompute `tags_source_hash`; when changed, enqueue async retag that writes suggestions into `pending_tags` (non-auto default)
- [ ] 3.5 Source the family's distinct existing tags as `knownTags` for suggestion calls
- [ ] 3.6 Backend tests: suggest endpoint, accept moves pending→confirmed, no-op save skips retag, hash stability under asset reorder

## 4. Phase 1 — Frontend (Next.js)

- [ ] 4.1 Render `pending_tags` as visually distinct suggestion chips with one-tap accept
- [ ] 4.2 Add "suggest tags" button in the editor calling the suggest endpoint
- [ ] 4.3 Add in-editor debounced auto-suggest (~4s idle, only if content changed since last call); always suggest, never apply
- [ ] 4.4 Settings UI toggle for `ai_tagging_enabled`
- [ ] 4.5 E2E: button suggests chips; accepting a chip confirms the tag; debounce fires once per content change

## 5. Phase 2 — Image-based suggestions

- [ ] 5.1 Add per-family `ai_tagging_use_images` config (default off)
- [ ] 5.2 Extend `TagSuggester` with multimodal input: include referenced image assets as inline parts when enabled
- [ ] 5.3 Resolve image asset paths from `GetAssetsFromMarkdown` and load bytes with MIME detection
- [ ] 5.4 Settings UI toggle + privacy note that images are sent to Gemini
- [ ] 5.5 Tests: images included only when flag on; text-only when off

## 6. Phase 3 — Video keyframe suggestions

- [ ] 6.1 Add per-family `ai_tagging_use_video` config (default off)
- [ ] 6.2 Add `ffmpeg` to the runtime image (Docker); document the dependency
- [ ] 6.3 Implement keyframe extraction (~3–5 frames/video, scene-change with fixed-interval fallback, frame cap)
- [ ] 6.4 Feed extracted frames through the existing image path; degrade gracefully if extraction unavailable
- [ ] 6.5 Settings UI toggle + privacy note; tests for frame-extraction fallback

## 7. Phase 4 — Backfill health check & auto mode

- [ ] 7.1 Add per-family `ai_tagging_backfill` and `ai_tagging_auto` config (default off); add confidence threshold τ
- [ ] 7.2 Implement `UntaggedCheck` satisfying `checker.Check`: emit an Issue per untagged/stale day, skipped unless `ai_tagging_backfill` enabled
- [ ] 7.3 Register `untagged` in the health Runner; confirm family scoping via `RunForFamily`
- [ ] 7.4 Auto mode: confident (≥τ) days get a populated `fix func()` writing tags and `Fixable:true`; uncertain days store `pending_tags`, `Fixable:false`, "solve manually" message
- [ ] 7.5 Non-auto mode: store `pending_tags`, report non-fixable issue (user accepts chips on the day)
- [ ] 7.6 Apply on-save unattended path through the same auto/non-auto routing
- [ ] 7.7 Settings UI toggles for backfill + auto mode
- [ ] 7.8 Tests: check skipped when disabled; fixable vs solve-manually routing by confidence; fix-all writes confident tags

## 8. Cross-cutting & verification

- [ ] 8.1 Update Diary `CLAUDE.md`/docs: `GEMINI_API_KEY`, `ai_tagging_*` flags, ffmpeg dependency, `pkg/ai/` overview
- [ ] 8.2 Run `make lint` and `make build`; fix issues
- [ ] 8.3 Run full E2E against the WIP stack; confirm graceful degradation with no API key set
- [ ] 8.4 On merge (last step before finishing the PR): flip `docs/adr/ADR-011` status from `Proposed` to `Accepted` and archive this OpenSpec change
