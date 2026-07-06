## Why

`POST /v1/items/suggest-tags` returns **HTTP 500** whenever Gemini responds with an empty body. `gemini.go` calls `parseSuggestions([]byte(resp.Text()))` after `GenerateContent` returns *without* an error; when the model produces no text (e.g. `finishReason` MAX_TOKENS / SAFETY / RECITATION, or no candidate), `resp.Text()` is `""` and `json.Unmarshal("")` fails with *"unexpected end of JSON input"*. Prod logs show this 500 on essentially every editing session, and the background "Untagged check" health check erroring across ~256 dates. A legitimate "the model had nothing / was blocked" outcome is being turned into a hard error, and there is currently **no logging of why** the response was empty, so it cannot be diagnosed.

## What Changes

- Treat an empty or unparseable-because-empty model response as **no suggestions** (empty list, no error, **no retry**). The endpoint returns **200 with an empty list** instead of 500.
- For an **interactive** (user-triggered) request that comes back empty, the **frontend shows a brief informational message** ("no suggestions produced") so the user gets feedback instead of the button appearing to do nothing. The background health check stays silent on empty.
- **Retry the model call on transient errors** from the Gemini API — 5xx and 429 (bounded attempts with short backoff, respecting the request context; a 429's `Retry-After` is honored but capped by the retry budget) — before giving up. Persisting failures still surface as an error (which the frontend now toasts). Empty 2xx responses and other 4xx are **not** retried.
- Add **observability**: when a response yields no usable content, log the `finishReason`, candidate count, and any safety/block reason; log each retried 5xx attempt so transient flakiness is visible.
- Apply the same graceful handling on the **background untagged/backfill health-check path** so it stops emitting ERROR logs for the empty case.
- Deliberately **out of scope** (deferred until the new logs reveal the dominant `finishReason`): non-transient root-cause tuning such as setting `MaxOutputTokens`, retrying on MAX_TOKENS, or changing multimodal/video-frame inputs. Retry here targets transient **server** errors, not empty-but-successful responses. Captured as open questions in design.

## Capabilities

### New Capabilities
<!-- none -->

### Modified Capabilities
- `ai-tagging`: extends graceful-degradation behavior — an empty or blocked model response SHALL degrade to "no suggestions" (with an informational UI message on the interactive path) rather than surfacing an error, the reason SHALL be logged, and transient 5xx/429 errors from the model provider SHALL be retried (bounded, `Retry-After`-aware for 429) before the request fails.

## Impact

- **Modified code**: `backend/pkg/ai/gemini.go` (empty-response handling, transient-5xx retry, and logging in `SuggestTags`), `backend/pkg/ai/suggester.go` (`parseSuggestions` tolerates empty input), `backend/pkg/checker/check_untagged.go` (graceful handling of empty suggestions).
- **API behavior**: `POST /v1/items/suggest-tags` returns 200 + empty list instead of 500 for empty model output. No request/response shape change.
- **Frontend**: `next-frontend/src/components/diary/EntryEditor.tsx` `fetchSuggestions` — when an interactive suggestion request returns zero suggestions, show an informational message (reusing the existing toast infra's non-error variant). Genuine failures still toast as errors (merged `error-feedback`).
- **Tests**: `backend/pkg/ai` (empty/blocked response → empty, no error) and `backend/pkg/checker` (untagged check tolerates empty suggestions).
