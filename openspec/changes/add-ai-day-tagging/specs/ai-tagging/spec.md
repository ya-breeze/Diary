## ADDED Requirements

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
Suggestions SHALL be stored in a `pending_tags` field separate from the confirmed `tags`, and confirmed tags SHALL only change through explicit acceptance or, under auto mode, confident auto-apply.

#### Scenario: Suggestion populates pending tags
- **WHEN** suggestions are produced for an entry in the default (non-auto) mode
- **THEN** the suggested tag names are stored in the entry's `pending_tags`
- **AND** the entry's confirmed `tags` are unchanged

#### Scenario: Accepting a suggestion moves it to confirmed tags
- **GIVEN** an entry with `pending_tags` containing `"beach"`
- **WHEN** the user accepts the `"beach"` suggestion
- **THEN** `"beach"` is added to the confirmed `tags`
- **AND** `"beach"` is removed from `pending_tags`

#### Scenario: Already-confirmed tags are not re-suggested
- **GIVEN** an entry whose confirmed `tags` already include `"work"`
- **WHEN** suggestions are produced
- **THEN** `"work"` does not appear in `pending_tags`

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
For triggers where the user is not present (save-and-leave, backfill), the destination of suggestions SHALL be governed by `ai_tagging_auto`.

#### Scenario: Default mode stages suggestions
- **GIVEN** a family with `ai_tagging_auto = false`
- **WHEN** an unattended trigger produces suggestions for a day
- **THEN** the suggestions are stored in that day's `pending_tags`
- **AND** the day is surfaced as a tagging health issue
- **AND** no confirmed tag is written

#### Scenario: Auto mode applies confident suggestions
- **GIVEN** a family with `ai_tagging_auto = true` and confidence threshold τ
- **WHEN** an unattended trigger produces a suggestion with confidence ≥ τ
- **THEN** that tag is added to the day's confirmed `tags`

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
