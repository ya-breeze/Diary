## Context

The Next.js frontend catches errors at many call sites and only calls `console.error`, so users get no feedback when actions fail. There is no notification infrastructure; the login page is the only place with visible error text, using local component state (`submitError`) plus an inline red banner.

Errors arrive in inconsistent shapes:
- `apiClient<T>()` (`src/lib/api/client.ts`) throws `ApiError(status, message)` for non-OK responses.
- `assetsApi.uploadAssetsBatch` (`src/lib/api/assets.ts`) uses a raw `XMLHttpRequest` and rejects with a plain `Error("Batch upload failed: <status>")` / `Error("Upload network error")`.
- Thrown values in `catch` are typed `unknown`.

Constraints: match the KinCart pattern (a custom toast context with `useToast` + an error normalizer) rather than introducing a dependency; keep the 401 refresh/redirect flow untouched.

## Goals / Non-Goals

**Goals:**
- Every user-initiated action failure shows a readable message to the user.
- One shared normalizer converts any error shape into a readable string.
- Background/optional fetches keep degrading silently (still `console.error`).
- No new runtime dependency.

**Non-Goals:**
- Retry affordances in toasts (message only).
- Migrating the login page off its inline banner.
- Changing backend behavior or the error shapes that `apiClient`/`assetsApi` throw.
- A general logging/telemetry system.

## Decisions

### Decision: Custom toast context over a library (e.g. `sonner`)
A small `ToastProvider` + `useToast()` in React context, mounted once in the root layout, rendering a fixed-position stack of dismissible toasts with auto-timeout.
- **Why:** Matches KinCart (`useToast` + `getApiError` from a `ToastContext`), keeps the two projects consistent, and adds zero dependencies for what is a small surface.
- **Alternative considered:** `sonner` / `react-hot-toast` — less code to write but a new dependency and a different API than the sibling project. Rejected for consistency + minimalism.

### Decision: A single `getErrorMessage(error: unknown): string` helper
Lives alongside the API layer (e.g. `src/lib/api/`). Handles `ApiError` (use its message/status), plain `Error` (use `.message`), and anything else (generic fallback). Never throws.
- **Why:** Call sites throw/reject different shapes; one normalizer keeps messages consistent and call sites to one line: `toast.error(getErrorMessage(e))`.
- **Alternative considered:** Normalizing every thrower to `ApiError`. Larger blast radius (touches the XHR path and its own error semantics) for no user-visible gain now.

### Decision: "User-initiated vs background" is the toggle for surfacing
User-initiated failures toast; background enhancement fetches and the 401 path stay silent (still logged).
- **Why:** Literal "toast everything" spams users for fetches they never triggered (aiEnabled probe, knownTags load) on flaky networks, training them to ignore toasts. Surfacing what the user asked for is the actual intent.
- **Boundary (from current code):**
  - Toast: `EntryEditor` upload / save (`onSubmit`) / suggest / accept / dismiss; `profile` AI-setting update; `tags` page action; `authStore` logout.
  - Silent (keep `console.error`): `EntryEditor` `aiEnabled` probe and `knownTags` load; `client.ts` 401 refresh/redirect.

### Decision: Toast placement and lifecycle
Toasts render top-level (portal/fixed container in the provider), auto-dismiss after a few seconds, are manually dismissible, and stack. An `error` variant is required now; a `success`/`info` variant is optional and cheap to include but not required by any spec scenario.

## Risks / Trade-offs

- **Judgment calls on the user/background boundary** → The boundary is enumerated explicitly above and encoded in the spec so reviewers can check each site.
- **`getErrorMessage` could leak raw backend text** (`ApiError` carries the response body) → Prefer status-aware, friendly phrasing; fall back to the raw message only when nothing better exists. Verified against real messages during implementation.
- **Toast provider placement in the App Router** (server vs client boundary) → The provider is a client component mounted inside the root layout; verify it wraps the authenticated routes where these actions occur.
- **Duplicate/toast spam if an action retries internally** → Toast at the outermost user-action catch only, not inside inner helpers.

## Migration Plan

Additive, frontend-only. No data or API migration. Ships behind normal deploy; rollback is reverting the change. No feature flag needed.

## Open Questions

- Include a non-error (`success`/`info`) toast variant now, or add it when first needed? (Leaning: include the variant type but only use `error` in this change.)
