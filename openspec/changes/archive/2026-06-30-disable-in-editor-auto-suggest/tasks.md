## 1. Frontend: remove debounce

- [x] 1.1 Delete `SUGGEST_DEBOUNCE_MS` constant from `EntryEditor.tsx`
- [x] 1.2 Delete the debounced auto-suggest `useEffect` block (lines 117–125)
- [x] 1.3 Verify `lastSuggestedRef` is still needed (used only by the deleted effect) and remove it too if unused

## 2. Spec update

- [x] 2.1 Apply the delta spec: run `openspec archive` after implementation is complete to fold `specs/ai-tagging/spec.md` into the live spec

## 3. Verify

- [x] 3.1 Open an existing entry in the WIP stack, wait 10+ seconds — confirm no suggestion chips appear automatically
- [x] 3.2 Click ✨ "Suggest Tags" — confirm suggestions still appear correctly
- [x] 3.3 Open the editor for a new date, wait — confirm no automatic fetch
- [x] 3.4 Confirm pre-loaded `pendingTags` still surface as chips when opening an entry that has them
- [x] 3.5 Run E2E tests against the WIP stack: `cd /data/Diary/e2e && BASE_URL=$(jq -r '.deployments["diary-wip"].url' /data/data.json) npx playwright test --reporter=line`
