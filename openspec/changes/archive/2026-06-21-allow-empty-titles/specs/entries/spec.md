## REMOVED Requirements

### Requirement: Title is required
**Reason**: Title is optional — users can save entries without one; the UI displays "Untitled" as a fallback.
**Migration**: No migration needed; empty titles were always accepted by the backend.

## ADDED Requirements

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
