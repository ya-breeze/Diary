## 1. Responsive viewer layout

- [x] 1.1 Restructure `EntryViewer`'s header classes so tags occupy a full, wrapping row below the `md` breakpoint and the date navigation/Edit controls remain accessible on their own row.
- [x] 1.2 Preserve the existing single-row (three-column) header layout and tag badge variants at `md` and wider breakpoints.
- [x] 1.3 Let the tag column wrap its badges at `md` and above so a many-tag entry is never clipped on desktop.

## 2. Verification

- [x] 2.1 Add a Playwright mobile-viewport scenario for an entry with a mood tag and multiple standard tags, asserting every tag badge is visible.
- [x] 2.2 Add a Playwright wide-viewport scenario asserting a many-tag entry wraps and every badge stays visible.
- [x] 2.3 Run the focused Playwright scenarios and the relevant frontend validation commands; resolve any failures.
