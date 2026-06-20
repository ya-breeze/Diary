package ai

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/ya-breeze/diary.be/pkg/config"
	"github.com/ya-breeze/diary.be/pkg/utils"
)

var supportedVideoMIMEs = map[string]bool{
	"video/mp4":        true,
	"video/quicktime":  true,
	"video/x-msvideo": true,
	"video/webm":       true,
	"video/x-matroska": true,
}

const maxFramesPerVideo = 5
const defaultFrameStep = 10 // seconds between frames when duration is unknown
const extractionTimeout = 30 * time.Second
const maxVideoFileSizeBytes = 500 << 20 // 500 MB

// LoadVideoKeyframes reads asset filenames from body, finds video files, extracts
// keyframes via ffmpeg, and returns at most budget ImageAssets (JPEG). Returns nil
// when ffmpeg is unavailable or no video assets are present.
func LoadVideoKeyframes(body, dataPath, familyID string, logger *slog.Logger, budget int) []ImageAsset {
	ffmpegPath, err := exec.LookPath("ffmpeg")
	if err != nil {
		logger.Warn("ffmpeg not found; video keyframe extraction skipped")
		return nil
	}

	names := utils.GetAssetsFromMarkdown(body)
	if len(names) == 0 {
		return nil
	}

	assetsDir := filepath.Join(dataPath, config.AssetsDirName, familyID)
	seen := make(map[string]struct{}, len(names))
	var frames []ImageAsset

	for _, name := range names {
		if len(frames) >= budget {
			break
		}
		clean := filepath.Clean(name)
		if strings.Contains(clean, "..") {
			continue
		}
		if _, dup := seen[clean]; dup {
			continue
		}
		seen[clean] = struct{}{}

		path := filepath.Join(assetsDir, clean)

		info, statErr := os.Stat(path)
		if statErr != nil || info.Size() > maxVideoFileSizeBytes {
			if statErr == nil {
				logger.Warn("Video file too large for keyframe extraction", "path", path, "size", info.Size())
			}
			continue
		}

		if !isVideoFile(path) {
			continue
		}

		extracted, err := extractKeyframes(path, ffmpegPath)
		if err != nil {
			logger.Warn("Video keyframe extraction failed", "path", path, "error", err)
			continue
		}
		remaining := budget - len(frames)
		if len(extracted) > remaining {
			extracted = extracted[:remaining]
		}
		frames = append(frames, extracted...)
	}
	return frames
}

func isVideoFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := io.ReadFull(f, buf)
	if err != nil && n == 0 {
		return false
	}
	mime := http.DetectContentType(buf[:n])
	if i := strings.IndexByte(mime, ';'); i >= 0 {
		mime = strings.TrimSpace(mime[:i])
	}
	return supportedVideoMIMEs[mime]
}

func extractKeyframes(path, ffmpegPath string) ([]ImageAsset, error) {
	step := videoDurationStep(path, ffmpegPath)

	tmpDir, err := os.MkdirTemp("", "diary-video-*")
	if err != nil {
		return nil, fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	outPattern := filepath.Join(tmpDir, "frame%03d.jpg")
	ctx, cancel := context.WithTimeout(context.Background(), extractionTimeout)
	defer cancel()

	//nolint:gosec // path is validated by isVideoFile before reaching here
	cmd := exec.CommandContext(ctx, ffmpegPath,
		"-i", path,
		"-vf", fmt.Sprintf("fps=1/%d", step),
		"-frames:v", strconv.Itoa(maxFramesPerVideo),
		"-q:v", "3",
		"-y",
		outPattern,
	)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg: %w", err)
	}

	entries, err := os.ReadDir(tmpDir)
	if err != nil {
		return nil, fmt.Errorf("reading frame dir: %w", err)
	}

	frames := make([]ImageAsset, 0, len(entries))
	for _, e := range entries {
		data, err := os.ReadFile(filepath.Join(tmpDir, e.Name()))
		if err != nil {
			continue
		}
		frames = append(frames, ImageAsset{MIMEType: "image/jpeg", Data: data})
	}
	return frames, nil
}

// videoDurationStep probes the video duration and returns a frame interval (in
// seconds) that yields approximately maxFramesPerVideo evenly spaced frames.
// Falls back to defaultFrameStep if the duration cannot be determined.
func videoDurationStep(path, ffmpegPath string) int {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// -i alone reads the container header and exits 1 (no output file specified)
	// without decoding any frames — much lighter than -f null -.
	//nolint:gosec // path validated before this call
	out, _ := exec.CommandContext(ctx, ffmpegPath, "-i", path).CombinedOutput()
	if len(out) == 0 {
		return defaultFrameStep
	}

	// Parse "Duration: H:MM:SS.ss" from ffmpeg stderr.
	outStr := string(out)
	idx := strings.Index(outStr, "Duration: ")
	if idx < 0 {
		return defaultFrameStep
	}
	durStr := outStr[idx+len("Duration: "):]

	// Trim fractional seconds before splitting so single-digit hours don't
	// push the decimal point inside the parsed fields.
	if i := strings.IndexByte(durStr, '.'); i >= 0 {
		durStr = durStr[:i]
	}
	parts := strings.Split(durStr, ":")
	if len(parts) != 3 {
		return defaultFrameStep
	}
	h, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
	m, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
	s, err3 := strconv.Atoi(strings.TrimSpace(parts[2]))
	if err1 != nil || err2 != nil || err3 != nil {
		return defaultFrameStep
	}
	totalSecs := h*3600 + m*60 + s
	if totalSecs <= 0 {
		return defaultFrameStep
	}
	step := totalSecs / maxFramesPerVideo
	if step < 1 {
		return 1
	}
	return step
}
