# Feature: Health & Orphan Management

## Purpose
How the server detects and helps resolve data-integrity issues (MIME mismatches, orphaned assets, broken references) for a family's diary data.

## Requirements

### Requirement: Periodic health checks
The server SHALL run health checks automatically in the background.

#### Scenario: Health check runs on startup
- **GIVEN** the server has just started
- **THEN** the health checker waits 30 seconds and then runs all checks for all families

#### Scenario: Health check runs on schedule
- **GIVEN** the server is running
- **THEN** health checks run again on the configured interval (default: every 24 hours)

#### Scenario: Three check types are run
- **THEN** the following checks are always included:
  - `mime` — verifies that asset files match their declared MIME type
  - `orphans` — finds asset files not referenced in any diary entry
  - `refs` — finds diary entry bodies that reference asset files that no longer exist on disk

### Requirement: View health issues
Users SHALL be able to see the current health status of their data.

#### Scenario: No check has run yet
- **GIVEN** the server just started and the first check has not completed (< 30s)
- **WHEN** `GET /v1/health/issues` is called
- **THEN** an empty issues list is returned (not an error)

#### Scenario: Issues list returned after check runs
- **GIVEN** a health check has completed
- **WHEN** `GET /v1/health/issues` is called
- **THEN** the response includes:
  - `issues`: list of issues, each with `check` (type), `path`, `message`, and `fixable` flag
  - `lastChecked`: timestamp of the most recent check
  - `ignoredOrphans`: list of filenames the family has chosen to ignore

#### Scenario: Healthy data returns empty issues list
- **GIVEN** the family has no MIME mismatches, no orphaned files, and no broken refs
- **WHEN** `GET /v1/health/issues` is called after a check
- **THEN** `issues` is an empty array

### Requirement: Fix health issues
Users SHALL be able to trigger automatic fixes for fixable issues.

#### Scenario: Fix-all runs all fixable checks
- **WHEN** `POST /v1/health/fix` is called with an empty `checks` array
- **THEN** all three checks are run with fix mode enabled
- **AND** a second scan is run immediately after to capture the clean state
- **AND** the updated issues list is returned

#### Scenario: Fix a specific check
- **WHEN** `POST /v1/health/fix` is called with `checks: ["mime"]`
- **THEN** only the `mime` check is run with fix mode enabled
- **AND** non-mime issues from the previous scan are preserved in the response

### Requirement: Orphan management
The system SHALL allow individual orphaned files to be deleted, attached to an entry, or ignored.

#### Scenario: Delete an orphan
- **GIVEN** an orphaned file `photo.jpg` exists on disk
- **WHEN** `DELETE /v1/health/orphans/photo.jpg` is called
- **THEN** the file is removed from disk
- **AND** the orphan check is re-run and the response reflects the updated state

#### Scenario: Delete a non-existent orphan
- **WHEN** `DELETE /v1/health/orphans/missing.jpg` is called for a file that doesn't exist
- **THEN** the server returns 404

#### Scenario: Attach an orphan to an entry
- **GIVEN** an orphaned file `photo.jpg` exists
- **WHEN** `POST /v1/health/orphans/photo.jpg/attach` is called with `date: "2024-03-15"`
- **THEN** a markdown image reference `![photo.jpg](photo.jpg)` is appended to the body of the entry for `2024-03-15`
- **AND** the orphan check is re-run (the file is now referenced and no longer an orphan)

#### Scenario: Attach orphan creates entry if none exists
- **GIVEN** no entry exists for `2024-03-15`
- **WHEN** `POST /v1/health/orphans/photo.jpg/attach` is called with `date: "2024-03-15"`
- **THEN** a new empty entry for `2024-03-15` is created with the image reference in its body

#### Scenario: Ignore an orphan
- **WHEN** `POST /v1/health/orphans/photo.jpg/ignore` is called
- **THEN** `photo.jpg` is added to the family's ignored-orphans list
- **AND** the orphan check is re-run and `photo.jpg` no longer appears as an issue
- **AND** `photo.jpg` appears in `ignoredOrphans` in subsequent health responses

#### Scenario: Unignore an orphan
- **GIVEN** `photo.jpg` is on the ignored-orphans list
- **WHEN** `DELETE /v1/health/orphans/photo.jpg/ignore` is called
- **THEN** `photo.jpg` is removed from the ignored list
- **AND** the orphan check is re-run; `photo.jpg` may now appear as an issue again

#### Scenario: Invalid filename is rejected
- **WHEN** any orphan endpoint is called with a filename containing `/`, `\`, or `..`
- **THEN** the server returns 400

#### Scenario: Invalid date is rejected for attach
- **WHEN** `POST /v1/health/orphans/photo.jpg/attach` is called with a date not matching `YYYY-MM-DD`
- **THEN** the server returns 400

### Requirement: Orphan check re-runs after every action
Every orphan mutation SHALL trigger a fresh orphan scan so the returned state is always current.

#### Scenario: Issues after delete reflect current disk state
- **GIVEN** two orphans `a.jpg` and `b.jpg`
- **WHEN** `a.jpg` is deleted
- **THEN** the response includes only `b.jpg` as an orphan (not `a.jpg`)

#### Scenario: Non-orphan issues are preserved after orphan actions
- **GIVEN** there are both `mime` issues and `orphans` issues cached
- **WHEN** any orphan action is taken
- **THEN** the `mime` issues remain in the response (only orphan results are replaced)
