package ai

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"google.golang.org/genai"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestParseSuggestionsEmpty(t *testing.T) {
	for _, raw := range [][]byte{nil, []byte(""), []byte("   \n\t")} {
		got, err := parseSuggestions(raw)
		if err != nil {
			t.Fatalf("parseSuggestions(%q) unexpected error: %v", raw, err)
		}
		if len(got) != 0 {
			t.Fatalf("parseSuggestions(%q) expected no suggestions, got %v", raw, got)
		}
	}
}

func TestLogEmptyResponseLogsReason(t *testing.T) {
	var buf bytes.Buffer
	g := &geminiSuggester{logger: slog.New(slog.NewTextHandler(&buf, nil))}

	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{{FinishReason: genai.FinishReasonMaxTokens}},
	}
	g.logEmptyResponse(resp)

	out := buf.String()
	for _, want := range []string{"finishReason=MAX_TOKENS", "candidateCount=1"} {
		if !strings.Contains(out, want) {
			t.Errorf("empty-response log missing %q\n%s", want, out)
		}
	}

	// Must not panic on a nil response / empty candidates.
	g.logEmptyResponse(nil)
	g.logEmptyResponse(&genai.GenerateContentResponse{})
}

func TestIsRetryable(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want bool
	}{
		{"500", genai.APIError{Code: 500}, true},
		{"503", genai.APIError{Code: 503}, true},
		{"599", genai.APIError{Code: 599}, true},
		{"429", genai.APIError{Code: 429}, true},
		{"400", genai.APIError{Code: 400}, false},
		{"404", genai.APIError{Code: 404}, false},
		{"wrapped 502", fmt.Errorf("gemini: %w", genai.APIError{Code: 502}), true},
		{"non-api error", errors.New("boom"), false},
		{"nil", nil, false},
	}
	for _, c := range cases {
		if got := isRetryable(c.err); got != c.want {
			t.Errorf("isRetryable(%s)=%v want %v", c.name, got, c.want)
		}
	}
}

func TestRetryAfter(t *testing.T) {
	err := genai.APIError{Code: 429, Details: []map[string]any{
		{"@type": "type.googleapis.com/google.rpc.RetryInfo", "retryDelay": "2s"},
	}}
	d, ok := retryAfter(err)
	if !ok || d != 2*time.Second {
		t.Fatalf("retryAfter got (%v,%v) want (2s,true)", d, ok)
	}

	if _, ok := retryAfter(genai.APIError{Code: 429}); ok {
		t.Error("retryAfter with no details should return ok=false")
	}
	if _, ok := retryAfter(errors.New("boom")); ok {
		t.Error("retryAfter on non-api error should return ok=false")
	}
}

func TestBackoffForRetryAfterBudget(t *testing.T) {
	// Within budget: honored.
	within := genai.APIError{Code: 429, Details: []map[string]any{{"retryDelay": "500ms"}}}
	if d, ok := backoffFor(1, within); !ok || d != 500*time.Millisecond {
		t.Errorf("backoffFor within budget got (%v,%v) want (500ms,true)", d, ok)
	}
	// Exceeds budget: fail fast.
	tooLong := genai.APIError{Code: 429, Details: []map[string]any{{"retryDelay": "60s"}}}
	if _, ok := backoffFor(1, tooLong); ok {
		t.Error("backoffFor should return ok=false when Retry-After exceeds maxRetryWait")
	}
	// No hint: exponential, capped at maxRetryWait.
	d, ok := backoffFor(1, genai.APIError{Code: 500})
	if !ok || d <= 0 || d > maxRetryWait {
		t.Errorf("backoffFor no-hint got (%v,%v) want a positive duration <= %v", d, ok, maxRetryWait)
	}
}

func TestRetryTransient(t *testing.T) {
	okResp := &genai.GenerateContentResponse{}

	// failThenOK returns a transient error for the first (failFor) calls, then ok.
	failThenOK := func(code, failFor int) func() (*genai.GenerateContentResponse, error) {
		n := 0
		return func() (*genai.GenerateContentResponse, error) {
			n++
			if n <= failFor {
				return nil, genai.APIError{Code: code}
			}
			return okResp, nil
		}
	}

	cases := []struct {
		name      string
		gen       func() (*genai.GenerateContentResponse, error)
		wantOK    bool
		wantCalls int
	}{
		{"success first attempt", failThenOK(0, 0), true, 1},
		{"transient then success", failThenOK(503, 1), true, 2},
		{"persistent transient surfaces error", failThenOK(500, maxAttempts), false, maxAttempts},
		{"non-transient not retried", failThenOK(400, maxAttempts), false, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			calls := 0
			gen := func() (*genai.GenerateContentResponse, error) {
				calls++
				return c.gen()
			}
			resp, err := retryTransient(context.Background(), discardLogger(), gen)
			gotOK := err == nil && resp == okResp
			if gotOK != c.wantOK || calls != c.wantCalls {
				t.Fatalf("gotOK=%v calls=%d; want ok=%v calls=%d (err=%v)",
					gotOK, calls, c.wantOK, c.wantCalls, err)
			}
		})
	}
}
