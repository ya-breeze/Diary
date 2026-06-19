## Why

On mobile (Android), closing the diary after editing a day requires several taps of the Back button. Each transition into and out of edit mode pushes a new browser-history entry, so Back walks the user back through `viewer → edit mode → viewer → list` instead of closing the app. The user expects Back to dismiss the entry in one press.

## What Changes

- Entering edit mode (`Edit` button) uses `router.replace` instead of `router.push`, so `?edit=true` does not add a history entry.
- Exiting edit mode (Cancel) and saving an entry use `router.replace` to return to `/diary/[date]` without stacking history.
- Cancelling a brand-new (no existing entry) editor uses `router.replace('/diary')` instead of `push`.
- **BREAKING (behavior)**: the browser Back button no longer closes the editor from within edit mode. Because edit mode is now a replaced URL state, Back returns to the previous real page (the entry list), not the viewer.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `entries`: the "Edit an entry" requirement's URL/history behavior changes — edit mode is a replaced URL state rather than a pushed history entry, and the Back button no longer closes the editor.

## Impact

- `next-frontend/src/app/(dashboard)/diary/[date]/page.tsx` — `handleEdit`, `handleCloseEdit`, new-entry `onClose`.
- `next-frontend/src/components/diary/EntryEditor.tsx` — `onSubmit` post-save navigation.
- No backend, API, or data changes. Pure client-side navigation behavior.
