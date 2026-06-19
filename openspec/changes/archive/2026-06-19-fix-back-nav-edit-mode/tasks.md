## 1. Navigation changes

- [x] 1.1 In `next-frontend/src/app/(dashboard)/diary/[date]/page.tsx`, change `handleEdit` to use `router.replace` for `/diary/[date]?edit=true`
- [x] 1.2 In the same file, change `handleCloseEdit` (Cancel) to use `router.replace` for `/diary/[date]`
- [x] 1.3 In the same file, change the new-entry `onClose` to use `router.replace('/diary')`
- [x] 1.4 In `next-frontend/src/components/diary/EntryEditor.tsx`, change the post-save navigation in `onSubmit` to use `router.replace`

## 2. Tests

- [x] 2.1 Add an e2e test in `e2e/tests/entry.spec.ts` that builds a real history stack (viewer → Edit → Save) and asserts a single Back press does not re-enter the editor (would have caught the bug)

## 3. Verification

- [x] 3.1 Run `tsc --noEmit` in `next-frontend` and confirm no type errors
- [x] 3.2 Deploy the branch to the diary WIP stack and run the e2e suite against it (`BASE_URL=<wip-url> npx playwright test`)
- [x] 3.3 Confirm on mobile that one Back press returns to the list and a second leaves the diary
