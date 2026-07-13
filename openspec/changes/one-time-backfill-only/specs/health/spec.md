## MODIFIED Requirements

### Requirement: Untagged days check
When a family has `ai_tagging_enabled` and `ai_tagging_backfill` enabled and its one-time backfill has not completed (`ai_tagging_backfill_done = false`), the health subsystem SHALL run an `untagged` check that finds pre-existing days that have not yet been analyzed and surfaces them through the existing issues flow. The check SHALL mark a day as analyzed even when it yields no suggestions, and SHALL NOT re-analyze a day merely because its content changed after a previous analysis. When no un-analyzed days remain, the family's backfill is complete and the check produces no further analysis for that family.

#### Scenario: Backfill disabled means no untagged check
- **GIVEN** a family with `ai_tagging_backfill = false`
- **WHEN** health checks run
- **THEN** no `untagged` issues are produced for that family

#### Scenario: Completed backfill means no untagged check
- **GIVEN** a family with `ai_tagging_backfill_done = true`
- **WHEN** health checks run
- **THEN** no new AI analysis runs and no new `untagged` issues are produced for that family

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
- **GIVEN** a day that has already been analyzed (or created after the backfill model took effect)
- **WHEN** the user edits its content and saves
- **THEN** the `untagged` check does not re-analyze it and does not report a new `untagged` issue for it

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
