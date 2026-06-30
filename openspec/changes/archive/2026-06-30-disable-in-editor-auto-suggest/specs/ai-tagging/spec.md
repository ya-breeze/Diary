## REMOVED Requirements

### Requirement: In-editor debounced auto-suggest
**Reason**: The debounce fires on mount (before the user types) and users find the automatic fetch distracting. The explicit "suggest tags" button provides sufficient control.
**Migration**: Use the ✨ "Suggest Tags" button in the editor to request suggestions.

## MODIFIED Requirements

### Requirement: AI availability and graceful degradation
The system SHALL provide AI tag suggestion only when a Gemini API key is configured, and SHALL otherwise behave exactly as without the feature.

#### Scenario: API key not configured
- **GIVEN** the server has no `GEMINI_API_KEY` set
- **WHEN** any AI tagging trigger fires (button, save, or backfill)
- **THEN** no AI call is made
- **AND** no error is surfaced to the user for normal entry operations
- **AND** the "suggest tags" action reports that AI tagging is unavailable

#### Scenario: Family has AI tagging disabled
- **GIVEN** a family with `ai_tagging_enabled = false`
- **WHEN** any AI tagging trigger fires for that family
- **THEN** no suggestion is produced and no AI call is made

### Requirement: Pending suggestions never overwrite confirmed tags
Every suggestion call SHALL return its suggestions in the API response. Suggestions persisted on an entry SHALL live in a `pending_tags` field separate from the confirmed `tags`, and confirmed tags SHALL only change through explicit acceptance or, under auto mode, confident auto-apply.

#### Scenario: Attended suggestion returns chips without persisting
- **GIVEN** a brand-new, unsaved entry being edited
- **WHEN** the explicit "suggest tags" action produces suggestions
- **THEN** the suggestions are returned in the response and shown as chips
- **AND** nothing is written to storage (no `pending_tags` row exists yet for an unsaved entry)

#### Scenario: Unattended suggestion persists to pending tags
- **WHEN** an unattended trigger (save-and-leave or backfill) produces suggestions for an existing entry in the default (non-auto) mode
- **THEN** the suggested tag names are stored in the entry's `pending_tags`
- **AND** the entry's confirmed `tags` are unchanged

#### Scenario: Reopening an entry shows previously persisted suggestions
- **GIVEN** an entry whose `pending_tags` were populated by an earlier unattended trigger
- **WHEN** the user opens that entry
- **THEN** the persisted `pending_tags` are shown as reviewable chips
- **AND** they are shown regardless of whether AI tagging is currently enabled (they were already generated)

#### Scenario: Accepting a suggestion moves it to confirmed tags
- **GIVEN** an existing entry with confirmed `tags` `["work"]` and `pending_tags` containing `"beach"`
- **WHEN** the user accepts the `"beach"` suggestion
- **THEN** `"beach"` is added to the confirmed `tags` additively (`"work"` is preserved)
- **AND** `"beach"` is removed from `pending_tags`
- **AND** the change is persisted immediately (it sticks even if the entry is not separately saved)

#### Scenario: Already-confirmed tags are not re-suggested
- **GIVEN** an entry whose confirmed `tags` already include `"work"`
- **WHEN** suggestions are produced
- **THEN** `"work"` does not appear in `pending_tags`

### Requirement: Per-family AI tagging configuration
The system SHALL expose per-family AI tagging settings, all of which default to off, and the explicit suggestion path SHALL require only `ai_tagging_enabled`.

#### Scenario: Settings exposed and persisted
- **WHEN** a family views its settings
- **THEN** it can read and update `ai_tagging_enabled`, `ai_tagging_use_images`, `ai_tagging_use_video`, `ai_tagging_backfill`, and `ai_tagging_auto`

#### Scenario: Defaults are conservative
- **GIVEN** a family that has never changed AI settings
- **THEN** all `ai_tagging_*` settings are off
- **AND** no AI suggestion or media upload occurs until explicitly enabled
