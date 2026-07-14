## Why

The entry viewer renders tags in its header, but on a mobile-width viewport that header constrains the tag region and clips its contents. Tags are part of an entry's context and must remain available alongside date navigation and editing controls.

## What Changes

- Give entry tags a dedicated, wrapping row on mobile-sized viewports.
- Preserve the current compact, single-row viewer header on tablet and desktop viewports.
- Add responsive end-to-end coverage for an entry with multiple tags.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `entries`: Require every displayed entry tag to remain visible and usable on narrow viewports.

## Impact

- `next-frontend/src/components/diary/EntryViewer.tsx` header layout and responsive Tailwind classes.
- Diary Playwright coverage for mobile tag visibility.
- No API, persistence, or dependency changes.
