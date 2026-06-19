# CLAUDE.md — Diary

Guidance for Claude Code when working in the Diary repository.

## Architecture Decision Records (ADRs)

ADRs live in `docs/adr/` (`ADR-NNN-<slug>.md`). They follow a fixed format:
Status → Context and Problem Statement → Decision Drivers → Considered Options →
Decision Outcome → Pros/Cons → (optional) Related.

### ADR status lifecycle — MANDATORY

- A new ADR introduced by an in-flight change is created with **Status: Proposed**.
- **On merge**, flip the ADR's status from `Proposed` to `Accepted` as the **last step of the PR, before finishing it** — together with archiving the OpenSpec change. Do not leave a merged feature's ADR as `Proposed`.
- When an ADR replaces an earlier decision, mark the old one `Superseded by ADR-NNN` and reference it.

### When a change warrants an ADR

Add an ADR for genuinely architectural decisions: a new external dependency, a
new cross-cutting pattern, security/data-model/migration shifts. Implementation
details belong in the OpenSpec change's `design.md`, not an ADR. If an existing
ADR's stated facts (e.g. an enumeration) become stale, add a brief
`> **Update (ADR-NNN):** …` note rather than rewriting an Accepted record.

## OpenSpec workflow

This project uses OpenSpec (`openspec/`). Before writing implementation code,
ensure an active change exists (`openspec/changes/<name>/`); use `opsx:propose`
to create one and `opsx:apply` to implement from its `tasks.md`. Archive with
`opsx:archive` on completion. See the global working-environment guide for the
full mandatory workflow.
