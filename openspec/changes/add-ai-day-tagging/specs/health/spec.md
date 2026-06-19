## ADDED Requirements

### Requirement: Untagged days check
When a family has `ai_tagging_backfill` enabled, the health subsystem SHALL run an `untagged` check that finds days which have no tags or whose tags are stale relative to their content, and surface them through the existing issues flow.

#### Scenario: Backfill disabled means no untagged check
- **GIVEN** a family with `ai_tagging_backfill = false`
- **WHEN** health checks run
- **THEN** no `untagged` issues are produced for that family

#### Scenario: Untagged day surfaced as an issue
- **GIVEN** a family with `ai_tagging_backfill = true` and `ai_tagging_enabled = true`
- **AND** a day with content but no confirmed tags
- **WHEN** health checks run
- **THEN** an issue with `check` = `untagged` is reported for that day

#### Scenario: Stale day surfaced as an issue
- **GIVEN** a family with `ai_tagging_backfill = true`
- **AND** a day whose current content hash differs from its stored `tags_source_hash`
- **WHEN** health checks run
- **THEN** an issue with `check` = `untagged` is reported for that day

#### Scenario: Confident auto-apply under auto mode
- **GIVEN** a family with `ai_tagging_backfill = true` and `ai_tagging_auto = true` and threshold τ
- **AND** an untagged day for which the model returns a suggestion with confidence ≥ τ
- **WHEN** the issue's fix is applied (via `POST /v1/health/fix`)
- **THEN** the confident tags are written to the day's confirmed `tags`
- **AND** the issue is reported as `fixable`

#### Scenario: Uncertain day routed to manual review
- **GIVEN** a family with `ai_tagging_backfill = true` and `ai_tagging_auto = true` and threshold τ
- **AND** an untagged day for which all suggestions have confidence < τ
- **WHEN** health checks run
- **THEN** the day's suggestions are stored in `pending_tags`
- **AND** the issue is reported as not `fixable` with a "solve manually" message

#### Scenario: Default (non-auto) mode stages suggestions for manual review
- **GIVEN** a family with `ai_tagging_backfill = true` and `ai_tagging_auto = false`
- **AND** an untagged day
- **WHEN** health checks run
- **THEN** the day's suggestions are stored in `pending_tags`
- **AND** the issue is reported as not `fixable` (the user accepts chips on the day itself)

## MODIFIED Requirements

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
