## Context

`backend/pkg/ai/gemini.go:78` does `return parseSuggestions([]byte(resp.Text()))` after `GenerateContent` returns with no error. `parseSuggestions` (`suggester.go:116`) `json.Unmarshal`s the raw bytes; on `""` it returns `fmt.Errorf("decoding tag suggestions: %w", ...)` → "unexpected end of JSON input". That error bubbles up to the `POST /v1/items/suggest-tags` handler as a 500, and to the background untagged check as a logged ERROR.

The genai SDK returns a successful response with empty `Text()` in several real cases: `finishReason` MAX_TOKENS, SAFETY, RECITATION, or a response with zero candidates. None of these are program errors — they are legitimate "no output" outcomes that should degrade to "no suggestions."

## Goals / Non-Goals

**Goals:**
- Empty/blocked model output → empty suggestion list, no error; endpoint returns 200.
- Retry transient 5xx model-provider errors (bounded, context-aware) before failing.
- Log the finish reason / candidate count / block reason so the empties are diagnosable; log retried attempts.
- Background untagged/backfill check stops logging ERROR for the empty case.

**Non-Goals:**
- Changing request/response shapes or the frontend.
- Non-transient root-cause tuning (MaxOutputTokens, MAX_TOKENS retry, multimodal/video inputs) — deferred until logs show the dominant reason.
- Altering behavior when the model returns valid JSON (parse/normalize path is unchanged).
- Changing the API-key-absent or family-disabled degradation (already handled).

## Decisions

### Decision: Distinguish "empty" from "malformed" in the suggester
Treat a blank/empty raw response as **no suggestions (nil, nil)**; keep returning an error only for a **non-empty but malformed** payload (genuine bug worth surfacing).
- **Why:** Empty is a normal model outcome; malformed JSON is not. Collapsing both to "no error" would hide real integration breakage.
- **Where:** `parseSuggestions` returns `(nil, nil)` when `len(bytes.TrimSpace(raw)) == 0`; `SuggestTags` short-circuits to `(nil, nil)` before parsing when `resp.Text()` is blank.
- **Alternative considered:** Make the HTTP handler swallow the decode error. Rejected — pushes the policy to the wrong layer and would also mask malformed responses.

### Decision: Log the reason at the empty-detection point
When output is empty, log (at WARN/INFO) the `finishReason`, candidate count, and any prompt-feedback/safety block reason from the genai response, plus the family context, before returning empty.
- **Why:** Today there is zero signal on why it's empty; this is the data needed to decide any follow-up tuning.
- **Note:** Inspect the genai response fields for finish/block reason; guard for nil candidates so logging itself never panics.

### Decision: Retry only transient 5xx server errors, with a bounded context-aware backoff
Wrap the `GenerateContent` call in a small retry loop: retry when the returned error is a model-provider **5xx** (detected via the genai SDK's API-error type/`Code`), up to a bounded number of attempts (target ~3 total) with a short backoff between attempts (e.g. ~200ms → ~400ms, small jitter). Abort early if `ctx` is cancelled or its deadline passes; on exhausting attempts, return the wrapped error.
- **Why:** `suggest-tags` is a synchronous, user-facing call — a transient 5xx should be absorbed, but total added latency must stay bounded (~sub-second) so a persistent outage fails fast rather than hanging the request.
- **Scope of "transient":** HTTP 5xx only for now. **429 (rate limit)** is a candidate but has different backoff semantics (respect `Retry-After`) — left as an open question. Non-5xx client errors (4xx) and empty-but-200 responses are **not** retried.
- **Detection:** inspect the genai error for its HTTP status code; if the SDK does not expose a clean status, fall back to a conservative check and treat unknown errors as non-retryable (fail fast) to avoid retry storms.
- **Alternative considered:** a generic retry-any-error wrapper. Rejected — retrying 4xx/config errors or empty responses wastes latency and can mask real problems.

### Decision: Interactive empty result shows an informational UI message
The backend returns 200 + empty list for an empty model response. The **frontend** `fetchSuggestions` (`EntryEditor.tsx`) already awaits the call; on a successful response with zero suggestions it shows a brief informational message (e.g. "No tag suggestions for this entry") via the existing toast infrastructure's non-error variant (`info`/`success`), so the "Suggest tags" button never silently does nothing.
- **Why:** The user explicitly wants feedback for the empty case; it is a normal outcome, so it is informational, not an error toast.
- **Scope:** only the interactive (button) path surfaces this. The background untagged/backfill check treats empty as no-op silently — no message, no error.
- **Alternative considered:** inline text near the tag area. Rejected for now — the toast infra already exists app-wide and keeps this change small; can revisit if a persistent inline hint is preferred.

### Decision: Background check consumes the graceful suggester
`check_untagged.go` already calls `Suggester.SuggestTags`; once that returns `(nil, nil)` on empty, the check naturally treats the day as "no suggestions." Verify no separate ERROR is logged there for the empty case.

## Risks / Trade-offs

- **Hiding a real integration break** (e.g. always-empty due to a bad key/model) → Mitigated by (a) still erroring on non-empty malformed payloads and (b) logging every empty with its reason, so a systemic empty is visible in logs even though it no longer 500s.
- **Nil-deref while reading finish/block reason** → Guard every optional field; treat missing as "unknown reason."

## Open Questions

- Once logs show the dominant `finishReason`, do we follow up with tuning (set `MaxOutputTokens`, retry on MAX_TOKENS, or adjust image/video inputs)? Deferred to a separate change informed by this change's new logs.
- Should **429 (rate limit)** also be retried? It is transient but wants `Retry-After`-aware backoff rather than the fixed 5xx backoff — decide during implementation or defer.
- Exact retry bound and backoff timings — target ~3 attempts / sub-second total; confirm against real Gemini latency during verification.
