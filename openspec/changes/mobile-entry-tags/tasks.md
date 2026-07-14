## 1. Responsive viewer layout

- [x] 1.1 Restructure `EntryViewer`'s header classes so tags occupy a full, wrapping row below the `md` breakpoint and the date navigation/Edit controls remain accessible on their own row.
- [x] 1.2 Preserve the existing single-row header layout and tag badge variants at `md` and wider breakpoints.

## 2. Verification

- [x] 2.1 Add a Playwright mobile-viewport scenario for an entry with a mood tag and multiple standard tags, asserting every tag badge is visible.
- [x] 2.2 Run the focused Playwright scenario and the relevant frontend validation commands; resolve any failures.
