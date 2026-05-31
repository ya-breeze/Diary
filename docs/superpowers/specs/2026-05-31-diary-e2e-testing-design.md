# Diary E2E Testing Design

**Date:** 2026-05-31
**Status:** Approved

## Problem

Diary has backend API flow tests (Ginkgo) but zero frontend test coverage. There is no way to verify that UI flows work end-to-end before shipping a change.

## Approach

Two independent test layers with clear responsibilities:

- **Ginkgo flow tests** — own API contracts and error/edge cases. Fast, no infrastructure needed.
- **Playwright E2E** — own UI flows against a deployed WIP stack. Mirrors the pattern established in KinCart.

Both layers can be run independently. A combined `make test-all` runs them in sequence.

## Layer 1: Ginkgo flow tests (backend)

**Location:** `backend/test/flows/`

**How it works:** Each test file spins up a real in-process Go server on a random port with a temp SQLite database, fires HTTP requests via `TestAPIClient`, and asserts responses using Gomega matchers. Existing infrastructure (`SharedTestSetup`) is reused as-is.

**New files:**

| File | Coverage |
|---|---|
| `auth_errors_test.go` | Wrong password → 401; missing Authorization header → 401; malformed JWT → 401; valid token after password change still works |
| `item_edge_cases_test.go` | Fetch non-existent date → empty list; invalid date format → 400; PUT same date twice → second write overwrites first; empty title → 400 or accepted (match server behavior) |

**Run:** `make test` (already runs all Ginkgo specs)

## Layer 2: Playwright E2E (frontend)

**Location:** `e2e/` at repo root (new directory)

**Structure:**
```
e2e/
  playwright.config.ts
  package.json
  tsconfig.json
  tests/
    auth.spec.ts
    entry.spec.ts
    navigation.spec.ts
```

**Configuration:** `BASE_URL` env var, defaulting to `http://localhost:80`. Targets the WIP stack at `http://192.168.1.54:8885` in practice.

**Credentials:** Read from `data.json` (`test@test.com` / `test`). Hardcoded in test fixtures since they are seeded test credentials.

**Shared setup:** Each spec logs in via UI in `beforeEach` and waits for the diary page to be ready.

**Specs:**

`auth.spec.ts`
- Valid login lands on `/diary`
- Wrong password shows an error message (does not navigate away)
- Accessing `/diary` while unauthenticated redirects to `/login`

`entry.spec.ts`
- Write an entry for today, reload, verify text persists
- Edit an existing entry, reload, verify the update is saved
- The editor is empty for a date that has no entry

`navigation.spec.ts`
- Navigate directly to `/diary/2025-01-15`, verify the URL and date display
- Click "next day" arrow, verify URL increments by one day
- Click "previous day" arrow, verify URL decrements by one day

**Run:** `BASE_URL=http://192.168.1.54:8885 npx playwright test --reporter=line`

## Makefile targets

```makefile
.PHONY: test-e2e
test-e2e:
	@cd e2e && BASE_URL=$(BASE_URL) npx playwright test --reporter=line

.PHONY: test-all
test-all: test test-e2e
```

## What is NOT covered

- Asset/image upload (deferred — covered by backend flow tests at the API level)
- Search UI (deferred — search API is covered by backend; UI can be added later)
- Sync protocol edge cases (covered by existing `sync_integration_test.go`)
