# Feature: Diary Entries (delta)

## MODIFIED Requirements

### Requirement: Date navigation
The viewer shows prev/next links to adjacent entries that have content, and the date badge is now an interactive control for jumping to any arbitrary date.

#### Scenario: Both previous and next entries exist
- **GIVEN** an entry for `2024-03-15` with a previous entry on `2024-03-10` and next entry on `2024-03-20`
- **THEN** the viewer shows a left arrow linking to `/diary/2024-03-10` and a right arrow linking to `/diary/2024-03-20`

#### Scenario: No previous entry
- **GIVEN** the current entry is the oldest in the diary
- **THEN** the left arrow is rendered but is non-interactive (greyed out)

#### Scenario: No next entry
- **GIVEN** the current entry is the newest in the diary
- **THEN** the right arrow is rendered but is non-interactive (greyed out)

#### Scenario: Navigation skips dates with no entries
- **GIVEN** entries exist on `2024-03-01` and `2024-03-15` but not on any date in between
- **WHEN** the user is viewing the `2024-03-01` entry
- **THEN** the next arrow links directly to `2024-03-15` (not to `2024-03-02`)

#### Scenario: Date badge is interactive
- **WHEN** the user views any entry
- **THEN** the date badge appears clickable (cursor-pointer, subtle hover ring)
- **AND** clicking it opens a date picker (see date-jump capability)

### Requirement: Edit an entry
The user can edit an existing entry or create a new one for a date. Changing the date field in the editor immediately navigates to that date's content rather than deferring to save time.

#### Scenario: Open editor from viewer
- **GIVEN** the user is viewing an existing entry
- **WHEN** they click the Edit button
- **THEN** the URL changes to `/diary/[date]?edit=true`
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
- **THEN** the editor closes, the URL returns to `/diary/[date]` (without `?edit=true`)
- **AND** the viewer re-fetches and shows the updated entry

#### Scenario: Cancel discards changes
- **WHEN** the user clicks Cancel in the editor
- **THEN** the URL returns to `/diary/[date]` without saving
- **AND** the entry is unchanged

#### Scenario: Edit state syncs with URL
- **GIVEN** the user is on `/diary/[date]?edit=true`
- **WHEN** they press the browser Back button
- **THEN** the URL changes to `/diary/[date]` and the editor closes

#### Scenario: Date field change with no unsaved changes — existing entry
- **GIVEN** the user has opened the editor for `2024-03-15` and has made no changes to the form
- **WHEN** the user changes the date field to `2024-03-20`, which has an existing entry
- **THEN** the editor immediately reloads with the title, body, tags, and attached images from `2024-03-20`
- **AND** the date field shows `2024-03-20`

#### Scenario: Date field change with no unsaved changes — empty date
- **GIVEN** the user has opened the editor for `2024-03-15` and has made no changes to the form
- **WHEN** the user changes the date field to `2024-03-22`, which has no entry
- **THEN** the editor clears the title, body, tags, and attached images
- **AND** the date field shows `2024-03-22`

#### Scenario: Date field change with unsaved changes — user saves to current date first
- **GIVEN** the user is editing `2024-03-15` and has made unsaved changes
- **WHEN** the user changes the date field to `2024-03-20`
- **THEN** a confirmation dialog is shown with the options "Save to 2024-03-15", "Move to 2024-03-20", and "Cancel"
- **WHEN** the user selects "Save to 2024-03-15"
- **THEN** the current draft is saved to `2024-03-15`
- **AND** the editor then loads the data for `2024-03-20` (same reload behavior as the no-unsaved-changes scenarios above)

#### Scenario: Date field change with unsaved changes — user moves draft to new date
- **GIVEN** the user is editing `2024-03-15` and has made unsaved changes
- **WHEN** the user changes the date field to `2024-03-20`
- **THEN** a confirmation dialog is shown
- **WHEN** the user selects "Move to 2024-03-20"
- **THEN** the current draft content (title, body, tags, attached images) is kept in the editor
- **AND** the date field switches to `2024-03-20` (the draft will be saved to `2024-03-20` on submit)
- **AND** the entry at `2024-03-15` is left unchanged

#### Scenario: Date field change with unsaved changes — user cancels
- **GIVEN** the user is editing `2024-03-15` and has made unsaved changes
- **WHEN** the user changes the date field to `2024-03-20` and the confirmation dialog appears
- **WHEN** the user selects "Cancel"
- **THEN** the date field reverts to `2024-03-15`
- **AND** the editor content is unchanged

### Requirement: Upsert semantics
Saving an entry fully replaces the entry for that date. The date field in the editor determines the save target and drives navigation between dates.

#### Scenario: Save overwrites existing entry
- **GIVEN** an entry for `2024-03-15` with title "Old Title" and body "Old body"
- **WHEN** the user edits it and saves with title "New Title" and body "New body"
- **THEN** the entry for `2024-03-15` now has title "New Title" and body "New body"
- **AND** there is still exactly one entry for `2024-03-15`

#### Scenario: Save targets the current date field value
- **GIVEN** the user navigated to `2024-03-20` by changing the date field (with no unsaved changes on `2024-03-15`)
- **WHEN** the user fills in a title and body and saves
- **THEN** the entry is upserted at `2024-03-20`
- **AND** the entry at `2024-03-15` (if any) is left unchanged
