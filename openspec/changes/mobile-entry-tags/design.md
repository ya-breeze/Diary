## Context

`EntryViewer` currently places tags, date navigation, and the edit action in one flex row. On viewports below the `md` breakpoint, the date controls retain their intrinsic width while the tags container is both width-constrained and `overflow-hidden`; tags can therefore render outside the visible region.

## Goals / Non-Goals

**Goals:**

- Keep every confirmed tag visible and usable on a mobile viewport.
- Keep date navigation and editing accessible without horizontal page overflow.
- Preserve the existing desktop header arrangement.
- Verify the responsive behavior in the existing Playwright suite.

**Non-Goals:**

- Changing tag ordering, styling variants, or persistence.
- Adding tag filtering or navigation from viewer badges.
- Changing the editor, API, or database.

## Decisions

### Use a mobile-only tag row above the controls

At widths below `md`, the header will become a wrapping container: tags take a full-width first row and wrap naturally, while date navigation and Edit occupy a second row. At `md` and above, existing flex sizing restores the single-row header (tags / date navigation / Edit as three columns), and the tag column itself wraps its badges onto additional rows instead of clipping them.

This exposes all tags without a secondary interaction. A horizontal tag scroller was considered, but it keeps content off-screen and is less discoverable; truncation/clipping is the previous failure mode and is rejected at every viewport width.

### Keep the existing tag badges and data model

The component will continue to render the first tag as the mood badge and remaining tags as standard badges. Only layout classes change, so the behavior remains covered by the existing entry tag contract.

### Add an end-to-end narrow-viewport assertion

The Playwright test will create an entry with a mood and multiple standard tags, set a mobile viewport, then verify each tag badge is visible. This protects the user-facing layout outcome rather than coupling the test to Tailwind class names.

## Risks / Trade-offs

- [Many or long tag names make the header taller] → Tags deliberately wrap, so content remains readable and the date controls move down predictably.
- [Responsive classes could affect desktop alignment] → Keep the row-wrapping restructure below `md`; at `md` and above retain the three-column layout but let the tag column wrap its badges so no tag is clipped on desktop either.
- [A visibility check could pass while a badge is overlapped] → Use Playwright's visible assertion after navigation, which detects clipped or `display:none` content.

## Migration Plan

No data or deployment migration is required. Deploy with the normal frontend release; rollback is a reversion of the layout and test change.

## Open Questions

None.
