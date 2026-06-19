## Why

Every spec under `openspec/specs/` was authored in a pre-tool heading convention (`# Feature:` title, `## Requirement:`, `### Scenario:`, no `## Purpose`/`## Requirements` sections). The current OpenSpec CLI expects `## Purpose` + `## Requirements` with `### Requirement:` / `#### Scenario:`. As a result:

- `openspec validate --specs` fails for all 7 specs ("Spec must have a Purpose section").
- `openspec archive` cannot apply spec deltas — its matcher looks for `### Requirement: <name>` and the files have `## Requirement: <name>`, so it aborts. The last change (`fix-back-nav-edit-mode`) had to be synced and archived by hand.

Normalizing the specs to the tool's schema lets `openspec validate` pass and `opsx:archive` run automatically, which the documented spec-first workflow depends on.

## What Changes

For each of the 7 spec files (`assets`, `auth`, `backup`, `entries`, `health`, `profile`, `search`):

- Add a one-line `## Purpose` section after the title.
- Wrap requirements under a `## Requirements` heading.
- Promote requirement headings `## Requirement:` → `### Requirement:`.
- Promote scenario headings `### Scenario:` → `#### Scenario:`.
- Remove the redundant `---` separators between requirements.

This is a **format-only migration** — no requirement text or scenario is added, removed, or semantically changed. Acceptance is `openspec validate --specs` passing for all 7 capabilities.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
<!-- none — this change reformats existing spec files for tool compatibility; no requirement behavior changes, so there are no spec deltas. Archive with `--skip-specs`. -->

## Impact

- `openspec/specs/{assets,auth,backup,entries,health,profile,search}/spec.md` — reformatted (7 files).
- No application code, API, or behavior changes.
- Unblocks `openspec validate` and `openspec archive` for all future changes.
- Archived changes under `openspec/changes/archive/` are left as-is (historical snapshots).
