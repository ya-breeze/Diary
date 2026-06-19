# Feature: Automated Backup

## Purpose
How the server automatically creates, names, prunes, and atomically writes compressed backup archives of the database and assets.

## Requirements

### Requirement: Daily backup on a schedule
The server SHALL create a compressed backup archive automatically.

#### Scenario: First backup runs 30 seconds after startup
- **GIVEN** the server has just started
- **THEN** the backup task waits 30 seconds and then runs for the first time

#### Scenario: Backup runs on the configured interval
- **GIVEN** the server is running with `backup_interval` set to `"24h"`
- **THEN** after each run, the next backup fires 24 hours later

#### Scenario: Invalid backup_interval falls back to 24h
- **GIVEN** `backup_interval` is set to a non-parseable value (e.g. `"daily"`)
- **THEN** the server logs a warning and uses a 24-hour interval

### Requirement: Backup archive contents
Each backup SHALL be a single `.tar.gz` file containing the database and all assets.

#### Scenario: Archive is named by date
- **GIVEN** the backup runs on `2024-03-15`
- **THEN** the archive is named `diary-backup-2024-03-15.tar.gz` and stored in `<data_path>/backups/`

#### Scenario: Archive contains the database
- **THEN** the archive includes the SQLite database file produced via `VACUUM INTO` (a consistent, read-safe snapshot)

#### Scenario: Archive contains all asset files
- **THEN** the archive includes the entire `assets/` directory tree (all family subdirectories)

#### Scenario: Assets directory missing does not fail the backup
- **GIVEN** no assets have ever been uploaded (assets directory does not exist)
- **THEN** the archive is created successfully containing only the database

### Requirement: Skip if today's backup already exists
The backup task SHALL be idempotent within a single day.

#### Scenario: Backup skipped when already done today
- **GIVEN** `diary-backup-2024-03-15.tar.gz` already exists in the backups directory
- **WHEN** the task runs on `2024-03-15`
- **THEN** the task logs "today's backup already exists, skipping" and does not create a new archive

### Requirement: Atomic archive creation
Archives SHALL be written atomically to avoid corrupt partial files.

#### Scenario: Archive written to a temp file first
- **WHEN** a backup runs
- **THEN** the archive is written to `diary-backup-<date>.tar.gz.tmp`
- **AND** only renamed to the final name after the archive is fully written
- **AND** the `.tmp` file is cleaned up if the backup fails at any step

#### Scenario: Database snapshot is also temporary
- **WHEN** a backup runs
- **THEN** a temporary `<archive>.db.tmp` file is created for the VACUUM snapshot
- **AND** it is deleted after being added to the archive (regardless of success or failure)

### Requirement: Backup retention
Old backups SHALL be pruned to stay within a configured maximum count.

#### Scenario: Oldest backup pruned when limit exceeded
- **GIVEN** `backup_max_count` is `10` and 10 backups already exist
- **WHEN** a new backup is created (making 11)
- **THEN** the oldest backup (lexicographically first by filename, i.e. the earliest date) is deleted

#### Scenario: Default retention is 10 backups
- **GIVEN** `backup_max_count` is not configured or set to `0`
- **THEN** the server keeps at most 10 backups

#### Scenario: Multiple old backups pruned if needed
- **GIVEN** `backup_max_count` is `5` and 8 backups exist
- **WHEN** a new backup is created (making 9)
- **THEN** the 4 oldest backups are deleted, leaving exactly 5
