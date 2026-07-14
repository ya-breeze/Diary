## MODIFIED Requirements

### Requirement: Tag display
Tags SHALL be rendered as badges in the viewer, with the first tag styled distinctly as a "mood" indicator. Pending AI-suggested tags SHALL be visually distinct from confirmed tags and offer a one-tap accept action. On narrow viewports, every confirmed tag SHALL remain visible and usable without horizontal page scrolling or clipping.

#### Scenario: Entry has multiple tags
- **GIVEN** an entry with tags `["happy", "work", "outdoors"]`
- **THEN** the first tag (`"happy"`) is displayed as a mood badge (distinct style)
- **AND** the remaining tags (`"work"`, `"outdoors"`) are displayed as standard badges

#### Scenario: Entry has one tag
- **GIVEN** an entry with tags `["reflective"]`
- **THEN** only the mood badge is shown (no standard badges)

#### Scenario: Entry has no tags
- **GIVEN** an entry with an empty tags list
- **THEN** no badges are shown

#### Scenario: Entry has pending suggestions
- **GIVEN** an entry with confirmed tags `["work"]` and `pending_tags` `["beach", "family"]`
- **THEN** `"work"` is shown as a confirmed badge
- **AND** `"beach"` and `"family"` are shown as suggestion chips in a visually distinct style
- **AND** each suggestion chip offers a one-tap accept action

#### Scenario: Accepting a suggestion chip
- **GIVEN** an entry with `pending_tags` containing `"beach"`
- **WHEN** the user taps accept on the `"beach"` chip
- **THEN** `"beach"` becomes a confirmed tag badge
- **AND** `"beach"` is removed from the suggestion chips

#### Scenario: Multiple tags on a narrow viewport
- **GIVEN** an entry with tags `["happy", "work", "outdoors"]`
- **AND** the user views the entry on a viewport narrower than the `md` breakpoint
- **THEN** the mood badge and every standard tag badge are visible and usable
- **AND** the tag badges do not require horizontal page scrolling to be accessed
