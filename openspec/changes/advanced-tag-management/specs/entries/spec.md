## MODIFIED Requirements

### Requirement: Edit an entry
The user can edit an existing entry or create a new one for a date. Edit mode is reflected in the URL via `?edit=true`, but entering and leaving edit mode SHALL replace the current history entry rather than push a new one, so the browser/OS Back button does not accumulate edit-mode history entries. Tags SHALL be edited as a chip field: confirmed tags are shown as removable chips and new tags are added through an inline input.

#### Scenario: Open editor from viewer
- **WHEN** they click the Edit button
- **THEN** the URL changes to `/diary/[date]?edit=true` by replacing the current history entry (not pushing a new one)
- **AND** a full-screen editor overlay is shown pre-filled with the entry's title, date, tags, and body

#### Scenario: Confirmed tags render as removable chips
- **GIVEN** the entry being edited has tags `["happy","work","outdoors"]`
- **THEN** each tag is shown as a chip with an X (remove) control
- **AND** removing a chip drops only that tag, leaving the others intact

#### Scenario: Adding a tag via the inline input
- **GIVEN** the user types `outdoors` in the inline add-tag input and commits it (Enter or comma)
- **THEN** `outdoors` is added as a chip
- **AND** the inline input is cleared, ready for the next tag

#### Scenario: Tags are trimmed and de-duplicated on save
- **GIVEN** the user has chips `happy`, `work`, `outdoors` (a leading/trailing space on an added tag is trimmed when committed)
- **WHEN** the entry is saved
- **THEN** the server stores `["happy", "work", "outdoors"]` (trimmed, empty values removed, no duplicates)

#### Scenario: Whitespace-only input is not added as a tag
- **GIVEN** the user commits a blank or whitespace-only value in the inline input
- **THEN** no chip is added and the entry's tags are unchanged

#### Scenario: Blank tags are filtered from API responses
- **GIVEN** an entry whose stored `tags` or `pending_tags` contain empty or whitespace-only strings (legacy data written before the save-path filter existed)
- **WHEN** the entry is fetched via the API
- **THEN** the response omits those blank entries from both `tags` and `pending_tags`
- **AND** valid (non-blank) tags are returned unchanged

#### Scenario: Body is optional
- **WHEN** the user saves an entry with a title but no body text
- **THEN** the entry is saved with an empty body (no error)

#### Scenario: Save closes editor and refreshes data
- **WHEN** the user saves successfully
- **THEN** the editor closes and the URL returns to `/diary/[date]` (without `?edit=true`) by replacing the current history entry
- **AND** the viewer re-fetches and shows the updated entry

#### Scenario: Cancel discards changes
- **WHEN** the user clicks Cancel in the editor
- **THEN** the URL returns to `/diary/[date]` (without `?edit=true`) by replacing the current history entry, without saving
- **AND** the entry is unchanged

#### Scenario: Back button does not stack edit-mode history
- **GIVEN** the user opened the app, navigated to `/diary/[date]`, entered edit mode, and saved or cancelled
- **WHEN** they press the browser/OS Back button
- **THEN** they return to the previous real page (the entry list), not back into the editor or a duplicate viewer
- **AND** a single further Back press leaves the diary
