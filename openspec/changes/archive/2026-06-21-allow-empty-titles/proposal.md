## Why

The title field in the diary editor currently blocks saving if left blank, but many entries are naturally titled by their date alone — forcing a title adds friction without value. The backend already accepts empty titles; only the frontend Zod schema prevents it.

## What Changes

- Remove the `min(1)` constraint from the title field's Zod schema in `EntryEditor.tsx` so empty titles are allowed on save.
- Update the spec to reflect that title is optional (not required).
- No changes to display behavior — both `EntryCard` and `EntryViewer` already fall back to `'Untitled'` when title is empty.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `entries`: The "Title is required" validation rule is removed; title becomes optional.

## Impact

- `next-frontend/src/components/diary/EntryEditor.tsx` — Zod schema change (one line)
- `openspec/specs/entries/spec.md` — scenario update
- No backend changes
- No API changes
- No new dependencies
