# ADR-009: Date-Keyed Diary Entries (One Entry per Day per Family)

## Status
Accepted

## Context and Problem Statement

A personal diary is naturally organized by calendar date. The data model must decide whether to allow multiple entries per day or enforce one entry per day, and what the primary identifier for an entry should be.

## Decision Drivers

- The mental model of a diary is "what happened today" — one page per day
- Navigation UI is date-based (previous/next day, jump to date)
- The mobile sync client references entries by date, not an opaque ID
- Simplicity: no need to order or merge multiple same-day entries

## Considered Options

- **UUID primary key, any number of entries per day** — flexible, but complicates navigation and sync
- **Timestamp primary key** — unique per entry but arbitrary, breaks the "one page per day" model
- **Date string (`YYYY-MM-DD`) as the natural key per family** — maps directly to the user's mental model

## Decision Outcome

Chosen: **`date` (string `YYYY-MM-DD`) as the natural key**, scoped per family via a composite unique index `(family_id, date)`.

An entry is uniquely identified by `(family_id, date)`. Upsert semantics apply: creating an entry for an existing date updates it. The sync log (`ItemChange`) also keys changes by `date`, so the mobile client can apply patches by date without needing a separate ID lookup.

### Pros

- API, UI, and sync client all use the same human-readable key
- Date navigation (previous/next) is a simple string increment — no ID translation
- Prevents accidental duplicate entries for the same day

### Cons

- Cannot represent multiple distinct entries on the same day
- The date string format (`YYYY-MM-DD`) must be validated consistently across all entry points
- Changing an entry's date would require deleting and re-creating it (not a supported operation)
