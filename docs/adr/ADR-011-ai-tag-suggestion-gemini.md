# ADR-011: AI Tag Suggestion via Gemini

## Status

Accepted

(Introduced by OpenSpec change `add-ai-day-tagging`. Supersedes part of ADR-007's check enumeration — see ADR-007.)

## Context and Problem Statement

Tagging diary days by hand is tedious, so entries are often left untagged or tagged inconsistently (`travel` vs `traveling` vs `trip`), which weakens tag-based search and the profile's top-tags view. We want an assistant that reads a day's text (and optionally its photos and videos) and proposes relevant tags drawn from the family's existing vocabulary — without ever silently rewriting a user's personal memories. This is the first time the application would call an external LLM, so the dependency, failure mode, privacy posture, and the boundary between "AI suggests" and "AI writes" all need a deliberate decision.

## Decision Drivers

- Must degrade to today's behaviour when no API key is configured (AI is optional, not required to run)
- Personal text and media must never leave the system without explicit, per-family opt-in
- The user must stay in control: AI must not silently overwrite or delete user-curated tags
- Reuse a proven integration rather than invent a new one (the sibling KinCart project already ships a Gemini client)
- Keep cost and token usage predictable, including over large media histories
- Suggestions should converge on the family's existing vocabulary, not coin endless near-duplicates

## Considered Options

- **Provider**: Google Gemini vs. Anthropic Claude vs. OpenAI vs. a local/self-hosted model. KinCart already integrates Gemini (`google.golang.org/genai`, structured output, multimodal inline parts, "known items" vocabulary injection), so the pattern is proven in-house.
- **Apply model**: AI writes tags directly vs. AI stages suggestions the user accepts. Direct writes are frictionless but invasive on a personal journal.
- **Video handling**: native video upload to Gemini (Files API) vs. sampling a few keyframes through the image path. Native video is richest but expensive and adds new upload plumbing.
- **Backfill mechanism**: a new dedicated subsystem vs. a new check inside the existing health framework (ADR-007).

## Decision Outcome

Chosen: **Google Gemini** (`google.golang.org/genai`, model `gemini-2.5-flash-lite`) behind a new `pkg/ai/` package, mirroring KinCart's client shape — strict `ResponseSchema` structured output (`{tags:[{name,confidence}]}`) and a `knownTags` vocabulary list injected into the prompt ("prefer these, coin at most ~2 new").

- **Env-gated, graceful degradation**: the client is constructed only when `GEMINI_API_KEY` is set; absent the key, every tagging trigger is a no-op and the rest of the app is unaffected. Per-family `ai_tagging_*` settings (all default off) gate the feature beyond the key.
- **Suggest, not apply (default)**: suggestions land in a `pending_tags` field, separate from confirmed `tags`. The user accepts per-tag (chip click) to move a suggestion into `tags`. AI is **additive-only** — it never deletes, renames, or wholesale-replaces a confirmed tag. The two lists are kept disjoint.
- **Confidence routing for unattended triggers**: a per-family `ai_tagging_auto` flag may auto-apply suggestions with confidence ≥ τ, but **only on days with no confirmed tags**. Once a day is curated, unattended triggers only stage suggestions — preventing AI from re-adding a tag the user removed.
- **Media is opt-in**: images (`ai_tagging_use_images`) and video (`ai_tagging_use_video`) are off by default; text-only is the baseline.
- **Video via keyframes**: with video enabled, ~3–5 keyframes per clip are sampled with **ffmpeg** and fed through the same inline-image path. Native video upload is explicitly out of scope.
- **Backfill reuses the health framework**: an `UntaggedCheck` (4th check alongside mime/orphans/refs from ADR-007) finds untagged/stale days and surfaces them via the existing `GET /v1/health/issues` + `POST /v1/health/fix` flow.
- **Staleness signal**: a `tags_source_hash` (hash of title + body + sorted referenced asset filenames) on each entry drives all edit-triggered retagging and the backfill check.

### Pros

- No new infrastructure beyond the Gemini SDK (and ffmpeg for the video phase); backfill rides the existing background-task goroutine (ADR-007)
- Optional and reversible: unset the key or the family flags and behaviour reverts exactly to today
- Privacy-respecting by default — only text is sent unless media toggles are explicitly enabled
- User-curated tags are protected by hard invariants (additive-only, untagged-only auto-apply)
- Consistent with a working integration (KinCart), lowering implementation risk
- Hybrid vocabulary keeps the family's tag set stable over time

### Cons

- Introduces an external runtime dependency on a third-party LLM with its own latency, cost, and availability characteristics
- Media tagging sends personal photos/video frames to a third party (mitigated: opt-in, off by default)
- `gemini-2.5-flash-lite` is a specific model choice that will need revisiting as models evolve
- ffmpeg becomes a runtime dependency for the video phase
- No rejection memory in v1: a user who deletes a tag relies on the "untagged-only auto-apply" guard rather than an explicit dismissed-tags list (deferred enhancement)
- `tags_source_hash` includes asset filenames, so in the text-only phase a pure image change can trigger a redundant text-only retag (harmless; correct once media phases land)

## Related

- ADR-002 (OpenAPI-first) — the `suggest-tags` endpoint and `pending_tags` field follow the spec-first workflow
- ADR-004 (kin-core multitenancy) — all tagging and vocabulary context is family-scoped
- ADR-007 (background tasks) — the backfill `UntaggedCheck` runs in the existing checker goroutine
- ADR-009 (date-keyed entries) — `pending_tags` and `tags_source_hash` are additive fields on the date-keyed entry; upsert identity is unchanged
