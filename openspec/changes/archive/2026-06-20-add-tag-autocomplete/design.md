## Context

The entry editor (`EntryEditor.tsx`) stores tags as a single comma-separated text field. Tags are persisted per `(family, date)` as a string list. There is currently no way for a writer to see which tags already exist, so the same concept is often tagged inconsistently across entries. A separate, in-flight change (`add-ai-day-tagging`) adds AI-based suggestions; this change is deliberately the simpler, always-on, zero-cost complement: complete what the user is already typing from the existing vocabulary.

## Goals / Non-Goals

**Goals:**
- Help users reuse existing tags while typing, with no setup, no API key, and no cost.
- Keep it independent of the AI feature and its per-family flags.
- Be robust regardless of how many entries are currently loaded in the client.

**Non-Goals:**
- AI/model-based suggestion of *new* tags (that is `add-ai-day-tagging`).
- Tag renaming, merging, or a managed taxonomy.
- Cross-family vocabulary.

## Decisions

### 1. Dedicated `GET /v1/tags` endpoint, not client-derived
Return the family's distinct tags from the backend rather than deriving them from the entries already in the client cache. **Why:** the client only holds a paginated slice of entries, so a client-only derivation would miss tags from entries not currently loaded. A backend `SELECT DISTINCT`-style aggregation over all of the family's items is complete and cheap. Reuses a `Storage.GetDistinctTags(familyID)` helper (dedupe + sort in Go, matching how the AI change sources its vocabulary).

### 2. Fetch once on editor open, filter on the client
The editor fetches the full tag list when it opens and filters the dropdown locally against the active token as the user types. **Why:** a family's tag vocabulary is small and changes rarely; one fetch per editor session avoids a request per keystroke and keeps the typeahead instant. No debounce needed.

### 3. Token-aware completion of a comma-separated field
The field holds `tag1, tag2, partial`. Autocomplete operates on the **last** token (after the final comma), filters existing tags by case-insensitive prefix/substring, and excludes tags already present earlier in the field. Selecting a suggestion replaces just that last token and appends `, `. **Why:** preserves the existing single-field UX (no migration to a token/chip input), while making completion feel natural.

### 4. Non-AI and ungated
The endpoint and the editor behavior have no dependency on `GEMINI_API_KEY` or the `ai_tagging_*` family settings. **Why:** consistent vocabulary is valuable to everyone and should not require enabling (or paying for) the AI feature.

## Risks / Trade-offs

- **Overlap with `add-ai-day-tagging`'s `GetDistinctTags`** → both changes may introduce the same storage helper. Mitigation: identical signature/semantics; whichever merges second drops the duplicate. Low risk, trivial conflict.
- **Large vocabularies** → a family with thousands of distinct tags makes the dropdown long. Mitigation: filter by the active token and cap the rendered list (e.g. top N matches).
- **Stale list within a long editing session** → tags created elsewhere mid-session won't appear until reopen. Acceptable; the list is a convenience, not a constraint.
