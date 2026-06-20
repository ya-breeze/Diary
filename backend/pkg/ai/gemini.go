package ai

import (
	"context"
	"fmt"
	"log/slog"
	"os"

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

	resp, err := g.client.Models.GenerateContent(
		ctx,
		g.model,
		contents,
		&genai.GenerateContentConfig{
			ResponseMIMEType: "application/json",
			ResponseSchema:   tagSuggestionSchema(),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("gemini tag suggestion: %w", err)
	}

	return parseSuggestions([]byte(resp.Text()))
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
