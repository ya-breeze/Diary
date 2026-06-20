## ADDED Requirements

### Requirement: List the family's existing tags
The system SHALL expose the authenticated family's distinct existing tags via `GET /v1/tags`, deduplicated and sorted, without requiring any AI configuration.

#### Scenario: Distinct tags returned
- **GIVEN** a family whose entries use the tags `["family", "travel"]` and `["family", "work"]`
- **WHEN** `GET /v1/tags` is called
- **THEN** the response contains `["family", "travel", "work"]` (deduplicated and sorted)

#### Scenario: Empty when no tags exist
- **GIVEN** a family with no tagged entries
- **WHEN** `GET /v1/tags` is called
- **THEN** an empty list is returned

#### Scenario: Tags are family-scoped
- **GIVEN** two families each with their own tagged entries
- **WHEN** `GET /v1/tags` is called for one family
- **THEN** only that family's tags are returned

#### Scenario: Available without AI configuration
- **GIVEN** the server has no `GEMINI_API_KEY` and the family has not enabled AI tagging
- **WHEN** `GET /v1/tags` is called
- **THEN** the family's distinct tags are still returned (the endpoint does not depend on the AI feature)

### Requirement: Tags-field autocomplete in the editor
The entry editor SHALL offer a typeahead of existing tags while the user edits the comma-separated tags field, completing the current token from the family's vocabulary.

#### Scenario: Matching tags appear as the user types
- **GIVEN** the family's existing tags include `"travel"` and `"family"`
- **WHEN** the user types `tra` in the last token of the tags field
- **THEN** a dropdown shows `"travel"` as a match

#### Scenario: Selecting a suggestion completes the token
- **GIVEN** the tags field contains `work, tra` and a dropdown shows `"travel"`
- **WHEN** the user selects `"travel"`
- **THEN** the field becomes `work, travel, ` (only the last token is replaced; earlier tags are preserved)

#### Scenario: Already-entered tags are excluded
- **GIVEN** the tags field already contains `travel, ` and the user starts a new token
- **WHEN** the autocomplete list is shown
- **THEN** `"travel"` is not offered again

#### Scenario: No matches shows no dropdown
- **GIVEN** the user's current token matches none of the existing tags
- **THEN** no autocomplete dropdown is shown
- **AND** the user can still type a brand-new tag freely
