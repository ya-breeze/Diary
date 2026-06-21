## Context

Title validation currently lives only in the frontend Zod schema (`EntryEditor.tsx:33`). The backend has no `minLength` constraint on title in the OpenAPI spec or Go handlers — a backend integration test explicitly documents that empty titles return 200. Both list and detail views already render `entry.title || 'Untitled'` as a fallback, so the display layer is already prepared.

## Goals / Non-Goals

**Goals:**
- Allow users to save a diary entry with a blank title field
- Keep display behavior unchanged (`'Untitled'` fallback stays as-is)

**Non-Goals:**
- Changing the placeholder text or labelling the field as "(optional)" in the UI
- Adding any backend validation or OpenAPI schema constraint on title
- Changing how untitled entries appear in search results or AI tagging

## Decisions

**Remove `min(1)` from the Zod schema rather than making title fully optional at the type level.**
The `title` field remains a `z.string()` — it just no longer requires at least one character. This keeps the TypeScript type as `string` (not `string | undefined`), which requires no changes to the `saveEntry` call or API layer.

Alternative considered: make the field `z.string().optional()` and update call sites to pass `undefined`. Rejected — unnecessary complexity since the API already accepts an empty string, and `''` vs `undefined` has the same stored effect.

## Risks / Trade-offs

- Entries with no title all display as "Untitled" — users cannot distinguish them in the entry list by title alone. Accepted: the date is always shown alongside, so entries remain identifiable by date.
