## ADDED Requirements

### Requirement: Tag usage counts endpoint
The system SHALL expose the authenticated family's distinct tags together with the number of entries using each, via `GET /v1/tags/stats`, without altering the existing `GET /v1/tags` endpoint.

#### Scenario: Counts reflect how many entries use each tag
- **GIVEN** a family with entries tagged `["family","travel"]`, `["family","work"]`, and `["family"]`
- **WHEN** `GET /v1/tags/stats` is called
- **THEN** the response includes `family` with count `3`, `travel` with count `1`, and `work` with count `1`

#### Scenario: A tag counts once per entry regardless of duplicates
- **GIVEN** an entry whose stored tags somehow contain `family` twice
- **WHEN** `GET /v1/tags/stats` is called
- **THEN** that entry contributes `1` to the count for `family` (not 2)

#### Scenario: Results are sorted by count then name
- **GIVEN** tag counts `work=5`, `family=8`, `travel=5`
- **WHEN** `GET /v1/tags/stats` is called
- **THEN** the list is ordered `family (8)`, `travel (5)`, `work (5)` — count descending, then name ascending for ties

#### Scenario: Stats are family-scoped
- **GIVEN** two families each with their own tagged entries
- **WHEN** `GET /v1/tags/stats` is called for one family
- **THEN** only that family's tags and counts are returned

#### Scenario: Empty when no tags exist
- **GIVEN** a family with no tagged entries
- **WHEN** `GET /v1/tags/stats` is called
- **THEN** an empty list is returned

#### Scenario: Available without AI configuration
- **GIVEN** the server has no `GEMINI_API_KEY` and the family has not enabled AI tagging
- **WHEN** `GET /v1/tags/stats` is called
- **THEN** the family's tags and counts are still returned

### Requirement: Tags page
The application SHALL provide a dedicated Tags page that lists every distinct tag of the family with its usage count, sorted by count descending.

#### Scenario: Page lists tags with counts
- **GIVEN** the family's tag stats are `family=8`, `travel=5`, `work=3`
- **WHEN** the user opens the Tags page
- **THEN** each tag is shown with its count, ordered `family`, `travel`, `work`

#### Scenario: Reached from the profile Tags stat
- **GIVEN** the user is on the profile page
- **WHEN** they click the "Tags" statistic card
- **THEN** they are navigated to the Tags page

#### Scenario: Empty state when no tags exist
- **GIVEN** the family has no tagged entries
- **WHEN** the user opens the Tags page
- **THEN** an empty-state message is shown instead of a tag list

### Requirement: Browse entries by tag
The Tags page SHALL let the user view all entries carrying a chosen tag, using the existing tag filter (`GET /v1/items?tags=`), which matches entries containing any of the requested tags (OR semantics). The browsed tag SHALL be reflected in the page URL (e.g. `/tags?tag=<name>`) so the view is deep-linkable and bookmarkable.

#### Scenario: Selecting a tag shows its entries
- **GIVEN** three entries are tagged `travel`
- **WHEN** the user selects the `travel` tag to browse
- **THEN** all three entries are listed
- **AND** clicking an entry navigates to `/diary/[date]` for that entry

#### Scenario: Browsing a tag is reflected in the URL
- **GIVEN** the user selects the `travel` tag to browse
- **THEN** the page URL includes the tag (e.g. `/tags?tag=travel`)
- **AND** navigating directly to that URL shows the same browse view for `travel`

#### Scenario: Multiple tags use OR semantics
- **GIVEN** entry A is tagged only `travel` and entry B is tagged only `family`
- **WHEN** the user browses by both `travel` and `family`
- **THEN** both entry A and entry B are listed

### Requirement: Rename a tag across all entries
The system SHALL rename a tag on every entry of the family via `PATCH /v1/tags/{name}` with body `{ "newName": "<new>" }`, merging into the target name where an entry already carries it.

#### Scenario: Rename updates every entry carrying the tag
- **GIVEN** entries tagged `["vacaiton","work"]` and `["vacaiton"]`
- **WHEN** `PATCH /v1/tags/vacaiton` is called with `newName` `vacation`
- **THEN** the entries become `["vacation","work"]` and `["vacation"]`

#### Scenario: Rename into an existing tag merges without duplicates
- **GIVEN** an entry tagged `["vacaiton","vacation"]`
- **WHEN** `vacaiton` is renamed to `vacation`
- **THEN** the entry's tags become `["vacation"]` (a single, de-duplicated tag — the rename is not blocked)

#### Scenario: Rename is family-scoped
- **GIVEN** two families both use the tag `travel`
- **WHEN** one family renames `travel` to `trips`
- **THEN** only that family's entries change; the other family's `travel` is untouched

#### Scenario: Blank or unchanged new name is rejected
- **WHEN** `PATCH /v1/tags/{name}` is called with a blank `newName` or a `newName` equal to the existing name
- **THEN** the request is rejected with `400` and no entries are modified

#### Scenario: Rename of a non-existent tag changes nothing
- **GIVEN** no entry carries the tag `nonexistent`
- **WHEN** `nonexistent` is renamed
- **THEN** no entries are modified

#### Scenario: Rename is atomic
- **GIVEN** several entries carry the tag being renamed
- **WHEN** the rename fails partway through
- **THEN** no entry is left partially updated (the operation rolls back)

### Requirement: Delete a tag across all entries
The system SHALL remove a tag from every entry of the family via `DELETE /v1/tags/{name}`.

#### Scenario: Delete removes the tag from every entry
- **GIVEN** entries tagged `["misc","work"]` and `["misc"]`
- **WHEN** `DELETE /v1/tags/misc` is called
- **THEN** the entries become `["work"]` and `[]`

#### Scenario: Delete is family-scoped
- **GIVEN** two families both use the tag `misc`
- **WHEN** one family deletes `misc`
- **THEN** only that family's entries change

#### Scenario: Delete of a non-existent tag changes nothing
- **GIVEN** no entry carries the tag `nonexistent`
- **WHEN** `DELETE /v1/tags/nonexistent` is called
- **THEN** no entries are modified

### Requirement: Tag management actions are confirmed in the UI
The Tags page SHALL expose rename and delete actions per tag, with delete gated behind a confirmation that states how many entries are affected.

#### Scenario: Delete asks for confirmation with the affected count
- **GIVEN** the tag `misc` is used by 4 entries
- **WHEN** the user activates delete for `misc`
- **THEN** a confirmation prompt referencing "4 entries" is shown
- **AND** the tag is only removed if the user confirms

#### Scenario: Rename presents an edit affordance
- **GIVEN** the user is on the Tags page
- **WHEN** they activate the rename action for a tag
- **THEN** they can enter a new name, and on submit the tag is renamed across all entries
- **AND** the list refreshes to reflect the new name and merged count
