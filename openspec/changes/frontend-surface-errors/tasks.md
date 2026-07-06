## 1. Toast infrastructure

- [x] 1.1 Create a `ToastProvider` client component + `useToast()` hook (context) that renders a fixed-position, dismissible, auto-timeout toast stack with at least an `error` variant, wrapped in an ARIA live region (`aria-live="assertive"` / `role="alert"`) so errors are announced to screen readers
- [x] 1.2 Mount `ToastProvider` in the root layout so it wraps the authenticated routes where user actions occur
- [x] 1.3 Add `getErrorMessage(error: unknown): string` normalizer handling `ApiError`, plain `Error`, and unknown thrown values (never throws); produce safe, user-friendly, status-aware messages and never surface raw backend response bodies, stack traces, or internal details

## 2. Wire toasts into user-initiated action failures

- [x] 2.1 `EntryEditor` image upload: on failure, `toast.error(getErrorMessage(e))` and clear the progress indicator
- [x] 2.2 `EntryEditor` save (`onSubmit`): on failure, toast and keep the user in the editor with content intact
- [x] 2.3 `EntryEditor` save-and-switch (`handleSaveAndSwitch`): on failure, toast and stay on the current date with content intact (no date switch)
- [x] 2.4 `EntryEditor` load-on-date-change (`reloadForDate`): on failure, toast
- [x] 2.5 `EntryEditor` suggest tags: on failure, toast
- [x] 2.6 `EntryEditor` accept tag: on failure, toast
- [x] 2.7 `EntryEditor` dismiss tag: on failure, toast
- [x] 2.8 `profile` page update AI tagging setting: on failure, toast
- [x] 2.9 `authStore` logout: on failure, toast

## 3. Preserve silent degradation

- [x] 3.1 Confirm `EntryEditor` `aiEnabled` probe still degrades silently (AI treated as off) and keeps `console.error`; no toast
- [x] 3.2 Confirm `EntryEditor` `knownTags` autocomplete load still degrades silently (empty list) and keeps `console.error`; no toast
- [x] 3.3 Confirm `profile` `getFamily` load and `authStore.validateSession` still degrade silently; no toast
- [x] 3.4 Confirm `client.ts` 401 refresh/redirect path raises no toast
- [x] 3.5 Leave the `tags` page inline error banner and `authStore.login` rethrow (login inline banner) as-is — verify they still surface errors and are not double-reported via a toast

## 4. Verification

- [x] 4.1 Run `make lint` (runs eslint in `next-frontend`) and `make build` (runs `next build`, which type-checks) and fix any issues
- [x] 4.2 Add/adjust a test covering at least one failure path (e.g. upload failure shows a toast)
- [ ] 4.3 Run E2E against the diary-wip stack for the affected flows; confirm a forced failure surfaces a toast and success paths are unaffected
