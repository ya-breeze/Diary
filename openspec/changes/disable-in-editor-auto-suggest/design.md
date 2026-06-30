## Context

The EntryEditor has a `useEffect` that debounces suggestion fetches 4 seconds after the user stops typing. The trigger fires not just on user input but also on component mount (when `aiEnabled` transitions from `false` to `true` during initial API load), so suggestions appear without the user touching anything. The fix is a pure frontend removal — no backend changes needed.

## Goals / Non-Goals

**Goals:**
- Remove the debounced auto-fetch entirely
- Keep the explicit ✨ button as the only way to trigger a suggestion fetch
- Keep pre-loaded `pendingTags` (staged by backend unattended triggers) surfacing as chips on editor open

**Non-Goals:**
- Adding any new trigger mechanism
- Changing backend logic
- Changing how accepted/dismissed tags are persisted

## Decisions

**Delete the useEffect, not gate it.** We could add an `isDirty` guard (only auto-suggest after the user has edited), but that still results in automatic network requests. The user wants explicit control, so removing the auto-trigger entirely is cleaner and simpler than trying to refine when it fires.

**`SUGGEST_DEBOUNCE_MS` constant becomes dead code** once the `useEffect` is gone — remove it too.

## Risks / Trade-offs

- **Less discovery**: Users who don't know the button exists won't get suggestions. Acceptable trade-off — the ✨ button is visible whenever `aiEnabled` is true.
- **No rollback complexity**: It's a pure deletion; reverting is trivial if needed.
