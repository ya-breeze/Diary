# Tasks

## 1. Backend — distinct tags endpoint

- [ ] 1.1 Add `Storage.GetDistinctTags(familyID uuid.UUID) ([]string, error)` (dedupe + sort); regenerate the storage mock
- [ ] 1.2 Add `GET /v1/tags` to `api/openapi.yaml` returning `{tags: [string]}`; run `make generate`
- [ ] 1.3 Implement the service handler + interface/adapter bridge; register at startup
- [ ] 1.4 Backend tests: distinct/sorted/deduped, empty, family-scoped

## 2. Frontend — tags autocomplete

- [ ] 2.1 Add `diaryApi.getTags()` calling `GET /v1/tags`
- [ ] 2.2 In `EntryEditor`, fetch tags on open; render a dropdown filtering existing tags by the active (last) comma-separated token, excluding tags already entered
- [ ] 2.3 On select, replace only the active token and append `, `; keep free typing when there is no match
- [ ] 2.4 E2E: dropdown shows a matching existing tag; selecting it completes the token; already-entered tags excluded

## 3. Verification

- [ ] 3.1 `make build` + lint clean for changed files
- [ ] 3.2 Run E2E against the WIP stack
