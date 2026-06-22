# Feature: Profile

## Purpose
How the profile page presents the logged-in user's identity, writing statistics, top tags, family information, and logout.
## Requirements
### Requirement: User identity display
The profile page SHALL show who is logged in.

#### Scenario: Avatar initials derived from email
- **GIVEN** the logged-in user's email is `alice@example.com`
- **THEN** the avatar shows `AL` (first two characters, uppercased)

#### Scenario: Member since date
- **GIVEN** the user has a `startDate` field set
- **THEN** the profile shows "Member since [Month YYYY]" (e.g. "Member since March 2023")

#### Scenario: No startDate set
- **GIVEN** the user has no `startDate`
- **THEN** the "Member since" line is not shown

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

### Requirement: Top tags
The profile SHALL show the top 5 most-used tags. Each top tag SHALL be clickable and navigate to the Tags page browse view for that tag.

#### Scenario: Top 5 tags ranked by frequency
- **GIVEN** tag frequencies: `happy=10, work=8, outdoors=5, travel=3, food=2, misc=1`
- **THEN** the top tags section shows `["happy", "work", "outdoors", "travel", "food"]` (top 5 only)

#### Scenario: Fewer than 5 unique tags
- **GIVEN** the family has only 3 unique tags
- **THEN** all 3 are shown (no padding or placeholder for the missing slots)

#### Scenario: No tags used
- **GIVEN** no entry has any tags
- **THEN** the top tags section is not shown

#### Scenario: Clicking a top tag browses its entries
- **GIVEN** the top tags section shows `travel`
- **WHEN** the user clicks the `travel` tag
- **THEN** they are navigated to the Tags page browse view for `travel` (e.g. `/tags?tag=travel`)

### Requirement: Family information
The profile SHALL show the family the user belongs to.

#### Scenario: Family name and members shown
- **GIVEN** the user belongs to a family named "Smith Family" with members `alice@example.com` and `bob@example.com`
- **THEN** the family section shows the family name and both email addresses

#### Scenario: Family section hidden when data is unavailable
- **GIVEN** the family API call fails or returns no data
- **THEN** the family section is not shown (no error message)

### Requirement: Logout from profile
The user SHALL be able to log out from the profile page.

#### Scenario: Logout button visible
- **THEN** a "Log out" button is always visible on the profile page

#### Scenario: Logout clears session and redirects
- **WHEN** the user clicks "Log out"
- **THEN** the auth state is cleared and the user is redirected to `/login`
- (See also: auth/spec.md — Logout requirement for server-side behaviour)

