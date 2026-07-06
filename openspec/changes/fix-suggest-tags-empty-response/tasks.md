## 1. Suggester robustness

- [ ] 1.1 `parseSuggestions` (`suggester.go`): return `(nil, nil)` when the raw input is empty/whitespace-only; keep returning an error only for non-empty malformed JSON
- [ ] 1.2 `SuggestTags` (`gemini.go`): when `resp.Text()` is blank, return `(nil, nil)` (no suggestions) instead of attempting to parse

## 2. Transient-error retry

- [ ] 2.1 `SuggestTags` (`gemini.go`): wrap `GenerateContent` in a bounded retry loop that retries on a model-provider 5xx (detect via the genai API-error type/`Code`), ~3 attempts with a short backoff (~200ms→~400ms + jitter)
- [ ] 2.2 Abort the retry loop early on `ctx` cancellation/deadline; on exhausting attempts return the wrapped error; do not retry 4xx or empty-but-200 responses
- [ ] 2.3 Add a helper to classify a genai error as retryable (5xx) vs not, guarding for unknown error shapes (treat unknown as non-retryable)

## 3. Observability

- [ ] 3.1 When the model output is empty, log the finish reason, candidate count, and any safety/block reason from the genai response (guard all optional/nil fields), including family context
- [ ] 3.2 Log each retried 5xx attempt (attempt number, status, family context)
- [ ] 3.3 Confirm `SuggestTags` still returns a wrapped error for a non-retryable/persistent API error and for a non-empty malformed payload

## 4. Background check

- [ ] 4.1 `check_untagged.go`: verify that with the graceful suggester the untagged/backfill check treats an empty result as "no suggestions" and logs no ERROR for it; adjust its error handling if it logs on `(nil, nil)`

## 5. Frontend: inform on empty interactive result

- [ ] 5.1 `EntryEditor.tsx` `fetchSuggestions`: when the request succeeds but returns zero suggestions, show an informational (non-error) message via the existing toast infra (e.g. `toast.show('No tag suggestions for this entry', 'info')`)
- [ ] 5.2 Keep genuine failures on the error path (still `toast.error(getErrorMessage(e))`); the empty-result message must not fire when the request actually failed

## 6. Tests

- [ ] 6.1 `backend/pkg/ai`: test that an empty model response yields `(nil, nil)` (no error); malformed non-empty still errors
- [ ] 6.2 `backend/pkg/ai`: test that the empty-response path logs a reason (fake/inspectable logger or reason helper)
- [ ] 6.3 `backend/pkg/ai`: test the retry classifier (5xx retryable, 4xx/unknown not) and that a transient 5xx-then-success returns suggestions while a persistent 5xx returns an error
- [ ] 6.4 `backend/pkg/checker`: test the untagged check tolerates a suggester that returns empty with no error

## 7. Verification

- [ ] 7.1 Run `make lint` and `make test` (or `go test ./...` in `backend`) and fix issues
- [ ] 7.2 Deploy the branch to diary-wip and confirm `POST /v1/items/suggest-tags` returns 200 (empty list) rather than 500 when the model returns nothing; check logs show the reason
- [ ] 7.3 In diary-wip, confirm the editor shows the informational "no suggestions" message when a suggestion request returns empty, and an error toast when the request genuinely fails
