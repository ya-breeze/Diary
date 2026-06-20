## Why

AI tag suggestions currently use entry text and (optionally) images, but most diary entries contain video files that carry rich contextual information invisible to the model. Phase 3 adds keyframe extraction so videos contribute to tagging without sending raw video blobs to the API.

## What Changes

- Add `ai_tagging_use_video` per-family flag (default off); stored alongside the other AI settings
- Add `ffmpeg` to the backend Docker runtime image for keyframe extraction
- Implement a `pkg/ai/video_loader.go` that shells out to ffmpeg to extract ≈3–5 keyframes per video and returns them as `ImageAsset` slices
- Extend the suggestion flow to load video keyframes the same way image assets are loaded (feeds through the existing multimodal path)
- Degrade gracefully: if ffmpeg is missing or extraction fails, log and continue with text + images only
- Settings UI: add "Include video keyframes" toggle beneath the existing "Include images" toggle, with a privacy note that extracted frames are sent to Gemini
- Tests: ffmpeg unavailable falls back cleanly; extraction wired only when flag on

## Capabilities

### New Capabilities

_(none — video keyframes extend the existing ai-tagging capability, not a new domain)_

### Modified Capabilities

_(none — the `ai-tagging` spec already contains the "Video-based suggestions via keyframes" requirement and `ai_tagging_use_video` config scenario, added during Phase 2. No spec delta needed.)_

## Impact

- **Backend:** new `pkg/ai/video_loader.go`, extend `pkg/ai/image_loader.go` caller (`LoadImageAssets`), new `AITaggingUseVideo` column on `Family`, `SetFamilyAISettings` signature extended, generated server types updated
- **API:** `FamilyResponse` and `FamilySettingsRequest` gain `aiTaggingUseVideo` boolean
- **Docker:** ffmpeg added to `backend/Dockerfile`
- **Frontend:** profile settings page gains a new toggle
- **Dependency:** ffmpeg binary in the container runtime (no Go module changes)
