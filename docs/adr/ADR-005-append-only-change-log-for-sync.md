# ADR-005: Append-Only Change Log for Mobile Sync

## Status
Accepted

## Context and Problem Statement

A mobile companion app needs to synchronize diary entries without fetching the entire dataset on every poll. The sync mechanism must work over unreliable connections and allow the client to remain stateless between sessions.

## Decision Drivers

- Mobile client must be able to fetch only changes since its last sync point
- Client must not need to track its own local state to reconstruct the current dataset
- Deleted entries must be communicable to the client (tombstones)
- Implementation must be simple — no real-time infrastructure

## Considered Options

- **Full sync** — client fetches all entries every time; simple but wasteful
- **Append-only `ItemChange` log with snapshots** — client polls from a watermark ID
- **WebSocket / SSE push** — real-time, but requires persistent connections and infra

## Decision Outcome

Chosen: **Append-only `ItemChange` table**, each row recording the operation (`created`, `updated`, `deleted`) and a full snapshot of the item's state at the time of the change.

The client stores the highest `id` it has seen. On next sync it calls `GET /v1/sync?since=<id>` and receives only new change records. Because each record includes a full `ItemSnapshot`, the client can apply changes without querying individual items.

Deleted items include a snapshot of the last known state so the client can display what was removed.

### Pros

- Client is fully stateless between sessions — only needs to remember the watermark ID
- Simple HTTP polling — no persistent connection or push infrastructure
- Deleted entries are first-class, not just missing rows
- Works correctly over flaky mobile connections

### Cons

- The `item_changes` table grows without bound — no cleanup mechanism exists yet
- Snapshot storage duplicates data already in the `items` table
- High-frequency edits produce many change records for the same date
