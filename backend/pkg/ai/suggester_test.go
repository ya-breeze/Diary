package ai

import (
	"context"
	"testing"
)

func TestDisabledSuggester(t *testing.T) {
	var s Suggester = disabledSuggester{}
	if s.Enabled() {
		t.Fatal("disabled suggester should report Enabled()=false")
	}
	got, err := s.SuggestTags(context.Background(), "Title", "Body", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected no suggestions, got %v", got)
	}
}

func TestIsBlank(t *testing.T) {
	cases := []struct {
		title, body string
		want        bool
	}{
		{"", "", true},
		{"  ", "\n\t", true},
		{"Title", "", false},
		{"", "Body", false},
	}
	for _, c := range cases {
		if got := isBlank(c.title, c.body); got != c.want {
			t.Errorf("isBlank(%q,%q)=%v want %v", c.title, c.body, got, c.want)
		}
	}
}

func TestBuildPromptIncludesKnownTags(t *testing.T) {
	p := buildPrompt("My day", "went to the beach", []string{"travel", "family"})
	for _, want := range []string{"travel", "family", "My day", "went to the beach", "at most 2 new"} {
		if !contains(p, want) {
			t.Errorf("prompt missing %q\n---\n%s", want, p)
		}
	}
}

func TestBuildPromptNoKnownTags(t *testing.T) {
	p := buildPrompt("t", "b", nil)
	if !contains(p, "(none yet)") {
		t.Errorf("expected '(none yet)' when no known tags:\n%s", p)
	}
}

func TestParseSuggestions(t *testing.T) {
	raw := []byte(`{"tags":[
		{"name":"beach","confidence":0.9},
		{"name":" ","confidence":0.5},
		{"name":"Beach","confidence":0.4},
		{"name":"family","confidence":1.7},
		{"name":"work","confidence":-0.2}
	]}`)
	got, err := parseSuggestions(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// blank dropped; "beach"/"Beach" de-duped to highest (0.9); confidence clamped.
	if len(got) != 3 {
		t.Fatalf("expected 3 suggestions, got %d: %+v", len(got), got)
	}
	if got[0].Name != "family" || got[0].Confidence != 1 {
		t.Errorf("expected family clamped to 1 first, got %+v", got[0])
	}
	// confidence-sorted descending: family(1) > beach(0.9) > work(0)
	if got[1].Name != "beach" || got[1].Confidence != 0.9 {
		t.Errorf("expected beach 0.9 second, got %+v", got[1])
	}
	if got[2].Name != "work" || got[2].Confidence != 0 {
		t.Errorf("expected work clamped to 0 last, got %+v", got[2])
	}
}

func TestParseSuggestionsInvalidJSON(t *testing.T) {
	if _, err := parseSuggestions([]byte("not json")); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func contains(haystack, needle string) bool {
	return len(needle) == 0 || (len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
