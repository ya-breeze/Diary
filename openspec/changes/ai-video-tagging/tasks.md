# Tasks

## 1. Data model & API

- [x] 1.1 Add `AITaggingUseVideo bool gorm:"default:false"` to `pkg/database/models/family.go`; update `FromDB()` to include `aiTaggingUseVideo` in `FamilyResponse`
- [x] 1.2 Extend `SetFamilyAISettings` in `pkg/database/storage.go` to accept a 5th `useVideo bool` parameter; update all callers and the mock
- [x] 1.3 Add `aiTaggingUseVideo` to `FamilyResponse` and as an optional `boolean` field in `FamilySettingsRequest` in `api/openapi.yaml`; run `make generate` to regenerate server types
- [x] 1.4 Update `api_family_service.go`: read `current.AITaggingUseVideo`, apply PATCH semantics, pass `useVideo` to `SetFamilyAISettings`; enforce that `useVideo` is forced false when `enabled` is false

## 2. Video keyframe extraction

- [x] 2.1 Create `pkg/ai/video_loader.go` with `LoadVideoKeyframes(body, dataPath, familyID string) []ImageAsset`; detect video files by reading first 512 bytes via `http.DetectContentType`; supported MIME types: `video/mp4`, `video/quicktime`, `video/x-msvideo`, `video/webm`, `video/x-matroska`
- [x] 2.2 Implement `extractKeyframes(path string) ([]ImageAsset, error)`: check for ffmpeg via `exec.LookPath`; use `os.MkdirTemp` for output frames; run `ffmpeg -i <path> -vf fps=1/<step> -frames:v 5 -q:v 3 <tmpdir>/frame%03d.jpg` with a 30s `context.WithTimeout`; read output JPEGs; clean up temp dir in `defer`; return nil on any error (graceful degrade)
- [x] 2.3 Compute `step = max(1, duration/5)` using `-show_entries format=duration` from ffprobe (or ffmpeg itself via stderr probe); fall back to `step = 10` if duration cannot be determined
- [x] 2.4 Write unit tests in `video_loader_test.go`: ffmpeg unavailable returns nil; supported MIME types included; unsupported MIME types skipped; path traversal rejected; duplicate filenames skipped

## 3. Docker

- [x] 3.1 Add `ffmpeg` to the `apk add` line in the final stage of `backend/Dockerfile`

## 4. Suggestion pipeline wiring

- [x] 4.1 In `api_items_service.go` (`SuggestTagsImpl` and `Retag`): after loading image assets, if `family.AITaggingUseVideo`, call `ai.LoadVideoKeyframes`; combine with images respecting the `maxImages` cap; pass combined slice to `SuggestTags`
- [x] 4.2 In `pkg/checker/check_untagged.go` (`processItem`): apply the same pattern — load video keyframes when `family.AITaggingUseVideo` and combine with images

## 5. Frontend

- [x] 5.1 Add `aiTaggingUseVideo?: boolean` to the `Family` TypeScript type (generated or hand-written, whichever the project uses)
- [x] 5.2 Add "Include video keyframes" toggle to the profile settings page, displayed below "Include images"; include a privacy note: "Extracted frames are sent to Gemini for analysis"
- [x] 5.3 Wire the toggle to `PATCH /v1/family` with `aiTaggingUseVideo`; enforce client-side that it is only shown (and forcibly unchecked) when AI tagging master switch is on

## 6. Verification

- [x] 6.1 `make build` clean; backend linter clean for changed files
- [x] 6.2 Run backend unit tests: `make test` green
- [x] 6.3 Deploy to diary-wip; verify settings toggle persists; verify graceful degradation (ffmpeg present but flag off → no frames; flag on → frames extracted, log visible)
- [x] 6.4 Full E2E suite green against diary-wip (16/17 expected; pre-existing auth flakiness)
