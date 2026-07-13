## ADDED Requirements

### Requirement: Automatic AI tagging is a one-time backfill only

Automatic (unattended) AI tag suggestion SHALL occur only via a one-time backfill of pre-existing entries. Creating a new entry SHALL NOT trigger any automatic AI suggestion, and editing an entry SHALL NOT trigger any automatic AI suggestion or re-analysis. The only interactive AI path SHALL be the explicit "suggest tags" action. Saving an entry (create or edit) SHALL NOT enqueue, stage, or apply any AI suggestion.

#### Scenario: Creating an entry does not auto-tag

- **GIVEN** a family with AI tagging enabled
- **WHEN** the user creates and saves a new entry
- **THEN** no AI suggestion is generated, staged, or applied for that entry as a result of the save
- **AND** the entry is not added to the untagged-days review as a result of the save

#### Scenario: Editing an entry does not re-tag

- **GIVEN** an existing entry (with or without confirmed tags)
- **WHEN** the user edits its title or body and saves
- **THEN** no AI suggestion is generated, staged, or applied as a result of the edit
- **AND** any existing confirmed and pending tags are left as they were

#### Scenario: Manual suggestion still works

- **GIVEN** an entry open in the editor with AI tagging enabled
- **WHEN** the user invokes the explicit "suggest tags" action
- **THEN** suggestions are returned as accept-able chips without modifying confirmed tags

### Requirement: Backfill runs once per family and then stops

The background backfill SHALL run for a family only while `ai_tagging_enabled` and `ai_tagging_backfill` are set and the family's `ai_tagging_backfill_done` flag is false. The backfill SHALL process pre-existing entries (those never analyzed) in bounded batches across successive runs until none remain. Each processed entry SHALL be marked as analyzed — including entries that yield no suggestions — so that no entry is analyzed more than once. When a backfill run finds no remaining un-analyzed entries (the corpus is exhausted, as opposed to stopping at the per-run batch cap), the system SHALL set `ai_tagging_backfill_done = true` and SHALL NOT run the backfill again for that family until it is re-triggered.

#### Scenario: An entry yielding no suggestions is not re-analyzed

- **GIVEN** the backfill analyzes a pre-existing entry and the model returns no suggestions
- **THEN** the entry is marked as analyzed
- **AND** subsequent backfill runs do not analyze that entry again

#### Scenario: Backfill completes and stops

- **GIVEN** a family whose pre-existing entries have all been analyzed by the backfill
- **WHEN** a backfill run finds no remaining un-analyzed entries
- **THEN** `ai_tagging_backfill_done` is set to true
- **AND** no further automatic AI analysis runs for that family

#### Scenario: Re-triggering a completed backfill

- **GIVEN** a family whose `ai_tagging_backfill_done` is true
- **WHEN** the family's `ai_tagging_backfill` setting is toggled off and then on
- **THEN** `ai_tagging_backfill_done` is reset to false
- **AND** a fresh one-time backfill pass runs over any still-un-analyzed entries

## MODIFIED Requirements

### Requirement: Confidence-based routing for unattended triggers
For the backfill (the only unattended trigger), the destination of suggestions SHALL be governed by `ai_tagging_auto`. Confident auto-apply SHALL only seed days that have **no confirmed tags**; once a day has any confirmed tag, the backfill SHALL only stage suggestions in `pending_tags` and never auto-apply. This prevents AI from re-adding a tag the user has curated or removed. Saving an entry (create or edit) is not an unattended trigger and SHALL NOT route or produce suggestions.

#### Scenario: Default mode stages suggestions
- **GIVEN** a family with `ai_tagging_auto = false`
- **WHEN** the backfill produces suggestions for a day
- **THEN** the suggestions are stored in that day's `pending_tags`
- **AND** the day is surfaced as a tagging health issue
- **AND** no confirmed tag is written

#### Scenario: Auto mode applies confident suggestions to an untagged day
- **GIVEN** a family with `ai_tagging_auto = true` and confidence threshold τ
- **AND** a day with no confirmed tags
- **WHEN** the backfill produces a suggestion with confidence ≥ τ
- **THEN** that tag is added to the day's confirmed `tags`

#### Scenario: Auto mode does not auto-apply to a day the user has already tagged
- **GIVEN** a family with `ai_tagging_auto = true` and confidence threshold τ
- **AND** a day with at least one confirmed tag (curated by the user)
- **WHEN** the backfill produces a suggestion with confidence ≥ τ
- **THEN** that suggestion is stored in `pending_tags` (not auto-applied)
- **AND** the user's existing confirmed tags are unchanged

#### Scenario: Auto mode routes uncertain suggestions to manual review
- **GIVEN** a family with `ai_tagging_auto = true` and confidence threshold τ
- **WHEN** the backfill produces only suggestions with confidence < τ
- **THEN** those suggestions are stored in `pending_tags`
- **AND** the day is surfaced as a "solve manually" tagging health issue

### Requirement: Per-family AI tagging configuration
The system SHALL expose per-family AI tagging settings, all of which default to off, and the explicit suggestion path SHALL require only `ai_tagging_enabled`. The system SHALL track per family whether the one-time backfill has completed via `ai_tagging_backfill_done` (default false); toggling `ai_tagging_backfill` from off to on SHALL reset `ai_tagging_backfill_done` to false. `ai_tagging_auto` SHALL govern only the backfill's routing (it has no effect on save/create/edit, which never trigger AI).

#### Scenario: Settings exposed and persisted
- **WHEN** a family views its settings
- **THEN** it can read and update `ai_tagging_enabled`, `ai_tagging_use_images`, `ai_tagging_use_video`, `ai_tagging_backfill`, and `ai_tagging_auto`

#### Scenario: Defaults are conservative
- **GIVEN** a family that has never changed AI settings
- **THEN** all `ai_tagging_*` settings are off
- **AND** no AI suggestion or media upload occurs until explicitly enabled

#### Scenario: Re-enabling backfill resets completion
- **GIVEN** a family whose `ai_tagging_backfill_done` is true
- **WHEN** `ai_tagging_backfill` is toggled from off to on
- **THEN** `ai_tagging_backfill_done` becomes false
