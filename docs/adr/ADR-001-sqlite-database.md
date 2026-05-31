# ADR-001: Use SQLite as the Database

## Status
Accepted

## Context and Problem Statement

The Diary backend needs persistent storage for diary entries, user accounts, assets metadata, and change tracking. The storage solution must be simple to deploy and operate for a personal, self-hosted application used by a single family.

## Decision Drivers

- Zero operational overhead — no separate database process to manage
- File-based — easy to back up by copying a directory
- Single-host deployment — no need for network database access
- Low concurrency — one family, a handful of users

## Considered Options

- **SQLite** — embedded, file-based relational database
- **PostgreSQL** — full-featured server-based RDBMS
- **MySQL/MariaDB** — server-based RDBMS

## Decision Outcome

Chosen: **SQLite**, accessed via GORM ORM.

The application is a personal diary for a single family. There is no scenario requiring multiple concurrent writers at scale. SQLite's embedded nature eliminates a separate database process, and its single-file format integrates directly with the application's backup task (`diary-backup-YYYY-MM-DD.tar.gz`).

### Pros

- No separate process to start, monitor, or upgrade
- Database file is included in the tar.gz backup alongside assets
- GORM provides the same ORM interface as other drivers, so migration later is possible
- WAL mode handles the light concurrent read/write load

### Cons

- Not suitable for horizontal scaling or high concurrent write throughput
- Some SQL features (e.g., full ALTER TABLE) are limited
- `go-sqlite3` requires CGo, which complicates cross-compilation
