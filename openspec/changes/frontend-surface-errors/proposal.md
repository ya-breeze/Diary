## Why

When a user action fails in the frontend (uploading images, saving an entry, requesting tag suggestions), the error is caught and only written to `console.error` — the user sees nothing. The progress indicator simply disappears and the action silently fails, leaving the user to believe the app is broken or unresponsive. A real report of "can't upload images and there's no feedback" traced directly to this pattern. There is currently no mechanism in the frontend to display an error to the user (only the login page has bespoke inline error text).

## What Changes

- Introduce a toast notification system: a `ToastProvider` mounted at the root layout and a `useToast()` hook, matching the KinCart approach (custom, zero new dependencies).
- Add a shared `getErrorMessage(error)` helper that normalizes the different error shapes thrown across the app (`ApiError` from `apiClient`, plain `Error`, and the raw XHR reject from `assetsApi.uploadAssetsBatch`) into a user-readable string.
- Replace `console.error`-only handling at every **user-initiated** action failure with a toast:
  - `EntryEditor`: image upload, save entry, suggest tags, accept tag, dismiss tag
  - `profile`: update AI tagging setting
  - `tags` page: currently a silent `catch {}`
  - `authStore`: logout
- **Background enhancement** fetches (the `aiEnabled` probe, the `knownTags` autocomplete load) and the 401 refresh/redirect path continue to degrade silently, keeping `console.error` for debugging — they are not user-initiated and should not raise toasts.
- Toasts show a message only (no retry action).

Not breaking: existing login inline-error behavior is preserved (it may optionally adopt the toast infra later, but is out of scope here).

## Capabilities

### New Capabilities
- `error-feedback`: When a user-initiated action fails, the app must surface a human-readable error to the user (via toast); background/optional enhancements degrade silently without user-facing errors.

### Modified Capabilities
<!-- None: no existing capability's behavioral requirements change; this adds a new cross-cutting UI capability. -->

## Impact

- **New code**: toast context/provider + hook, `getErrorMessage` helper, mount point in root layout.
- **Modified code**: `next-frontend/src/components/diary/EntryEditor.tsx`, `next-frontend/src/app/(dashboard)/profile/page.tsx`, `next-frontend/src/app/(dashboard)/tags/page.tsx`, `next-frontend/src/store/authStore.ts`.
- **Referenced (no behavior change)**: `next-frontend/src/lib/api/client.ts` (`ApiError`), `next-frontend/src/lib/api/assets.ts` (raw XHR error shape) — consumed by `getErrorMessage`.
- **Dependencies**: none added (custom toast implementation).
- **Tests**: frontend/E2E coverage for at least one failure path (e.g. upload failure shows a toast).
