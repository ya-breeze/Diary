## Context

Phase 3 extends the AI tagging pipeline (established in Phases 1–2) to include video assets. The existing `LoadImageAssets` function (`pkg/ai/image_loader.go`) reads image files referenced in an entry body and returns `[]ImageAsset` for the Gemini multimodal call. Video files are currently silently skipped because their MIME type is not in `supportedImageMIMEs`. Phase 3 adds a parallel `LoadVideoKeyframes` function that shells out to ffmpeg to extract still frames from referenced videos, returning them as `[]ImageAsset` (JPEG), so they flow through the same multimodal path without any changes to the `Suggester` interface.

## Goals / Non-Goals

**Goals:**
- Extract ≈3–5 keyframes per video using ffmpeg fixed-interval sampling
- Feed frames through the existing `[]ImageAsset` path — no Suggester API change
- Degrade gracefully (ffmpeg missing, extraction error, unsupported format → skip silently)
- Add `ai_tagging_use_video` per-family flag (DB, API, UI, settings handler)
- Add ffmpeg to the backend Docker runtime image

**Non-Goals:**
- Native video upload to Gemini (against spec requirement)
- Scene-change detection (complex, unreliable for short clips; fixed-interval is good enough)
- Storing extracted frames on disk (always temp, always deleted after the request)
- Audio processing

## Decisions

### D1: Fixed-interval sampling, not scene-change detection

**Chosen:** `ffmpeg -i <input> -vf fps=1/<step> -frames:v 5 <tmpdir>/frame%03d.jpg`
where step = max(1, duration_seconds/5) so short videos still yield frames.

**Alternative considered:** `-vf "select='gt(scene,0.4)'"` — scene-change filter is unreliable (high threshold misses changes in slow-moving scenes, low threshold floods the frame count). Fixed intervals are predictable and sufficient for topical tagging where we just need a sample of what's in the video.

### D2: Shell-out to ffmpeg, not a Go binding

**Chosen:** `os/exec.CommandContext`. ffmpeg is already being added to the runtime image; no additional Go module dependency needed. The extraction is infrequent (at most one call per video per suggestion), so process-spawn overhead is negligible.

**Alternative considered:** `github.com/u2takey/ffmpeg-go` — adds a module dependency, uses the same shell-out internally. Not worth the indirection.

### D3: Temp-dir output, not stdout pipe

**Chosen:** Extract frames as files to `os.MkdirTemp`, read them, then `os.RemoveAll` in a deferred call.

**Alternative considered:** `-f image2pipe -vcodec mjpeg -` to stdout — frames arrive concatenated without delimiters; splitting JPEG streams reliably requires scanning for SOI/EOI markers. Temp files are simpler and the overhead is negligible (frames are small JPEGs).

### D4: `LoadVideoKeyframes(body, dataPath, familyID string) []ImageAsset` — separate function

Keeps `image_loader.go` focused on static images. Callers combine the slices and truncate to the global `maxImages` cap.

**Combined cap rule in callers:**
```go
images := ai.LoadImageAssets(body, cfg.DataPath, familyID)
if family.AITaggingUseVideo {
    frames := ai.LoadVideoKeyframes(body, cfg.DataPath, familyID)
    if room := maxImages - len(images); room > 0 {
        if len(frames) > room { frames = frames[:room] }
        images = append(images, frames...)
    }
}
```
The `maxImages = 10` constant in `image_loader.go` will be moved to a shared location (or duplicated in `video_loader.go` as `maxTotalAssets`).

### D5: Detect video by MIME type, not extension

Read the first 512 bytes of each asset, call `http.DetectContentType`. Only process files detected as `video/*`. This reuses the same pattern as image detection and avoids relying on file extensions (assets are stored as UUID filenames with no extension).

**Supported MIME types for keyframe extraction:** `video/mp4`, `video/quicktime`, `video/x-msvideo`, `video/webm`, `video/x-matroska`.

Note: `http.DetectContentType` for video types relies on magic bytes; MP4 and MOV detection is reliable; WebM is detected as `video/webm`. We won't pre-check ffmpeg can decode the specific codec — unsupported codecs will cause ffmpeg to exit non-zero, which is already the graceful-degrade path.

### D6: `SetFamilyAISettings` signature extended with `useVideo bool`

Add as the 5th parameter. All callers updated. The mock and generated code regenerated.

## Risks / Trade-offs

- **ffmpeg process overhead** → Acceptable: extraction is bounded to at most one ffmpeg call per video per suggestion. `maxFramesPerVideo = 5`. A 200 MB MP4 won't be fully read by our code (we only read 512 bytes for MIME detection); ffmpeg reads it directly.
- **Large videos slow down suggestion** → Mitigation: `CommandContext` with a reasonable timeout (30 seconds). If exceeded, extraction is skipped and suggestion continues with text + images.
- **ffmpeg unavailable in some environments** → Fully graceful: `exec.LookPath("ffmpeg")` check at extraction time; if not found, log once and return nil. No panic, no error surfaced to the caller.
- **Alpine `apk` image size** → `ffmpeg` on Alpine is ~30 MB. Acceptable for this use case.
- **MIME detection false negatives** → UUID filenames have no extension; `http.DetectContentType` reads magic bytes, so it's reliable for common video formats.

## Migration Plan

1. Schema migration is additive: `AITaggingUseVideo bool gorm:"default:false"` added via GORM AutoMigrate. Existing rows get `false`. No manual migration.
2. Docker image: add `ffmpeg` to `apk add` in `backend/Dockerfile`. WIP stack redeploy picks it up.
3. API is additive: `aiTaggingUseVideo` optional field in `FamilySettingsRequest` (already `*bool` pattern); `FamilyResponse` gains the field. Clients that don't know the field are unaffected.

## Open Questions

- _(none)_
