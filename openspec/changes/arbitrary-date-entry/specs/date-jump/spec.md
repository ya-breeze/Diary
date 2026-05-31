# Feature: Date Jump

## ADDED Requirements

### Requirement: Jump to arbitrary date from entry viewer
The date badge in the entry viewer SHALL be interactive, allowing the user to navigate to any date by selecting it from a date picker.

#### Scenario: User clicks the date badge
- **WHEN** the user clicks the date badge in the entry viewer
- **THEN** a native date picker opens, pre-filled with the currently displayed date

#### Scenario: User picks a date with an existing entry
- **GIVEN** the user has opened the date picker from the entry viewer
- **WHEN** the user selects a date for which an entry exists
- **THEN** the app navigates to `/diary/{selected-date}` and the viewer shows that entry

#### Scenario: User picks a date with no existing entry
- **GIVEN** the user has opened the date picker from the entry viewer
- **WHEN** the user selects a date for which no entry exists
- **THEN** the app navigates to `/diary/{selected-date}` and the editor opens immediately

#### Scenario: User dismisses the date picker without selecting
- **WHEN** the user opens the date picker and then closes it without selecting a date
- **THEN** the current page remains unchanged

### Requirement: Jump to arbitrary date from the sidebar
The sidebar SHALL provide a secondary affordance (calendar icon button) to open any date, independent of the "New Entry" button.

#### Scenario: User clicks the calendar icon button
- **WHEN** the user clicks the calendar icon button in the sidebar header
- **THEN** a native date picker opens, pre-filled with today's date

#### Scenario: User picks a date via the sidebar picker — entry exists
- **GIVEN** the user has opened the date picker from the sidebar
- **WHEN** the user selects a date for which an entry exists
- **THEN** the app navigates to `/diary/{selected-date}` and the viewer shows that entry

#### Scenario: User picks a date via the sidebar picker — no entry
- **GIVEN** the user has opened the date picker from the sidebar
- **WHEN** the user selects a date for which no entry exists
- **THEN** the app navigates to `/diary/{selected-date}` and the editor opens immediately

#### Scenario: "New Entry" button is unaffected
- **WHEN** the user clicks the "New Entry" button (not the calendar icon)
- **THEN** the app navigates to today's entry in edit mode, as before
