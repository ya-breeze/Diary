## Why

Navigating to an arbitrary date requires typing a URL manually — there's no UI affordance to jump to or create an entry for a specific past or future date. Users who want to backfill entries or look up a specific day have to know the `/diary/YYYY-MM-DD` convention.

## What Changes

- The date badge in the EntryViewer becomes clickable, opening an inline date picker that navigates to the selected date.
- The "New Entry" button in the Sidebar gains an optional date picker: clicking the button normally still opens today's entry; a secondary affordance (e.g., a small calendar icon or dropdown arrow next to the button) lets the user pick a different date before opening the editor.

## Capabilities

### New Capabilities
- `date-jump`: Navigate to any arbitrary date (past or future) via an interactive date picker, accessible from both the entry viewer and the sidebar. If no entry exists for the chosen date, the editor opens immediately (existing spec behavior). If an entry exists, the viewer opens.

### Modified Capabilities
- `entries`: The date badge in the viewer becomes an interactive element (date picker trigger). No change to the underlying requirements — routing, editor open-on-no-entry, and upsert semantics are unchanged.

## Impact

- `EntryViewer.tsx` — date badge becomes a clickable date-picker trigger
- `Sidebar.tsx` — "New Entry" button area gains a secondary affordance for choosing a date
- No backend changes required
- No new dependencies required if using the native `<input type="date">` element; a third-party date picker library may be considered in design if the native control is too limited on target browsers
