## Context

The diary entry page (`/diary/[date]`) renders either a viewer or a full-screen editor overlay. Edit state is reflected in the URL via the `?edit=true` query param. Today every transition uses `router.push`:

- Tapping **Edit** pushes `/diary/[date]?edit=true`.
- Tapping **Cancel** pushes `/diary/[date]`.
- Saving pushes `/diary/[date]`.

On a fresh app open the resulting history stack after one edit+save is:

```
/diary  →  /diary/[date]  →  /diary/[date]?edit=true  →  /diary/[date]
```

Android Back then walks the whole stack, so closing the entry takes 3–4 presses. The user wants Back to return to the list (and then close) in the fewest presses.

## Goals / Non-Goals

**Goals:**
- Back button from the entry returns to the list in one press, and closes the app in two.
- Keep `?edit=true` deep-linkable so an editor can still be opened directly via URL.

**Non-Goals:**
- No PWA manifest / `popstate` interception work.
- No removal of URL-based edit state (the param stays; only how we navigate to it changes).
- No backend/API changes.

## Decisions

**Use `router.replace` for all edit-mode transitions instead of `router.push`.**
Edit mode is a UI sub-state of viewing an entry, not a distinct destination a user would deliberately navigate "back" to. Replacing the current history entry keeps the stack flat (`/diary → /diary/[date]`), so Back behaves intuitively.

Alternatives considered:
- *Pure React state, drop the URL param* — simplest stack-wise, but loses deep-linking/refresh-into-edit. Rejected to preserve existing URL contract.
- *Intercept `popstate`* — fragile, more code, platform-specific. Rejected.

## Risks / Trade-offs

- [Back no longer closes the editor from within edit mode] → This is intended. Back from edit mode now goes to the list. The Cancel/Save buttons remain the in-editor exit paths, and the spec scenario is updated to match.
- [A direct deep link to `?edit=true` followed by Back] → Lands on whatever preceded it in history (or the list if opened fresh), consistent with the new model. Acceptable.
