# Feature: Diary Entries

## Requirement: View an entry by date
Navigating to `/diary/[date]` shows the entry for that date.

### Scenario: Entry exists for the date
- **GIVEN** an entry exists for `2024-03-15`
- **WHEN** the user navigates to `/diary/2024-03-15`
- **THEN** the entry's title, body (rendered as markdown), and tags are displayed

### Scenario: No entry exists for the date
- **GIVEN** no entry exists for `2024-03-15`
- **WHEN** the user navigates to `/diary/2024-03-15`
- **THEN** the editor is shown immediately (not the viewer) so the user can create a new entry
- **AND** the date field is pre-filled with `2024-03-15`

### Scenario: Entry data is family-scoped
- **GIVEN** two users share a family
- **WHEN** either user navigates to the same date
- **THEN** they see the same entry

---

## Requirement: Tag display
Tags are rendered as badges in the viewer, with the first tag styled distinctly as a "mood" indicator.

### Scenario: Entry has multiple tags
- **GIVEN** an entry with tags `["happy", "work", "outdoors"]`
- **THEN** the first tag (`"happy"`) is displayed as a mood badge (distinct style)
- **AND** the remaining tags (`"work"`, `"outdoors"`) are displayed as standard badges

### Scenario: Entry has one tag
- **GIVEN** an entry with tags `["reflective"]`
- **THEN** only the mood badge is shown (no standard badges)

### Scenario: Entry has no tags
- **GIVEN** an entry with an empty tags list
- **THEN** no badges are shown

---

## Requirement: Date navigation
The viewer shows prev/next links to adjacent entries that have content.

### Scenario: Both previous and next entries exist
- **GIVEN** an entry for `2024-03-15` with a previous entry on `2024-03-10` and next entry on `2024-03-20`
- **THEN** the viewer shows a left arrow linking to `/diary/2024-03-10` and a right arrow linking to `/diary/2024-03-20`

### Scenario: No previous entry
- **GIVEN** the current entry is the oldest in the diary
- **THEN** the left arrow is rendered but is non-interactive (greyed out)

### Scenario: No next entry
- **GIVEN** the current entry is the newest in the diary
- **THEN** the right arrow is rendered but is non-interactive (greyed out)

### Scenario: Navigation skips dates with no entries
- **GIVEN** entries exist on `2024-03-01` and `2024-03-15` but not on any date in between
- **WHEN** the user is viewing the `2024-03-01` entry
- **THEN** the next arrow links directly to `2024-03-15` (not to `2024-03-02`)

---

## Requirement: Edit an entry
The user can edit an existing entry or create a new one for a date.

### Scenario: Open editor from viewer
- **GIVEN** the user is viewing an existing entry
- **WHEN** they click the Edit button
- **THEN** the URL changes to `/diary/[date]?edit=true`
- **AND** a full-screen editor overlay is shown pre-filled with the entry's title, date, tags, and body

### Scenario: Title is required
- **WHEN** the user submits the editor form with an empty title
- **THEN** a validation error "Title is required" is shown
- **AND** the entry is not saved

### Scenario: Tags are comma-separated
- **GIVEN** the user types `"happy, work,  outdoors "` in the tags field
- **WHEN** the entry is saved
- **THEN** the server stores `["happy", "work", "outdoors"]` (trimmed, empty values removed)

### Scenario: Whitespace-only tags are stripped
- **GIVEN** the user submits tags `"happy, , ,work"`
- **WHEN** the entry is saved
- **THEN** the server stores `["happy", "work"]` (the empty segments are discarded)

### Scenario: Body is optional
- **WHEN** the user saves an entry with a title but no body text
- **THEN** the entry is saved with an empty body (no error)

### Scenario: Save closes editor and refreshes data
- **WHEN** the user saves successfully
- **THEN** the editor closes, the URL returns to `/diary/[date]` (without `?edit=true`)
- **AND** the viewer re-fetches and shows the updated entry

### Scenario: Cancel discards changes
- **WHEN** the user clicks Cancel in the editor
- **THEN** the URL returns to `/diary/[date]` without saving
- **AND** the entry is unchanged

### Scenario: Edit state syncs with URL
- **GIVEN** the user is on `/diary/[date]?edit=true`
- **WHEN** they press the browser Back button
- **THEN** the URL changes to `/diary/[date]` and the editor closes

---

## Requirement: Upsert semantics
Saving an entry fully replaces the entry for that date.

### Scenario: Save overwrites existing entry
- **GIVEN** an entry for `2024-03-15` with title "Old Title" and body "Old body"
- **WHEN** the user edits it and saves with title "New Title" and body "New body"
- **THEN** the entry for `2024-03-15` now has title "New Title" and body "New body"
- **AND** there is still exactly one entry for `2024-03-15`

### Scenario: Date can be changed in the editor
- **GIVEN** an entry for `2024-03-15`
- **WHEN** the user changes the date field to `2024-03-20` and saves
- **THEN** a new entry is created for `2024-03-20`
- **AND** the entry for `2024-03-15` is NOT deleted (upsert on target date, not a move)
