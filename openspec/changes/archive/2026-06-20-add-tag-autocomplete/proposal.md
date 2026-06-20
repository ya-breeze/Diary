## Why

When writing an entry, users type tags freely into a comma-separated field with no visibility into what tags already exist. This causes vocabulary drift (`travel` vs `traveling` vs `trip`), which fragments tag-based search and the profile's top-tags view. Surfacing the family's existing tags as the user types — a simple, AI-free autocomplete — nudges everyone toward a consistent shared vocabulary.

## What Changes

- Add a backend endpoint `GET /v1/tags` returning the family's distinct existing tags (deduplicated, sorted).
- In the entry editor's tags field, show a typeahead dropdown of matching existing tags as the user types the current (last, comma-separated) token; selecting one completes that token.
- The suggestion list is drawn **only from the family's existing tags** — no AI, no API key, no per-keystroke model calls. It always works and costs nothing.
- This is independent of the AI tag-suggestion feature: it is not gated behind any `ai_tagging_*` flag and does not use Gemini.

## Capabilities

### New Capabilities
- `tag-autocomplete`: typeahead completion of the entry tags field from the family's existing tag vocabulary, backed by a `GET /v1/tags` endpoint.

### Modified Capabilities
<!-- None: this adds a new field behavior without changing existing entry requirements. -->

## Impact

- **Backend**: new `GET /v1/tags` endpoint (OpenAPI + handler + route); new `Storage.GetDistinctTags(familyID)` method returning deduplicated, sorted tags. No schema change.
- **Frontend (Next.js)**: a tags-input autocomplete in `EntryEditor` (fetch `/v1/tags` on open, filter the dropdown client-side by the active token).
- **No new dependencies, no migration, no config.** Family-scoped like all entry data.
- Independent of and non-overlapping with `add-ai-day-tagging`; both may add a `GetDistinctTags` storage helper, so whichever merges second resolves a trivial overlap.
