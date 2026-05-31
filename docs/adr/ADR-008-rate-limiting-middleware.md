# ADR-008: Rate Limiting in Gorilla Mux Middleware

## Status
Accepted

## Context and Problem Statement

The API is exposed to the internet. Without rate limiting, endpoints such as `POST /v1/authorize` are vulnerable to brute-force attacks. The rate limiter must be easy to disable in test environments.

## Decision Drivers

- Protect the login endpoint from credential stuffing and brute force
- Must be bypassable in automated tests (`DIARY_DISABLERATELIMIT=true`)
- No external dependency (Redis, etc.) — rate limit state can be in-memory for a single-host deployment

## Considered Options

- **No rate limiting** — simplest, but leaves auth endpoints exposed
- **Reverse proxy rate limiting (nginx)** — offloads to nginx; works but couples rate-limit config to the deployment layer, not the application
- **Application-level middleware** — IP-keyed in-memory limiter as a Gorilla Mux `MiddlewareFunc`

## Decision Outcome

Chosen: **IP-based in-memory rate limiter as a Gorilla Mux middleware**, applied globally to all routes.

`RateLimitMiddleware` maintains a per-IP token bucket (via `RateLimiterStore`). Requests that exceed the limit receive `429 Too Many Requests`. Setting `DIARY_DISABLERATELIMIT=true` makes the middleware a no-op, which E2E tests rely on to avoid flakiness.

### Pros

- Rate limiting is part of the application binary — no nginx config changes required
- Disabling is a single environment variable, clean for CI and local testing
- No external state store — zero additional infrastructure

### Cons

- In-memory state is lost on server restart — a restarted server resets all rate limit counters
- Ineffective if all clients share one egress IP (e.g., behind a corporate NAT)
- Single-host only — does not coordinate across multiple instances
