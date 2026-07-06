## ADDED Requirements

### Requirement: Empty model responses yield no suggestions and inform the user

When AI tagging is enabled and a suggestion is requested, an empty, blank, or non-parseable-because-empty model response SHALL be treated as a successful "no suggestions" outcome — not a server error and not a retry. A model that returns no usable content (including responses stopped for token limits or blocked for safety/recitation, or responses with no candidate) MUST NOT cause the suggestion request to fail. The suggestion endpoint SHALL return success with an empty list in this case.

For an **interactive** (user-triggered) suggestion request that yields no suggestions, the UI SHALL show a brief informational message telling the user that no suggestions were produced (rather than appearing to do nothing). This message is informational, not an error. The **background** untagged/backfill health-check paths SHALL treat an empty result as "no suggestions" silently (no user message, no error log).

When a response yields no usable content, the backend SHALL log the reason (the model's finish reason, candidate count, and any safety/block reason where available) to support diagnosis.

#### Scenario: Interactive request with an empty model response informs the user

- **GIVEN** a family with AI tagging enabled and a configured API key
- **WHEN** the user requests suggestions from the editor and the model returns an empty or blank response
- **THEN** the endpoint returns success with an empty list (no server error, no retry)
- **AND** the UI shows an informational message that no suggestions were produced
- **AND** the backend logs the reason for the empty response

#### Scenario: Model response is blocked or truncated

- **GIVEN** a family with AI tagging enabled and a configured API key
- **WHEN** the model stops without usable output (e.g. token limit reached, content blocked, or no candidate returned)
- **THEN** an empty suggestion list is returned without error and without retry
- **AND** the finish/block reason is logged

#### Scenario: Background untagged check tolerates empty suggestions silently

- **WHEN** the background untagged/backfill health check requests suggestions and the model returns no usable content
- **THEN** the check treats the day as having no suggestions
- **AND** no user message is shown and no error is logged for the empty result

### Requirement: Transient model-provider errors are retried

When a suggestion request fails with a transient error from the model provider — an HTTP **5xx** server error or an HTTP **429** rate-limit error — the system SHALL retry the request a bounded number of times before failing, and SHALL respect the caller's context (cancellation/deadline). For a 429, the retry SHALL honor the provider's `Retry-After` hint when present (bounded so it cannot exceed the retry budget), otherwise it SHALL use the standard short backoff used for 5xx. Each retried attempt SHALL be logged. If all attempts fail, the request SHALL surface the error (rather than silently returning empty). Retry applies only to these transient errors — it SHALL NOT retry a non-transient (other 4xx) error or an empty-but-successful response (which is handled by degrading to no suggestions).

#### Scenario: Transient 5xx is retried and then succeeds

- **GIVEN** a family with AI tagging enabled and a configured API key
- **WHEN** the model call first returns a 5xx error and a subsequent attempt succeeds
- **THEN** suggestions from the successful attempt are returned
- **AND** the retried attempt(s) are logged

#### Scenario: Rate-limit (429) is retried, honoring Retry-After

- **GIVEN** a family with AI tagging enabled and a configured API key
- **WHEN** the model call returns a 429 with a `Retry-After` hint and a later attempt succeeds
- **THEN** the retry waits according to the `Retry-After` hint (bounded by the retry budget) before re-attempting
- **AND** suggestions from the successful attempt are returned

#### Scenario: Persistent transient error surfaces an error

- **WHEN** every attempt (up to the retry bound) returns a 5xx or 429 error
- **THEN** the suggestion request fails with an error rather than returning an empty list
- **AND** the failed attempts are logged

#### Scenario: Non-transient error is not retried

- **WHEN** the model call returns a non-transient error (a 4xx other than 429)
- **THEN** the request fails immediately without retrying

#### Scenario: Empty successful response is not retried

- **WHEN** the model returns a successful response with no usable content
- **THEN** the system returns an empty suggestion list without retrying the call
