## Tasks

- [x] Add `filterTags` call in `newItemResponse` for both `Tags` and `PendingTags`
- [x] Add `scrubBlankTags` startup migration in `migration.go`
- [x] Add `filterStringList` helper in `migration.go`
- [x] Wire `scrubBlankTags` into `storage.Open()` after `autoMigrateModels`
- [x] Verify no blank tags returned by items API (WIP E2E check)
