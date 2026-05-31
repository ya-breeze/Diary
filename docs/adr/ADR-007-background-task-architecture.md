# ADR-007: Background Task Architecture (In-Process Goroutines)

## Status
Accepted

## Context and Problem Statement

The application requires two recurring background operations: a health check that detects asset/database inconsistencies, and an automated backup that archives data daily. These tasks need access to the storage layer and must run on a configurable interval without external coordination.

## Decision Drivers

- No external scheduler or job queue infrastructure should be required
- Tasks need access to the same storage interface as the HTTP handlers
- Intervals must be configurable (not hard-coded at deploy time)
- Tasks must stop cleanly when the server shuts down

## Considered Options

- **External cron job** — host-level cron calls the binary or an API endpoint; simple but couples the task schedule to the host OS
- **In-process goroutines** — tasks started at `Serve()` time, cancelled via `context.Context`
- **Separate worker process** — task runner as a distinct binary; clean separation but requires orchestration

## Decision Outcome

Chosen: **In-process goroutines**, one per task, started in `server.Serve()` and cancelled when the server's root `context.Context` is cancelled.

`CheckerTask` runs health checks (`MimeCheck`, `OrphansCheck`, `RefsCheck`) on a configurable interval (`DIARY_HEALTH_CHECK_INTERVAL`, default `24h`). Results are stored in memory and served via `GET /v1/health`.

`BackupTask` creates a `diary-backup-YYYY-MM-DD.tar.gz` of the data directory on a configurable interval (`DIARY_BACKUP_INTERVAL`, default `24h`), retaining at most `DIARY_BACKUP_MAX_COUNT` archives.

### Pros

- Zero infrastructure — no external cron daemon, queue, or separate binary
- Direct access to the `database.Storage` interface without serialization
- Graceful shutdown via context cancellation propagates automatically
- Intervals are runtime-configurable via environment variables

### Cons

- Tasks stop if the server crashes — there is no persistence of "last run" time across restarts
- A long-running check or backup blocks its goroutine; no worker pool or timeout per task
- Not suitable if tasks need to run on a distributed schedule across multiple instances
