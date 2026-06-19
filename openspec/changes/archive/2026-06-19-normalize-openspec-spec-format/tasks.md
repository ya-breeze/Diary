## 1. Normalize each spec file

For each file: add `## Purpose` (one line) after the title, add a `## Requirements` heading, promote `## Requirement:` → `### Requirement:` and `### Scenario:` → `#### Scenario:`, drop `---` separators, and reword each requirement statement to use a normative keyword (SHALL/MUST). Scenario bodies must be copied verbatim. Run `openspec validate --specs <capability>` after each.

- [x] 1.1 `openspec/specs/assets/spec.md` (4 requirements)
- [x] 1.2 `openspec/specs/auth/spec.md` (5 requirements)
- [x] 1.3 `openspec/specs/backup/spec.md` (5 requirements)
- [x] 1.4 `openspec/specs/entries/spec.md` (5 requirements)
- [x] 1.5 `openspec/specs/health/spec.md` (5 requirements)
- [x] 1.6 `openspec/specs/profile/spec.md` (5 requirements)
- [x] 1.7 `openspec/specs/search/spec.md` (3 requirements)

## 2. Verify

- [x] 2.1 `openspec validate --specs` passes for all 7 capabilities (0 failed)
- [x] 2.2 Spot-check diffs to confirm scenario bodies unchanged (heading/wrapper + normative rewording only; scenario counts match `main`)

## 3. Finalize

- [ ] 3.1 Get user approval, then archive with `openspec archive normalize-openspec-spec-format --skip-specs` (no spec deltas; archive the change folder only)
- [ ] 3.2 Squash-merge to `main`
