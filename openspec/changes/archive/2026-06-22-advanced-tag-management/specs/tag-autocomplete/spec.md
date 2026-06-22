## MODIFIED Requirements

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
