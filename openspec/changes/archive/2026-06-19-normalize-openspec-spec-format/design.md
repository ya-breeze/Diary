## Context

The 7 spec files use a heading convention that predates the current OpenSpec CLI schema. The tool's parser/validator requires, per spec file:

```
## Purpose
<one line>

## Requirements
### Requirement: <name>
<text>

#### Scenario: <name>
- **WHEN** ...
- **THEN** ...
```

The repo's files instead use a top-level `# Feature: <title>`, then `## Requirement:` and `### Scenario:` with `---` separators and no Purpose/Requirements sections. This breaks `openspec validate --specs` (all 7 fail) and `openspec archive` (delta header matching fails).

## Goals / Non-Goals

**Goals:**
- All 7 specs pass `openspec validate --specs`.
- `openspec archive` can apply deltas against these specs going forward.
- Zero change to requirement/scenario meaning — purely structural.

**Non-Goals:**
- Editing requirement text, adding/removing requirements, or changing behavior.
- Reformatting archived change snapshots under `openspec/changes/archive/` (left as historical record).
- Changing application code.

## Decisions

- **Mechanical heading promotion + wrapper insertion.** For each file: keep the `# Feature:` title line, insert `## Purpose` (one sentence derived from the existing title/content) and `## Requirements`, then bump every `## Requirement:` to `### Requirement:` and every `### Scenario:` to `#### Scenario:`. Drop the `---` separators (redundant under a `## Requirements` parent).
- **Purpose text is descriptive, not normative.** One neutral sentence summarizing the capability; it must not introduce new requirements.
- **Verify with the tool, not by eye.** Acceptance is `openspec validate --specs` returning all-passed; this is the objective signal that the migration matches the schema.
- **No spec deltas for this change.** Because no requirement changes, the change carries no delta specs and is archived with `openspec archive --skip-specs`.

## Risks / Trade-offs

- [Accidental semantic change while editing 7 files] → Edits are heading-level and wrapper-only; requirement/scenario bodies are copied verbatim. Diff-review each file and rely on `openspec validate` plus a line-count/word check on requirement bodies.
- [Heading bump misaligns a nested list or code block] → Validate per file after editing; fix before moving on.
