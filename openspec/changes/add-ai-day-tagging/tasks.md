# Tasks

Organized by the four phases from the proposal. Each phase is independently shippable.

## 1. Phase 1 — Data model & migration (text-only foundation)

- [x] 1.1 Add `PendingTags StringList` and `TagsSourceHash string` fields to the `Item` model
- [x] 1.2 Add GORM auto-migration for the two new columns (additive nullable columns handled by existing `autoMigrateModels` AutoMigrate; existing rows get NULL → empty → stale)
- [x] 1.3 Implement `ComputeTagsSourceHash(title, body)` helper in `pkg/utils` (strips image refs from the text component, re-adds sorted asset filenames, so reordering is invariant); covered by `tags_hash_test.go`
- [x] 1.4 Add `pendingTags` to `ItemsResponse` in `api/openapi.yaml`, run `make generate`, surface it in `GetItems`/`PutItems`. Storage `PutItem` sets the hash and keeps pending/confirmed disjoint; added `SetPendingTags` for the async retag

## 2. Phase 1 — `pkg/ai/` suggester (text-only)

- [x] 2.1 Add `google.golang.org/genai` to `go.mod` (v1.61.0); builds clean
- [x] 2.2 Create `pkg/ai/` with `NewSuggester` reading `GEMINI_API_KEY`; returns a disabled suggester when unset (graceful degrade)
- [x] 2.3 Implement text-only `SuggestTags(ctx, title, body, knownTags)` using `gemini-2.0-flash`, strict `ResponseSchema` for `{tags:[{name,confidence}]}`
- [x] 2.4 Hybrid-vocabulary prompt: inject `knownTags`, instruct "prefer these, ≤2 new". (Exclusion of the entry's confirmed tags is owned by the service/storage layer, not the suggester — keeps the suggester focused.)
- [x] 2.5 Unit tests: disabled suggester, blank-text short-circuit, prompt building, schema decode/clamp/dedupe, invalid JSON

## 3. Phase 1 — Suggestion wiring & API

- [x] 3.1 Per-family `ai_tagging_enabled` config: `Family.AITaggingEnabled` column, `SetFamilyAITaggingEnabled`, `aiTaggingEnabled` on `FamilyResponse`, `PATCH /v1/family`
- [x] 3.2 `POST /v1/items/suggest-tags` (draft `date`/`title`/`body` in body) returning `{tags:[{name,confidence}]}` without writing; 503 when key or family flag missing
- [x] 3.3 Accept path: client appends an accepted name to confirmed `tags` and saves; storage `PutItem` prunes it from `pending_tags` (disjoint invariant)
- [x] 3.4 On save: recompute `tags_source_hash`; when changed, async retag writes suggestions into `pending_tags` (suggest-only, non-auto default), confirmed tags excluded
- [x] 3.5 `knownTags` sourced from `GetDistinctTags(familyID)` for suggest + retag
- [x] 3.6 Backend tests: suggest 401/503/200 paths (fake suggester); `GetDistinctTags` dedupe/sort + AI toggle round-trip; hash stability under asset reorder (utils)

## 4. Phase 1 — Frontend (Next.js)

- [x] 4.1 Render `pendingTags` as visually distinct suggestion chips with one-tap accept (seeded from `entry.pendingTags`)
- [x] 4.2 "Suggest tags" button in the editor calling the suggest endpoint
- [x] 4.3 In-editor debounced auto-suggest (~4s idle, content-change-gated via `lastSuggestedRef`); always suggest, never apply
- [x] 4.4 Settings UI toggle for `ai_tagging_enabled` on the profile page
- [~] 4.5 E2E spec written (`e2e/tests/ai-tagging.spec.ts`): toggle persists; suggest button visibility follows the setting. Suggestion content needs a live key (non-deterministic), so not asserted. **Run pending a diary-wip deploy.**

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
