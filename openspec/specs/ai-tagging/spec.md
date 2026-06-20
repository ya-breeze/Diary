# ai-tagging Specification

## Purpose
TBD - created by archiving change add-ai-day-tagging. Update Purpose after archive.
## Requirements
### Requirement: AI availability and graceful degradation
The system SHALL provide AI tag suggestion only when a Gemini API key is configured, and SHALL otherwise behave exactly as without the feature.

#### Scenario: API key not configured
- **GIVEN** the server has no `GEMINI_API_KEY` set
- **WHEN** any AI tagging trigger fires (button, debounce, save, or backfill)
- **THEN** no AI call is made
- **AND** no error is surfaced to the user for normal entry operations
- **AND** the "suggest tags" action reports that AI tagging is unavailable

#### Scenario: Family has AI tagging disabled
- **GIVEN** a family with `ai_tagging_enabled = false`
- **WHEN** any AI tagging trigger fires for that family
- **THEN** no suggestion is produced and no AI call is made

### Requirement: Tag suggestion from entry text
When AI tagging is enabled, the system SHALL produce tag suggestions for a day from its title and body, each suggestion carrying a confidence value between 0 and 1.

#### Scenario: Suggest tags for a text entry
- **GIVEN** a family with `ai_tagging_enabled = true` and a configured API key
- **WHEN** suggestions are requested for an entry with a non-empty title or body
- **THEN** the system returns a list of `{name, confidence}` suggestions
- **AND** confidence values are between 0 and 1 inclusive

#### Scenario: Entry with no text yields no suggestions
- **GIVEN** an entry with an empty title and empty body
- **WHEN** suggestions are requested
- **THEN** an empty suggestion list is returned without calling the model

### Requirement: Hybrid vocabulary
The system SHALL bias suggestions toward the family's existing tags and SHALL limit how many new tags a single suggestion call may introduce.

#### Scenario: Existing tags are preferred
- **GIVEN** a family whose entries already use the tags `["travel", "family", "work"]`
- **WHEN** suggestions are produced for an entry about a family trip
- **THEN** the model is given the existing tag set as context with instruction to prefer it
- **AND** matching concepts reuse existing tags (e.g. `"travel"`) rather than coining near-duplicates (e.g. `"traveling"`)

#### Scenario: New tags are capped
- **WHEN** a single suggestion call is made
- **THEN** at most approximately two tags not already in the family's existing set are introduced

### Requirement: Pending suggestions never overwrite confirmed tags
Every suggestion call SHALL return its suggestions in the API response. Suggestions persisted on an entry SHALL live in a `pending_tags` field separate from the confirmed `tags`, and confirmed tags SHALL only change through explicit acceptance or, under auto mode, confident auto-apply.

#### Scenario: Attended suggestion returns chips without persisting
- **GIVEN** a brand-new, unsaved entry being edited
- **WHEN** the explicit "suggest tags" action or the in-editor debounce produces suggestions
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

### Requirement: Dismiss a pending suggestion
A user SHALL be able to dismiss a single pending suggestion, removing it from the entry's `pending_tags` without adding it to the confirmed `tags`.

#### Scenario: Dismissing a suggestion removes it from pending
- **GIVEN** an entry with `pending_tags` containing `"hiking"` and `"mountains"`
- **WHEN** the user dismisses the `"hiking"` suggestion
- **THEN** `"hiking"` is removed from `pending_tags`
- **AND** `"mountains"` remains in `pending_tags`
- **AND** the confirmed `tags` are unchanged

#### Scenario: Dismissing the last suggestion clears the entry from review
- **GIVEN** an entry whose only pending suggestion is `"hiking"` and which appears in the backfill review list
- **WHEN** the user dismisses `"hiking"`
- **THEN** `pending_tags` becomes empty
- **AND** the entry no longer appears as an untagged review item

### Requirement: AI never removes a confirmed tag
AI tagging SHALL only ever add to the confirmed `tags` list; it SHALL NOT delete, rename, or replace any confirmed tag. Removal of a confirmed tag SHALL only occur through explicit user action.

#### Scenario: Retag preserves a user-added tag not suggested by AI
- **GIVEN** an entry whose confirmed `tags` include `"anniversary"` (added manually, never suggested by AI)
- **WHEN** the entry's text changes and the day is retagged (in any mode)
- **THEN** `"anniversary"` remains in the confirmed `tags`

#### Scenario: Retag never replaces the confirmed tag list wholesale
- **GIVEN** an entry with confirmed `tags` `["work", "anniversary"]`
- **WHEN** the day is retagged and the model returns a completely different set of suggestions
- **THEN** the confirmed `tags` still contain `"work"` and `"anniversary"`
- **AND** any new tags are only added (auto mode, confident) or staged in `pending_tags` — never substituted for the existing list

### Requirement: Pending and confirmed tags are kept disjoint
The system SHALL ensure a tag never appears in both `pending_tags` and confirmed `tags` at the same time. Whenever confirmed tags change (user edit, acceptance, or auto-apply) or a retag runs, any pending entry already present in confirmed `tags` SHALL be pruned.

#### Scenario: User manually adds a tag the AI had pending
- **GIVEN** an entry with `pending_tags` containing `"beach"` and no confirmed `"beach"` tag
- **WHEN** the user manually adds `"beach"` to the confirmed `tags` and saves
- **THEN** `"beach"` is present in confirmed `tags`
- **AND** `"beach"` is removed from `pending_tags` (no duplicate across the two lists)

#### Scenario: Retag prunes pending entries now confirmed
- **GIVEN** an entry whose confirmed `tags` include `"family"`
- **WHEN** a retag runs and would otherwise stage `"family"` as a suggestion
- **THEN** `"family"` is not added to `pending_tags`

### Requirement: Explicit suggestion trigger
The system SHALL expose an explicit "suggest tags" action for a day that returns suggestions without modifying confirmed tags.

#### Scenario: User requests suggestions from the editor
- **GIVEN** an entry open in the editor
- **WHEN** the user invokes "suggest tags"
- **THEN** suggestions are returned and shown as accept-able chips
- **AND** no confirmed tag is written until the user accepts a chip

### Requirement: In-editor debounced auto-suggest
While an entry is being edited, the system SHALL fetch suggestions automatically after a short period of inactivity, and this attended path SHALL only suggest (never auto-apply) regardless of the auto-tag setting.

#### Scenario: Suggestions appear after the user stops typing
- **GIVEN** the user is editing an entry and has changed its content
- **WHEN** roughly 4 seconds elapse with no further changes
- **THEN** suggestions are fetched and shown as chips
- **AND** no confirmed tag is written automatically

#### Scenario: No re-suggest without a content change
- **GIVEN** suggestions were just fetched for the current content
- **WHEN** the debounce interval elapses again with no content change since the last fetch
- **THEN** no new suggestion call is made

#### Scenario: Auto mode does not change in-editor behavior
- **GIVEN** a family with `ai_tagging_auto = true`
- **WHEN** the in-editor debounce fires
- **THEN** suggestions are still only shown as chips and are not auto-applied

### Requirement: Confidence-based routing for unattended triggers
For triggers where the user is not present (save-and-leave, backfill), the destination of suggestions SHALL be governed by `ai_tagging_auto`. Confident auto-apply SHALL only seed days that have **no confirmed tags**; once a day has any confirmed tag, unattended triggers SHALL only stage suggestions in `pending_tags` and never auto-apply. This prevents AI from re-adding a tag the user has curated or removed.

#### Scenario: Default mode stages suggestions
- **GIVEN** a family with `ai_tagging_auto = false`
- **WHEN** an unattended trigger produces suggestions for a day
- **THEN** the suggestions are stored in that day's `pending_tags`
- **AND** the day is surfaced as a tagging health issue
- **AND** no confirmed tag is written

#### Scenario: Auto mode applies confident suggestions to an untagged day
- **GIVEN** a family with `ai_tagging_auto = true` and confidence threshold τ
- **AND** a day with no confirmed tags
- **WHEN** an unattended trigger produces a suggestion with confidence ≥ τ
- **THEN** that tag is added to the day's confirmed `tags`

#### Scenario: Auto mode does not auto-apply to a day the user has already tagged
- **GIVEN** a family with `ai_tagging_auto = true` and confidence threshold τ
- **AND** a day with at least one confirmed tag (curated by the user)
- **WHEN** an unattended trigger produces a suggestion with confidence ≥ τ
- **THEN** that suggestion is stored in `pending_tags` (not auto-applied)
- **AND** the user's existing confirmed tags are unchanged

#### Scenario: Removed tag is not re-applied after subsequent edits
- **GIVEN** a family with `ai_tagging_auto = true`
- **AND** a day where the user deleted an AI-applied tag, leaving at least one other confirmed tag
- **WHEN** the day's text changes and it is retagged
- **THEN** the previously removed tag is not auto-applied again (the day is no longer untagged, so auto-apply does not run)

#### Scenario: Auto mode routes uncertain suggestions to manual review
- **GIVEN** a family with `ai_tagging_auto = true` and confidence threshold τ
- **WHEN** an unattended trigger produces only suggestions with confidence < τ
- **THEN** those suggestions are stored in `pending_tags`
- **AND** the day is surfaced as a "solve manually" tagging health issue

### Requirement: Image-based suggestions
When `ai_tagging_use_images` is enabled, the system SHALL include the entry's referenced image assets in the suggestion request.

#### Scenario: Images included when enabled
- **GIVEN** a family with `ai_tagging_use_images = true`
- **AND** an entry whose body references image assets
- **WHEN** suggestions are produced
- **THEN** the referenced images are sent alongside the text to the model

#### Scenario: Images excluded when disabled
- **GIVEN** a family with `ai_tagging_use_images = false`
- **WHEN** suggestions are produced
- **THEN** only the entry text is sent to the model

### Requirement: Video-based suggestions via keyframes
When `ai_tagging_use_video` is enabled, the system SHALL sample a small number of keyframes from each referenced video and include them as images in the suggestion request; native video upload SHALL NOT be used.

#### Scenario: Video keyframes included when enabled
- **GIVEN** a family with `ai_tagging_use_video = true`
- **AND** an entry whose body references a video asset
- **WHEN** suggestions are produced
- **THEN** a small number (approximately 3–5) of keyframes are extracted from the video
- **AND** those keyframes are sent as images to the model

#### Scenario: Frame extraction unavailable degrades gracefully
- **GIVEN** `ai_tagging_use_video = true` but keyframe extraction cannot run
- **WHEN** suggestions are produced for an entry with a video
- **THEN** suggestions are still produced from the available text and images
- **AND** no error blocks the suggestion

### Requirement: Per-family AI tagging configuration
The system SHALL expose per-family AI tagging settings, all of which default to off, and the in-editor and explicit suggestion paths SHALL require only `ai_tagging_enabled`.

#### Scenario: Settings exposed and persisted
- **WHEN** a family views its settings
- **THEN** it can read and update `ai_tagging_enabled`, `ai_tagging_use_images`, `ai_tagging_use_video`, `ai_tagging_backfill`, and `ai_tagging_auto`

#### Scenario: Defaults are conservative
- **GIVEN** a family that has never changed AI settings
- **THEN** all `ai_tagging_*` settings are off
- **AND** no AI suggestion or media upload occurs until explicitly enabled

