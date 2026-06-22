## MODIFIED Requirements

### Requirement: Writing statistics
The profile SHALL show three summary stats derived from all of the family's entries. The "Tags" stat SHALL link to the dedicated Tags page.

#### Scenario: Entry count
- **GIVEN** the family has 42 diary entries
- **THEN** the "Entries" stat shows `42`

#### Scenario: Unique tag count
- **GIVEN** entries use tags `["happy", "work", "happy", "outdoors"]` across all entries
- **THEN** the "Tags" stat shows `3` (unique tags: happy, work, outdoors)

#### Scenario: Tags stat links to the Tags page
- **GIVEN** the user is on the profile page
- **WHEN** they click the "Tags" statistic card
- **THEN** they are navigated to the Tags page

#### Scenario: Writing streak — active streak
- **GIVEN** entries exist for today and each of the previous 4 consecutive days
- **THEN** the "Streak" stat shows `5d`

#### Scenario: Writing streak — streak counts from yesterday if no entry today
- **GIVEN** the most recent entry is from yesterday and the 3 days before that also have entries
- **THEN** the streak is `4d` (timezone flexibility: streak is not broken if most recent entry is yesterday)

#### Scenario: Writing streak — broken by a gap
- **GIVEN** the most recent entry is from 2 days ago
- **THEN** the streak is `0d` (more than one day gap breaks the streak)

#### Scenario: Writing streak — no entries
- **GIVEN** the family has no entries
- **THEN** the streak is `0d`

#### Scenario: Stats show loading placeholder while data loads
- **GIVEN** the entries request is in flight
- **THEN** all three stat values display `—` instead of numbers
