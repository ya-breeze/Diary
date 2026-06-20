## Why

Items created before the save-path `filterTags` guard was added could carry blank
or whitespace-only tag strings (e.g. `[""]`, `[""," "]`). These propagated to the
"Top tags" UI as empty chips, degrading the user experience.

## What Changes

Two-pronged fix to eliminate blank tags from all surfaces:

1. **API output filter** — `newItemResponse` applies `filterTags` to `item.Tags` and
   `item.PendingTags` before building the JSON response, so no blank tag ever reaches
   the frontend regardless of stored data.

2. **Startup scrub** — `scrubBlankTags` runs after `AutoMigrate` and rewrites any item
   whose stored tag lists contain empty or whitespace-only entries, logging the count.
   Idempotent and non-fatal.

## Capabilities

### Modified Capabilities
- `items`: items API now strips blank tags at output; startup migration cleans stored blanks

### Removed Capabilities
<!-- none -->
