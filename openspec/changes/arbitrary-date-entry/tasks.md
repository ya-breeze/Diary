## 1. EntryViewer — interactive date badge

- [ ] 1.1 Add a hidden `<input type="date">` inside `EntryViewer.tsx`, ref'd via `useRef`
- [ ] 1.2 Set the hidden input's `defaultValue` to `entry.date` so the picker opens at the right month
- [ ] 1.3 Wire the date badge's `onClick` to call `inputRef.current?.showPicker()` (fallback: `.click()`)
- [ ] 1.4 Add `cursor-pointer` and a hover ring to the date badge to indicate interactivity
- [ ] 1.5 On the hidden input's `onChange`, call `router.push(`/diary/${e.target.value}`)` to navigate
- [ ] 1.6 Verify: clicking the badge opens the picker; selecting a date navigates correctly; dismissing does nothing

## 2. Sidebar — secondary date-jump affordance

- [ ] 2.1 Import `Calendar` from `lucide-react` (already available) and `useRef` in `Sidebar.tsx`
- [ ] 2.2 Add a hidden `<input type="date">` ref'd via `useRef`, defaulting to today's date
- [ ] 2.3 Add a small icon-only button (calendar icon) immediately to the right of the "New Entry" button
- [ ] 2.4 Wire the calendar icon button's `onClick` to call `inputRef.current?.showPicker()` (fallback: `.click()`)
- [ ] 2.5 On the hidden input's `onChange`, call `router.push(`/diary/${e.target.value}`)` (no `?edit=true`)
- [ ] 2.6 Verify: "New Entry" still navigates to today in edit mode; calendar icon opens picker and navigates to selected date

## 3. EntryEditor — date field as live navigation

- [ ] 3.1 In `EntryEditor.tsx`, track a `isDirty` flag using react-hook-form's `formState.isDirty`
- [ ] 3.2 Add a `pendingDate` state that holds the newly selected date value before the user confirms
- [ ] 3.3 Replace the plain `<Input type="date">` registration with a controlled `onChange` handler that, instead of updating the form directly, sets `pendingDate` and shows a confirmation dialog when `isDirty` is true; otherwise applies the date and triggers a reload
- [ ] 3.4 Implement the confirmation dialog (use the existing `Modal` or `Drawer` component): show options "Save to [current date]", "Move to [new date]", and "Cancel"
- [ ] 3.5 "Save to current date" path: call `saveEntry.mutateAsync(...)` with the current form values at the current date, then reload editor content for `pendingDate`
- [ ] 3.6 "Move to new date" path: call `setValue('date', pendingDate)` only — keep all other form field values, dismiss the dialog
- [ ] 3.7 "Cancel" path: dismiss the dialog, leave the date field unchanged
- [ ] 3.8 Implement the reload-for-date logic: fetch the entry for `pendingDate` from the API; if found, `reset()` the form with its data and update `attachedImages`; if not found, `reset()` with empty fields and `setAttachedImages([])`
- [ ] 3.9 Verify: changing date with no unsaved changes silently reloads the editor with the new date's data
- [ ] 3.10 Verify: changing date with unsaved changes shows the dialog; all three options behave correctly

## 4. Manual verification

- [ ] 4.1 Deploy to WIP stack and confirm date badge in viewer is clickable and navigates correctly
- [ ] 4.2 Confirm sidebar calendar icon navigates to an existing entry (viewer shown) and a new date (editor shown)
- [ ] 4.3 Confirm editor date field: no-unsaved-changes path reloads silently; dirty-form path shows dialog; all three dialog choices work correctly
- [ ] 4.4 Run E2E tests against WIP stack and confirm no regressions
