// Package ai provides AI-assisted tag suggestion for diary entries.
//
// The package degrades gracefully: when no GEMINI_API_KEY is configured,
// NewSuggester returns a disabled implementation whose Enabled() reports false
// and whose SuggestTags is a no-op. Callers can therefore wire the suggester
// unconditionally and let configuration decide whether it does anything.
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// TagSuggestion is a single suggested tag with a confidence in [0,1].
type TagSuggestion struct {
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"`
}

// Suggester produces tag suggestions for an entry's text.
type Suggester interface {
	// Enabled reports whether AI suggestion is actually available (API key set).
	Enabled() bool
	// SuggestTags returns suggested tags for the given title/body. knownTags is
	// the family's existing vocabulary; the model is asked to prefer it and coin
	// at most a couple of new tags. Results are de-duplicated and confidence-sorted
	// but NOT filtered against the entry's confirmed tags — keeping pending and
	// confirmed disjoint is the caller's/storage's responsibility. Returns an empty
	// slice (no error) when there is nothing to tag or the suggester is disabled.
	SuggestTags(ctx context.Context, title, body string, knownTags []string) ([]TagSuggestion, error)
}

// disabledSuggester is used when no API key is configured.
type disabledSuggester struct{}

func (disabledSuggester) Enabled() bool { return false }

func (disabledSuggester) SuggestTags(
	_ context.Context, _, _ string, _ []string,
) ([]TagSuggestion, error) {
	return nil, nil
}

// isBlank reports whether the entry has no taggable text.
func isBlank(title, body string) bool {
	return strings.TrimSpace(title) == "" && strings.TrimSpace(body) == ""
}

// buildPrompt assembles the hybrid-vocabulary tagging prompt.
func buildPrompt(title, body string, knownTags []string) string {
	var b strings.Builder
	b.WriteString("You assign short topical tags to a personal diary entry.\n\n")
	b.WriteString("Rules:\n")
	b.WriteString("- Prefer tags from the existing tag list below; reuse an existing tag ")
	b.WriteString("whenever it fits rather than inventing a near-duplicate.\n")
	b.WriteString("- You may introduce at most 2 new tags not in the existing list.\n")
	b.WriteString("- Tags are lowercase, single words or short phrases, no punctuation.\n")
	b.WriteString("- Return 1 to 6 tags. Use a confidence between 0 and 1 for each.\n")
	b.WriteString("- If nothing fits, return an empty list.\n\n")

	if len(knownTags) > 0 {
		b.WriteString("Existing tags: ")
		b.WriteString(strings.Join(knownTags, ", "))
		b.WriteString("\n\n")
	} else {
		b.WriteString("Existing tags: (none yet)\n\n")
	}

	b.WriteString("Entry title: ")
	b.WriteString(title)
	b.WriteString("\n\nEntry body:\n")
	b.WriteString(body)
	return b.String()
}

// parseSuggestions decodes the model's strict-JSON response and normalizes it:
// drops blank names, clamps confidence to [0,1], and de-duplicates by name
// (case-insensitive) keeping the highest confidence, sorted confidence-first.
func parseSuggestions(raw []byte) ([]TagSuggestion, error) {
	var resp struct {
		Tags []TagSuggestion `json:"tags"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("decoding tag suggestions: %w", err)
	}

	best := map[string]TagSuggestion{}
	order := []string{}
	for _, s := range resp.Tags {
		name := strings.TrimSpace(s.Name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		conf := s.Confidence
		if conf < 0 {
			conf = 0
		}
		if conf > 1 {
			conf = 1
		}
		if prev, ok := best[key]; !ok || conf > prev.Confidence {
			if !ok {
				order = append(order, key)
			}
			best[key] = TagSuggestion{Name: name, Confidence: conf}
		}
	}

	out := make([]TagSuggestion, 0, len(order))
	for _, k := range order {
		out = append(out, best[k])
	}
	// Highest confidence first for stable, useful ordering.
	sort.SliceStable(out, func(i, j int) bool { return out[i].Confidence > out[j].Confidence })
	return out, nil
}
