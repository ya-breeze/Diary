# Feature: Diary Entries

## Purpose
How diary entries are viewed, navigated, tagged, edited, and saved — one entry per date per family with upsert semantics.
## Requirements
### Requirement: View an entry by date
Navigating to `/diary/[date]` SHALL show the entry for that date.

#### Scenario: Entry exists for the date
- **GIVEN** an entry exists for `2024-03-15`
- **WHEN** the user navigates to `/diary/2024-03-15`
- **THEN** the entry's title, body (rendered as markdown), and tags are displayed

#### Scenario: No entry exists for the date
- **GIVEN** no entry exists for `2024-03-15`
- **WHEN** the user navigates to `/diary/2024-03-15`
- **THEN** the editor is shown immediately (not the viewer) so the user can create a new entry
- **AND** the date field is pre-filled with `2024-03-15`

#### Scenario: Entry data is family-scoped
- **GIVEN** two users share a family
- **WHEN** either user navigates to the same date
- **THEN** they see the same entry

### Requirement: Tag display
Tags SHALL be rendered as badges in the viewer, with the first tag styled distinctly as a "mood" indicator. Pending AI-suggested tags SHALL be visually distinct from confirmed tags and offer a one-tap accept action.

#### Scenario: Entry has multiple tags
- **GIVEN** an entry with tags `["happy", "work", "outdoors"]`
- **THEN** the first tag (`"happy"`) is displayed as a mood badge (distinct style)
- **AND** the remaining tags (`"work"`, `"outdoors"`) are displayed as standard badges

#### Scenario: Entry has one tag
- **GIVEN** an entry with tags `["reflective"]`
- **THEN** only the mood badge is shown (no standard badges)

#### Scenario: Entry has no tags
- **GIVEN** an entry with an empty tags list
- **THEN** no badges are shown

#### Scenario: Entry has pending suggestions
- **GIVEN** an entry with confirmed tags `["work"]` and `pending_tags` `["beach", "family"]`
- **THEN** `"work"` is shown as a confirmed badge
- **AND** `"beach"` and `"family"` are shown as suggestion chips in a visually distinct style
- **AND** each suggestion chip offers a one-tap accept action

#### Scenario: Accepting a suggestion chip
- **GIVEN** an entry with `pending_tags` containing `"beach"`
- **WHEN** the user taps accept on the `"beach"` chip
- **THEN** `"beach"` becomes a confirmed tag badge
- **AND** `"beach"` is removed from the suggestion chips

### Requirement: Date navigation
The viewer SHALL show prev/next links to adjacent entries that have content, and the date badge is now an interactive control for jumping to any arbitrary date.

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
The user can edit an existing entry or create a new one for a date. Edit mode is reflected in the URL via `?edit=true`, but entering and leaving edit mode SHALL replace the current history entry rather than push a new one, so the browser/OS Back button does not accumulate edit-mode history entries. Tags SHALL be edited as a chip field: confirmed tags are shown as removable chips and new tags are added through an inline input. Changing the date field in the editor immediately navigates to that date's content rather than deferring to save time.

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

### Requirement: Title is optional
The system SHALL accept diary entries with an empty or blank title. When a title is absent, the UI SHALL display "Untitled" in place of a title.

#### Scenario: Empty title is accepted
- **WHEN** the user submits the editor form with an empty title field
- **THEN** the entry is saved successfully (no validation error)

#### Scenario: Untitled entry displayed in entry card
- **GIVEN** a saved entry with an empty title
- **WHEN** the entry appears in the entry list
- **THEN** the entry card displays "Untitled" in place of a title

#### Scenario: Untitled entry displayed in viewer
- **GIVEN** a saved entry with an empty title
- **WHEN** the user navigates to that entry's detail view
- **THEN** the viewer displays "Untitled" in place of a title

### Requirement: Upsert semantics
Saving an entry SHALL fully replace the entry for that date, preserving and updating its `tags_source_hash`, and preserving `pending_tags` except where acceptance or retagging changes them. The date field in the editor determines the save target and drives navigation between dates.

#### Scenario: Save overwrites existing entry
- **GIVEN** an entry for `2024-03-15` with title "Old Title" and body "Old body"
- **WHEN** the user edits it and saves with title "New Title" and body "New body"
- **THEN** the entry for `2024-03-15` now has title "New Title" and body "New body"
- **AND** there is still exactly one entry for `2024-03-15`

#### Scenario: Date can be changed in the editor
- **GIVEN** an entry for `2024-03-15`
- **WHEN** the user changes the date field to `2024-03-20` and saves
- **THEN** a new entry is created for `2024-03-20`
- **AND** the entry for `2024-03-15` is NOT deleted (upsert on target date, not a move)

#### Scenario: Save targets the current date field value
- **GIVEN** the user navigated to `2024-03-20` by changing the date field (with no unsaved changes on `2024-03-15`)
- **WHEN** the user fills in a title and body and saves
- **THEN** the entry is upserted at `2024-03-20`
- **AND** the entry at `2024-03-15` (if any) is left unchanged

#### Scenario: Save updates the tag source hash
- **GIVEN** an entry for `2024-03-15` with stored `tags_source_hash` `H`
- **WHEN** the user saves changed content for `2024-03-15`
- **THEN** the stored `tags_source_hash` is updated to match the new content

### Requirement: Tag source hash and retag on change
Each entry SHALL store a `tags_source_hash` derived from its title, body, and the sorted list of asset filenames it references. Saving an entry SHALL recompute this hash and trigger retagging when it differs from the stored value.

#### Scenario: Hash recorded on save
- **WHEN** an entry is saved
- **THEN** the system stores `tags_source_hash = hash(title + body + sorted(referenced asset filenames))`

#### Scenario: Content change triggers retag
- **GIVEN** an entry whose stored `tags_source_hash` is `H`
- **WHEN** the entry is saved with a title, body, or asset set that produces a different hash
- **THEN** the day is queued for retagging
- **AND** retagging runs asynchronously without blocking the save response

#### Scenario: No-op save does not retag
- **GIVEN** an entry whose stored `tags_source_hash` is `H`
- **WHEN** the entry is saved with content that produces the same hash `H`
- **THEN** no retagging is triggered

#### Scenario: Asset reordering does not retag
- **GIVEN** an entry referencing assets `a.jpg` and `b.jpg`
- **WHEN** the body is edited only to reorder those two image references
- **THEN** the computed hash is unchanged
- **AND** no retagging is triggered

### Requirement: Pending tags on an entry
An entry SHALL expose a `pending_tags` list, separate from confirmed `tags`, representing AI suggestions awaiting user acceptance.

#### Scenario: Pending tags returned with the entry
- **GIVEN** an entry that has pending AI suggestions
- **WHEN** the entry is fetched for viewing or editing
- **THEN** the response includes `pending_tags` distinct from confirmed `tags`

#### Scenario: Pending tags do not affect tag-based search
- **GIVEN** an entry with `pending_tags` containing `"beach"` and no confirmed `"beach"` tag
- **WHEN** entries are filtered by the tag `"beach"`
- **THEN** the entry is NOT returned (only confirmed tags match)

