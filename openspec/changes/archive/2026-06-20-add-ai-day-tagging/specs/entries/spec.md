## ADDED Requirements

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

## MODIFIED Requirements

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

### Requirement: Upsert semantics
Saving an entry SHALL fully replace the entry for that date, preserving and updating its `tags_source_hash`, and preserving `pending_tags` except where acceptance or retagging changes them.

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

#### Scenario: Save updates the tag source hash
- **GIVEN** an entry for `2024-03-15` with stored `tags_source_hash` `H`
- **WHEN** the user saves changed content for `2024-03-15`
- **THEN** the stored `tags_source_hash` is updated to match the new content
