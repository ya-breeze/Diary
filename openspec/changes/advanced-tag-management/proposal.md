## Why

Tags are central to navigating a diary, but the current tooling is thin: there is no way to browse or search by tag from a tag list, the only usage insight is an unranked "top 5" on the profile, tags can only be edited as raw comma-separated text (removing one means deleting characters by hand), and a typo'd or unwanted tag can never be fixed or removed across entries. This change makes tags first-class: editable as chips, browsable and countable on a dedicated page, and manageable (rename/delete) across the whole family.

## What Changes

- **Tag chip input in the editor** — the comma-separated tags text field is replaced by a row of `[tag ×]` chips plus an inline input for adding new tags. Each confirmed tag gets an X button to remove it directly. Existing-tag autocomplete continues to work, now anchored to the inline input.
- **Tag usage counts** — a new `GET /v1/tags/stats` endpoint returns each distinct family tag with the number of entries using it. The existing `GET /v1/tags` (`string[]`) is left unchanged so editor autocomplete is unaffected.
- **Dedicated Tags page** — lists every tag with its count, sorted by count descending. Reached by making the profile page's "X Tags" stat card a link.
- **Browse by tag** — tapping a tag shows all entries carrying it, using the existing `GET /v1/items?tags=` filter (OR semantics across multiple tags).
- **Tag filter on search** — the search page gains a tag filter row (`[tag ×]` chips + add-tag) combined with free-text search; multiple selected tags use OR among themselves, AND with the text.
- **Rename a tag across all entries** — `PATCH /v1/tags/{name}` with `{newName}` renames the tag on every family entry. If `newName` already exists on an entry, the two merge (no duplicate tag on a single entry; the rename is not blocked).
- **Delete a tag across all entries** — `DELETE /v1/tags/{name}` removes the tag from every family entry. The Tags page exposes rename (✎) and delete (🗑) per tag, each behind a confirmation step ("remove from N entries?").

## Capabilities

### New Capabilities
- `tag-management`: the family's tag vocabulary as a managed resource — usage counts (`GET /v1/tags/stats`), the dedicated Tags page (list + counts + browse), and cross-entry rename (`PATCH /v1/tags/{name}`) and delete (`DELETE /v1/tags/{name}`).

### Modified Capabilities
- `entries`: the editor's tags field becomes a chip input (per-tag X removal + inline add) instead of a comma-separated text field.
- `tag-autocomplete`: autocomplete now completes the inline add-tag input of the chip field rather than the last token of a comma-separated string.
- `search`: the search page exposes a tag filter as selectable chips (OR among tags, AND with text), surfacing the already-supported `tags` query param in the UI.
- `profile`: the "Tags" statistic card links to the new Tags page.

## Impact

- **Backend (Go):** new handlers for `GET /v1/tags/stats`, `PATCH /v1/tags/{name}`, `DELETE /v1/tags/{name}`; storage gains tag-count aggregation and bulk rename/delete that load each affected entry, rewrite its JSON tags array, and re-save (SQLite JSON arrays cannot be bulk-updated in a single statement). `api/openapi.yaml` gains the three operations and a `TagStat` schema; regenerate via `make generate`.
- **Frontend (Next.js):** new Tags page route; `EntryEditor` tags field rewritten as a chip component; search page gains a tag filter row; profile "Tags" card becomes a link; `lib/api/diary.ts` and `types/` gain the new endpoints and types.
- **No data migration:** tag storage format is unchanged; rename/delete operate on existing entries in place.
- **E2E:** new/updated Playwright specs for the chip editor, Tags page, browse, search filter, and rename/delete flows.
