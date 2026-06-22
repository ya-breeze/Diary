## ADDED Requirements

### Requirement: Tag filter chips on the search page
The search page SHALL let the user filter results by tags selected as chips, combined with the free-text query. Multiple selected tags match with OR semantics among themselves and AND with the text.

#### Scenario: Adding a tag chip filters results
- **GIVEN** the user is on the search page
- **WHEN** they add the tag `travel` to the filter
- **THEN** the request includes `tags=travel` and results are limited to entries carrying `travel`

#### Scenario: Multiple tag chips use OR among tags
- **GIVEN** entry A is tagged only `travel` and entry B is tagged only `family`
- **WHEN** the user adds both `travel` and `family` as filter chips
- **THEN** both entry A and entry B appear in the results

#### Scenario: Tag filter combines with text using AND
- **GIVEN** the user has text `beach` and a tag chip `travel`
- **WHEN** the search runs
- **THEN** results are limited to entries matching the text `beach` AND carrying the tag `travel`

#### Scenario: Removing a tag chip widens results
- **GIVEN** the filter has chips `travel` and `family`
- **WHEN** the user removes the `family` chip
- **THEN** the request includes only `tags=travel` and results update accordingly

#### Scenario: Tag-only filter runs without text
- **GIVEN** the free-text query is empty
- **WHEN** the user has at least one tag chip selected
- **THEN** a search runs filtered by the selected tag(s), independent of the 2-character text minimum
