## MODIFIED Requirements

### Requirement: Edit an entry
The user can edit an existing entry or create a new one for a date. Edit mode is reflected in the URL via `?edit=true`, but entering and leaving edit mode SHALL replace the current history entry rather than push a new one, so the browser/OS Back button does not accumulate edit-mode history entries.

#### Scenario: Open editor from viewer
- **WHEN** they click the Edit button
- **THEN** the URL changes to `/diary/[date]?edit=true` by replacing the current history entry (not pushing a new one)
- **AND** a full-screen editor overlay is shown pre-filled with the entry's title, date, tags, and body

#### Scenario: Title is required
- **WHEN** the user submits the editor form with an empty title
- **THEN** a validation error "Title is required" is shown
- **AND** the entry is not saved

#### Scenario: Tags are comma-separated
- **GIVEN** the user types `"happy, work,  outdoors "` in the tags field
- **WHEN** the entry is saved
- **THEN** the server stores `["happy", "work", "outdoors"]` (trimmed, empty values removed)

#### Scenario: Whitespace-only tags are stripped
- **GIVEN** the user submits tags `"happy, , ,work"`
- **WHEN** the entry is saved
- **THEN** the server stores `["happy", "work"]` (the empty segments are discarded)

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
