# ADR-006: Filesystem Storage for Assets

## Status
Accepted

## Context and Problem Statement

Users can upload images, videos, and other files that are referenced from diary entries as markdown links. These binary assets need a storage backend that is simple to operate and consistent with the rest of the self-hosted deployment model.

## Decision Drivers

- Assets can be large (up to 200 MB per file by default)
- Storage must survive container recreates (persistent volume)
- Backup must be achievable with the same tar.gz mechanism used for the database
- No cloud infrastructure — the app is entirely self-hosted

## Considered Options

- **Database BLOBs** — store files as binary columns in SQLite
- **Filesystem** — store files under `$DIARY_DATAPATH/assets/`
- **Object storage (S3-compatible)** — MinIO or cloud S3

## Decision Outcome

Chosen: **Filesystem storage** under the configured `GB_DATAPATH` directory, alongside the SQLite database file.

Files are written to `<datapath>/assets/<filename>` by the upload handler. The HTTP server serves them directly. The backup task (`BackupTask`) archives the entire data directory — database and assets together — into a single `diary-backup-YYYY-MM-DD.tar.gz`.

A health check (`OrphansCheck`) detects files present on disk but not referenced by any diary entry, and vice versa.

### Pros

- No DB bloat — SQLite remains fast even with many large uploads
- Backup includes assets automatically (single tar.gz covers everything)
- Files can be served by the Go HTTP handler or a reverse proxy directly
- Simple to inspect and manage on the host

### Cons

- No transactional consistency between the database and the filesystem — a crash mid-upload can leave orphan files
- The `OrphansCheck` health task is needed to detect drift between DB references and disk files
- Not suitable if the app were ever distributed across multiple hosts
