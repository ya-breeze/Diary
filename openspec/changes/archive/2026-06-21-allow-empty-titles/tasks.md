## 1. Frontend Validation

- [x] 1.1 In `next-frontend/src/components/diary/EntryEditor.tsx`, change `z.string().min(1, 'Title is required')` to `z.string()` in the `entrySchema` Zod object

## 2. E2E Test

- [x] 2.1 In `e2e/tests/entry.spec.ts`, add a test case: create an entry with an empty title field, save, and assert the viewer heading shows "Untitled"

## 3. Verification

- [x] 3.1 Run `make lint` (or `npx tsc --noEmit` in `next-frontend/`) to confirm no type errors
- [x] 3.2 Deploy to WIP stack and manually verify: open the editor, clear the title field, save — confirm no validation error and entry is saved
- [x] 3.3 Confirm the saved entry shows "Untitled" in the entry list and viewer
- [x] 3.4 Run E2E tests against the WIP stack and confirm no regressions
