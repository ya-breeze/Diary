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
- [x] 2.3 Implement text-only `SuggestTags(ctx, title, body, knownTags)` using `gemini-2.5-flash-lite`, strict `ResponseSchema` for `{tags:[{name,confidence}]}`
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
- [x] 4.5 E2E (`e2e/tests/ai-tagging.spec.ts`, registered as a playwright project): toggle persists; suggest button visibility follows the setting. **Passed against diary-wip (3/3); entry + navigation suites also green (no regressions).** Suggestion content needs a live key (non-deterministic), so not asserted.

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

- [x] 7.1 Per-family `AITaggingBackfill`/`AITaggingAuto` columns + `SetFamilyAISettings`; flags on `FamilyResponse`/`FamilySettingsRequest`; server config `AITaggingThreshold` (τ, default 0.8)
- [x] 7.2 `UntaggedCheck` (`pkg/checker/check_untagged.go`) satisfies `checker.Check`; no-op unless suggester enabled AND family `AITaggingBackfill`; bounded by `maxBackfillPerRun`; treats untagged OR stale (hash mismatch) days as candidates
- [x] 7.3 Registered `untagged` in the checker task (instance `checks`, injected suggester); `selectChecks` knows it; runs in the 24h sweep and via `RunForFamily` (fix path)
- [x] 7.4 Auto mode: confident (≥τ) untagged days → `Fixable:true` with a `fix()` that additively writes tags; uncertain → `pending_tags` + non-fixable "solve manually"
- [x] 7.5 Non-auto mode: stages `pending_tags`, non-fixable issue "review on the entry"
- [x] 7.6 On-save retag now routes through auto/non-auto (confident+untagged → apply; else stage), using the `AITaggingThreshold`
- [x] 7.7 Settings UI: backfill + auto toggles on the profile page (shown when AI tagging is enabled)
- [x] 7.8 Tests (`check_untagged_test.go`): disabled suggester, backfill off, non-auto stages pending, auto-confident fix applies, auto-uncertain stages, tagged days skipped

## 7b. Phase 4 — backfill UX refinements

- [x] 7b.1 Auto mode applies confident tags **during the sweep** (no manual "Fix" click); resolved days emit no issue
- [x] 7b.2 Staging `pending_tags` stamps `tags_source_hash`, so already-suggested days aren't re-queried every sweep (only on content change) — added `TestUntaggedDoesNotRequeryStagedDays`
- [x] 7b.3 `untagged` issues are non-fixable review items; health panel renders them as **links to the entry** (`/diary/{date}?edit=true`), no generic Fix button
- [x] 7b.4 Spec deltas updated (health): apply-on-sweep, no-re-query, review-link requirement

## 8. Cross-cutting & verification

- [x] 8.1 Document `GEMINI_API_KEY` + the AI tag suggestion feature in `README.md` (ffmpeg/`ai_tagging_*` media+backfill flags arrive with phases 2–4)
- [x] 8.2 `make build` clean; backend + frontend lint clean for changed files (remaining findings are pre-existing baseline)
- [x] 8.3 Full E2E green against diary-wip; graceful degradation confirmed — no `GEMINI_API_KEY` on the stack, suggest endpoint returns 503, app otherwise unaffected
- [ ] 8.4 On merge (last step before finishing the PR): flip `docs/adr/ADR-011` status from `Proposed` to `Accepted` and archive this OpenSpec change
