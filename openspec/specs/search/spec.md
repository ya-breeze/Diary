# Feature: Search

## Requirement: Search diary entries by text
The search page provides full-text search across entries.

### Scenario: Query is too short — no search triggered
- **GIVEN** the user is on the search page
- **WHEN** they type 1 character (fewer than 2)
- **THEN** no API request is made and no results are shown

### Scenario: Query reaches minimum length — search triggers
- **WHEN** the user has typed at least 2 characters
- **THEN** a search request is made to `GET /v1/items?search=<query>` after a 300ms debounce

### Scenario: Debounce prevents rapid requests
- **WHEN** the user types quickly ("diary")
- **THEN** only one request is sent (after 300ms of no further typing), not one per keystroke

### Scenario: Results shown as entry cards
- **GIVEN** entries match the search query
- **THEN** each matching entry is shown as a card displaying its date, title, and tags

### Scenario: Clicking a result navigates to the entry
- **WHEN** the user clicks an entry card in search results
- **THEN** they are navigated to `/diary/[date]` for that entry

### Scenario: No results
- **GIVEN** the query matches no entries
- **THEN** an empty state is shown (no cards)

### Scenario: Loading state
- **GIVEN** a search request is in flight
- **THEN** a loading indicator is shown while waiting for results

---

## Requirement: Filter by tags
Tags can be included in the search request.

### Scenario: Tags filter combined with text search
- **GIVEN** the user searches with text and tags parameters
- **WHEN** the request reaches the server
- **THEN** results are filtered to entries that match both the text and the specified tags
- **AND** tags in the query string are comma-separated and trimmed of whitespace

### Scenario: Empty tag strings are ignored
- **GIVEN** the tags parameter contains empty segments (e.g. `"happy,,work"`)
- **THEN** the server ignores the empty segments and filters only by `"happy"` and `"work"`

---

## Requirement: Search input auto-focus
The search input is focused when the search page loads.

### Scenario: Page load focuses the search input
- **WHEN** the user navigates to the search page
- **THEN** the search input has focus so the user can type immediately
