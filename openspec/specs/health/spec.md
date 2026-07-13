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

#### Scenario: Core check types are always run
- **THEN** the following checks are always included:
  - `mime` — verifies that asset files match their declared MIME type
  - `orphans` — finds asset files not referenced in any diary entry
  - `refs` — finds diary entry bodies that reference asset files that no longer exist on disk

#### Scenario: Untagged check runs only when enabled
- **GIVEN** a family with `ai_tagging_backfill = true`
- **THEN** the `untagged` check is additionally run for that family
- **AND** for families with `ai_tagging_backfill = false`, the `untagged` check is skipped

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

### Requirement: Untagged days check
When a family has `ai_tagging_enabled` and `ai_tagging_backfill` enabled, the health subsystem SHALL run an `untagged` check. New AI analysis (model calls) SHALL occur only while `ai_tagging_backfill_done = false`, targeting pre-existing days that have not yet been analyzed. The check SHALL mark a day as analyzed even when it yields no suggestions, and SHALL NOT re-analyze a day merely because its content changed after a previous analysis. When no un-analyzed days remain, the family's backfill is complete (`ai_tagging_backfill_done = true`) and no further model calls are made for that family. Regardless of `ai_tagging_backfill_done`, the check SHALL continue to surface days that already have staged `pending_tags` as (non-fixable) review issues until the user resolves them — completion stops new analysis, not the review of results already produced.

#### Scenario: Backfill disabled means no untagged check
- **GIVEN** a family with `ai_tagging_backfill = false`
- **WHEN** health checks run
- **THEN** no `untagged` issues are produced for that family

#### Scenario: Completed backfill stops new analysis but keeps surfacing pending
- **GIVEN** a family with `ai_tagging_backfill_done = true`
- **AND** some days already have staged `pending_tags` from the backfill, and other un-tagged days have never been analyzed
- **WHEN** health checks run
- **THEN** no new model calls are made (the never-analyzed days are not analyzed)
- **AND** the days that already have staged `pending_tags` are still reported as non-fixable `untagged` review issues until resolved

#### Scenario: Un-analyzed day surfaced as an issue
- **GIVEN** a family with `ai_tagging_backfill = true` and `ai_tagging_enabled = true` and `ai_tagging_backfill_done = false`
- **AND** a pre-existing day with content but no confirmed tags that has not been analyzed
- **WHEN** health checks run
- **THEN** an issue with `check` = `untagged` is reported for that day

#### Scenario: A day yielding no suggestions is marked analyzed and not re-queried
- **GIVEN** the backfill analyzes a pre-existing day and the model returns no suggestions
- **THEN** the day is marked analyzed
- **AND** a later `untagged` check does not analyze that day again

#### Scenario: Editing a day does not resurface it in the untagged check
- **GIVEN** a day that has already been analyzed (or created after the backfill model took effect) and has no staged `pending_tags`
- **WHEN** the user edits its content and saves
- **THEN** the `untagged` check does not re-analyze it and does not report a new `untagged` issue for it
- **AND** if a day still has previously staged `pending_tags`, its existing review issue continues to be surfaced (the edit neither generates new pending nor clears the old)

#### Scenario: Confident auto-apply happens during the backfill (no manual fix)
- **GIVEN** a family with `ai_tagging_backfill = true` and `ai_tagging_auto = true` and threshold τ
- **AND** a pre-existing un-analyzed day with no confirmed tags for which the model returns a suggestion with confidence ≥ τ
- **WHEN** the `untagged` check runs
- **THEN** the confident tags are written to the day's confirmed `tags` immediately
- **AND** no issue is reported for that day (it is resolved)

#### Scenario: An un-analyzed day that already has confirmed tags is not auto-applied
- **GIVEN** a family with `ai_tagging_backfill = true` and `ai_tagging_auto = true`
- **AND** a pre-existing un-analyzed day that already has at least one confirmed tag
- **WHEN** the `untagged` check runs
- **THEN** suggestions are stored in `pending_tags` (never auto-applied over the user's tags)
- **AND** the issue is reported as not `fixable` (the user reviews chips on the day)

#### Scenario: Uncertain day routed to manual review
- **GIVEN** a family with `ai_tagging_backfill = true` and `ai_tagging_auto = true` and threshold τ
- **AND** a pre-existing un-analyzed day for which all suggestions have confidence < τ
- **WHEN** health checks run
- **THEN** the day's suggestions are stored in `pending_tags`
- **AND** the issue is reported as not `fixable`, identifying the day so the user can open and review it

#### Scenario: Default (non-auto) mode stages suggestions for manual review
- **GIVEN** a family with `ai_tagging_backfill = true` and `ai_tagging_auto = false`
- **AND** a pre-existing un-analyzed day
- **WHEN** health checks run
- **THEN** the day's suggestions are stored in `pending_tags`
- **AND** the issue is reported as not `fixable`, identifying the day so the user can open and accept the chips

#### Scenario: Untagged check does not repeatedly re-query the model
- **GIVEN** an untagged day whose suggestions were already staged in `pending_tags` by a prior run
- **AND** the day's content has not changed since
- **WHEN** the `untagged` check runs again
- **THEN** the staged suggestions are surfaced without calling the model again

### Requirement: Reviewing untagged days
The health UI SHALL present each unresolved `untagged` day as a link to that entry, where its staged suggestions can be reviewed and accepted. The `untagged` group SHALL NOT present a generic auto-fix button (suggestions are applied automatically under auto mode and reviewed per-entry otherwise).

#### Scenario: Untagged issue links to the entry
- **GIVEN** an `untagged` issue for the day `2024-03-15`
- **WHEN** the health panel renders it
- **THEN** it shows a link that opens `/diary/2024-03-15` in edit mode
- **AND** no "Fix" button is shown for the untagged group

