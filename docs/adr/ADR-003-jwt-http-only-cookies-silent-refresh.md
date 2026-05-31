# ADR-003: JWT Auth with HTTP-Only Cookies and Silent Refresh

## Status
Accepted

## Context and Problem Statement

The application must authenticate both the Next.js web frontend and a mobile sync client. The web frontend needs tokens that are safe from XSS, and the user experience must not require re-login when the access token expires.

## Decision Drivers

- Access tokens must not be accessible via JavaScript (XSS protection)
- Users should not be forced to re-login when a short-lived access token expires
- Multiple concurrent API calls returning 401 at the same time must not trigger multiple refresh requests
- Mobile sync client must also be able to authenticate (uses the same JWT mechanism)

## Considered Options

- **JWT in localStorage** — simple, but accessible to JavaScript (XSS risk)
- **JWT in HTTP-only cookies with manual re-login on expiry** — secure storage, but poor UX
- **JWT in HTTP-only cookies with silent refresh on 401** — secure storage, seamless UX
- **Server-side sessions** — no token management on the client, but requires session store

## Decision Outcome

Chosen: **JWT access token + refresh token in HTTP-only `SameSite=Strict` cookies, with silent refresh on 401.**

When the frontend receives a 401, it:
1. Queues all concurrent in-flight requests
2. Calls `POST /v1/refresh` once to obtain a new access token
3. Replays all queued requests with the new token

Refresh tokens are stored in the database and blacklisted on logout or rotation. A background goroutine periodically cleans up expired entries.

### Pros

- HTTP-only cookies prevent JavaScript access — immune to XSS token theft
- `SameSite=Strict` mitigates CSRF
- Concurrent 401 queuing avoids a thundering-herd of simultaneous refresh calls
- Transparent to the user — no visible re-login prompts on token expiry

### Cons

- Refresh token blacklist requires a database table and periodic cleanup
- Slightly more complex frontend fetch wrapper (`authStore` queuing logic)
- HTTP-only cookies require the API and frontend to share a domain or use CORS credentials
