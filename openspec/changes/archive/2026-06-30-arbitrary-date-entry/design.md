## Context

Currently the only way to reach an arbitrary date is to type the URL manually. The "New Entry" button always opens today, and the prev/next arrows step through existing entries only. Users who want to backfill a past entry or jump ahead have no UI path.

The frontend is Next.js 15 with React, TanStack Query, and `router.push` for navigation. No backend changes are needed — the existing `GET /api/diary/{date}` and `PUT /api/diary/{date}` endpoints already handle any date, and the page already shows the editor immediately when no entry exists.

## Goals / Non-Goals

**Goals:**
- Let the user navigate to any date from the EntryViewer (via the date badge)
- Let the user navigate to any date from the Sidebar (secondary affordance next to "New Entry")
- Zero new dependencies: use the native `<input type="date">` element
- Keep "New Entry → today" as the primary, zero-friction action

**Non-Goals:**
- Month-view calendar or date range selection
- Bulk entry management
- Any backend change

## Decisions

**Decision 1 — Native `<input type="date">` over a third-party picker**

The native date input is sufficient: it supports keyboard entry (YYYY-MM-DD), has a built-in picker on all modern browsers and mobile OSes, and requires no new packages. Third-party pickers (react-datepicker, etc.) add bundle weight and custom styling without a meaningful UX gain for this use case.

Implementation: a hidden `<input type="date">` is triggered programmatically. Most browsers support `inputEl.showPicker()`; as a fallback, `.click()` works. On `change`, call `router.push(`/diary/${value}`)`.

**Decision 2 — Make the date badge in EntryViewer the click target**

The existing date badge (showing the formatted date with a calendar icon) is the natural affordance. On click it will call `.showPicker()` on a co-located hidden date input. The badge retains its current styling; a subtle `cursor-pointer` and hover ring indicate interactivity.

Alternative considered: replace the badge with a visible `<input type="date">` styled to match — rejected because the native date input is hard to style consistently across browsers and would look inconsistent with the badge system.

**Decision 3 — Small calendar icon button alongside "New Entry" in the Sidebar**

The "New Entry" button keeps its current behavior (→ today's entry in edit mode). A secondary icon-only button (calendar icon, same height) is added immediately to its right. Clicking it opens a hidden date input; on selection the user is routed to `/diary/{date}` **without** forcing edit mode — if an entry exists they see the viewer, if not they get the editor (standard page behavior).

Alternative: dropdown arrow on the "New Entry" button that reveals a calendar. Rejected as more complex to implement and overcomplicates what is a simple nav action.

## Risks / Trade-offs

- **`showPicker()` availability**: Not supported in older browsers (Safari < 15.4). Mitigation: fall back to `.click()` on the input, which opens the native picker in all browsers except Firefox on desktop (where it shows the input inline instead). Acceptable given the app's target audience.
- **Native date input styling**: The picker popover is OS/browser-native and cannot be styled. Mitigation: the input is always hidden; only the trigger (badge or icon button) is visible, so styling is fully under our control.
- **Date input default value**: The hidden input should be pre-filled with the current page's date so the picker opens at the right month. The `entry.date` prop is available in `EntryViewer`; today's date can be used in the Sidebar.

## Migration Plan

Frontend-only change. No deploy coordination needed. The change is self-contained within two components (`EntryViewer.tsx`, `Sidebar.tsx`) and can be deployed in a single release.

Rollback: revert the two component files.

## Open Questions

None — the approach is straightforward and all required APIs are already in place.
