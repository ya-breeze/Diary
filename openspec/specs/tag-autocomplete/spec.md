# tag-autocomplete Specification

## Purpose
TBD - created by archiving change add-tag-autocomplete. Update Purpose after archive.
## Requirements
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
The entry editor SHALL offer a typeahead of existing tags while the user types in the chip field's inline add-tag input, completing the current input from the family's vocabulary.

#### Scenario: Matching tags appear as the user types
- **GIVEN** the family's existing tags include `"travel"` and `"family"`
- **WHEN** the user types `tra` in the inline add-tag input
- **THEN** a dropdown shows `"travel"` as a match

#### Scenario: Selecting a suggestion adds it as a chip
- **GIVEN** the chip field already contains `work` and the inline input shows a dropdown match `"travel"`
- **WHEN** the user selects `"travel"`
- **THEN** `travel` is added as a chip alongside `work` (existing chips are preserved)
- **AND** the inline input is cleared, ready for the next tag

#### Scenario: Already-entered tags are excluded
- **GIVEN** the chip field already contains `travel`
- **WHEN** the autocomplete list is shown for a new input
- **THEN** `"travel"` is not offered again

#### Scenario: No matches shows no dropdown
- **GIVEN** the user's current input matches none of the existing tags
- **THEN** no autocomplete dropdown is shown
- **AND** the user can still commit a brand-new tag freely

