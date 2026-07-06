package ai

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math/rand"
	"os"
	"strings"
	"time"

	"google.golang.org/genai"
)

// defaultModel is a low-cost flash-lite model — ample for short tagging tasks.
// Pinned (not a -latest alias) for predictable cost/behavior; bump deliberately.
const defaultModel = "gemini-2.5-flash-lite"

// geminiSuggester is the live, API-backed Suggester.
type geminiSuggester struct {
	client *genai.Client
	model  string
	logger *slog.Logger
}

// NewSuggester builds a Suggester. If GEMINI_API_KEY is unset it returns a
// disabled suggester (no error) so callers can wire it unconditionally.
func NewSuggester(ctx context.Context, logger *slog.Logger) (Suggester, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		logger.Info("AI tagging disabled: GEMINI_API_KEY not set")
		return disabledSuggester{}, nil
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, fmt.Errorf("creating gemini client: %w", err)
	}

	return &geminiSuggester{client: client, model: defaultModel, logger: logger}, nil
}

func (g *geminiSuggester) Enabled() bool { return true }

// maxBodyChars caps how much entry text is sent to the model — tags come from
// the gist, so the whole of a very long entry isn't needed (bounds token cost).
const maxBodyChars = 8000

func (g *geminiSuggester) SuggestTags(
	ctx context.Context, title, body string, images []ImageAsset, knownTags []string,
) ([]TagSuggestion, error) {
	if isBlank(title, body) {
		return nil, nil
	}

	if len(body) > maxBodyChars {
		body = body[:maxBodyChars]
	}
	prompt := buildPrompt(title, body, knownTags)

	parts := make([]*genai.Part, 0, 1+len(images))
	parts = append(parts, &genai.Part{Text: prompt})
	for _, img := range images {
		parts = append(parts, genai.NewPartFromBytes(img.Data, img.MIMEType))
	}
	contents := []*genai.Content{{Role: genai.RoleUser, Parts: parts}}

	resp, err := g.generateWithRetry(ctx, contents)
	if err != nil {
		return nil, fmt.Errorf("gemini tag suggestion: %w", err)
	}

	// An empty response is a legitimate "no suggestions" outcome (blocked,
	// token-limited, or no candidate) — log the reason and degrade gracefully
	// rather than failing. parseSuggestions also tolerates empty input.
	text := resp.Text()
	if strings.TrimSpace(text) == "" {
		g.logEmptyResponse(resp)
		return nil, nil
	}

	return parseSuggestions([]byte(text))
}

const (
	// maxAttempts bounds total tries (1 initial + retries) for a transient failure.
	maxAttempts = 3
	// baseBackoff is the first inter-attempt wait; it grows ~exponentially.
	baseBackoff = 200 * time.Millisecond
	// maxRetryWait caps a single wait (including a server Retry-After hint) so a
	// synchronous suggest-tags request cannot hang; longer hints fail fast.
	maxRetryWait = 1 * time.Second
)

// generateWithRetry calls the model, retrying transient failures.
func (g *geminiSuggester) generateWithRetry(
	ctx context.Context, contents []*genai.Content,
) (*genai.GenerateContentResponse, error) {
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		ResponseSchema:   tagSuggestionSchema(),
	}
	return retryTransient(ctx, g.logger, func() (*genai.GenerateContentResponse, error) {
		return g.client.Models.GenerateContent(ctx, g.model, contents, config)
	})
}

// retryTransient runs gen, retrying transient (5xx/429) failures a bounded
// number of times with a short, context-aware backoff. A 429's Retry-After hint
// is honored but capped by maxRetryWait; a hint that exceeds the cap fails fast.
// Non-transient errors, an exhausted budget, or the final attempt return the
// error for the caller to wrap.
func retryTransient(
	ctx context.Context,
	logger *slog.Logger,
	gen func() (*genai.GenerateContentResponse, error),
) (*genai.GenerateContentResponse, error) {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, err := gen()
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if !isRetryable(err) || attempt == maxAttempts {
			return nil, err
		}
		wait, ok := backoffFor(attempt, err)
		if !ok {
			return nil, err // Retry-After exceeds our budget — fail fast.
		}
		logger.Warn("gemini call failed; retrying",
			"attempt", attempt, "maxAttempts", maxAttempts, "wait", wait.String(), "error", err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(wait):
		}
	}
	return nil, lastErr
}

// isRetryable reports whether a model-provider error is transient: an HTTP 5xx
// server error or a 429 rate-limit. Unknown error shapes are treated as
// non-retryable to avoid retry storms.
func isRetryable(err error) bool {
	var apiErr genai.APIError
	if errors.As(err, &apiErr) {
		return apiErr.Code == 429 || (apiErr.Code >= 500 && apiErr.Code <= 599)
	}
	return false
}

// backoffFor returns how long to wait before the next attempt and whether it is
// within budget. It prefers a server-provided Retry-After hint (Google RPC
// RetryInfo); otherwise it uses exponential backoff with jitter. A wait longer
// than maxRetryWait returns ok=false so the caller fails fast.
func backoffFor(attempt int, err error) (time.Duration, bool) {
	if d, ok := retryAfter(err); ok {
		if d > maxRetryWait {
			return 0, false
		}
		return d, true
	}
	d := baseBackoff * time.Duration(int64(1)<<(attempt-1))
	//nolint:gosec // jitter for retry backoff is not security-sensitive
	d += time.Duration(rand.Int63n(int64(baseBackoff / 2)))
	if d > maxRetryWait {
		d = maxRetryWait
	}
	return d, true
}

// retryAfter extracts a server-provided retry delay from a transient error, if
// the provider included Google RPC RetryInfo details ({"retryDelay": "5s"}).
func retryAfter(err error) (time.Duration, bool) {
	var apiErr genai.APIError
	if !errors.As(err, &apiErr) {
		return 0, false
	}
	for _, detail := range apiErr.Details {
		if delay, ok := detail["retryDelay"].(string); ok {
			if d, perr := time.ParseDuration(delay); perr == nil && d > 0 {
				return d, true
			}
		}
	}
	return 0, false
}

// logEmptyResponse records why a model response carried no usable content so the
// empties are diagnosable. All optional fields are nil-guarded.
func (g *geminiSuggester) logEmptyResponse(resp *genai.GenerateContentResponse) {
	finishReason := ""
	candidateCount := 0
	if resp != nil {
		candidateCount = len(resp.Candidates)
		if candidateCount > 0 && resp.Candidates[0] != nil {
			finishReason = string(resp.Candidates[0].FinishReason)
		}
	}
	blockReason := ""
	if resp != nil && resp.PromptFeedback != nil {
		blockReason = string(resp.PromptFeedback.BlockReason)
	}
	g.logger.Warn("gemini returned no usable tag suggestions",
		"finishReason", finishReason,
		"candidateCount", candidateCount,
		"blockReason", blockReason,
	)
}

// tagSuggestionSchema is the strict response schema: {tags:[{name,confidence}]}.
func tagSuggestionSchema() *genai.Schema {
	return &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"tags": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"name":       {Type: genai.TypeString, Description: "short topical tag"},
						"confidence": {Type: genai.TypeNumber, Description: "confidence 0..1"},
					},
					Required: []string{"name", "confidence"},
				},
			},
		},
		Required: []string{"tags"},
	}
}
