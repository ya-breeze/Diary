## ADDED Requirements

### Requirement: Failed user-initiated actions surface an error to the user

The frontend SHALL display a human-readable error message to the user whenever an action the user explicitly initiated fails. A user-initiated action is one triggered by a direct user gesture, including: uploading images, saving a diary entry (whether saving in place or saving before switching to another date), switching to another date in the editor, requesting tag suggestions, accepting a suggested tag, dismissing a suggested tag, updating the AI tagging setting, and logging out. Such failures MUST NOT be swallowed into logging only.

Exception: a failure whose handling is the session-refresh/redirect flow (an expired session that results in a redirect to login) is considered handled by that redirect and SHALL NOT additionally raise a user-facing error notification, even when it occurred during a user-initiated action.

#### Scenario: Image upload fails

- **WHEN** the user uploads one or more images and the upload request fails
- **THEN** the app displays an error notification containing a human-readable message
- **AND** the upload progress indicator is cleared

#### Scenario: Saving an entry fails

- **WHEN** the user saves a diary entry and the save request fails
- **THEN** the app displays an error notification containing a human-readable message
- **AND** the user remains in the editor with their content intact

#### Scenario: Saving before switching to another date fails

- **WHEN** the user switches to another date while the entry has unsaved changes and chooses to save first, and that save fails
- **THEN** the app displays an error notification containing a human-readable message
- **AND** the editor stays on the current date with the unsaved content intact (the date switch does not occur)

#### Scenario: Loading an entry after switching dates fails

- **WHEN** the user switches to another date in the editor and loading that date's entry fails
- **THEN** the app displays an error notification containing a human-readable message

#### Scenario: Tag suggestion request fails

- **WHEN** the user requests tag suggestions and the request fails
- **THEN** the app displays an error notification containing a human-readable message

#### Scenario: Accepting or dismissing a suggested tag fails

- **WHEN** the user accepts or dismisses a suggested tag and the request fails
- **THEN** the app displays an error notification containing a human-readable message

#### Scenario: Updating the AI tagging setting fails

- **WHEN** the user changes the AI tagging setting and the request fails
- **THEN** the app displays an error notification containing a human-readable message

#### Scenario: Logout fails

- **WHEN** the user logs out and the request fails
- **THEN** the app displays an error notification containing a human-readable message

### Requirement: Error messages are normalized to a readable string

The frontend SHALL normalize the different error shapes produced across the app — the `ApiError` thrown by the shared API client, a plain `Error` (e.g. the raw XHR rejection from the batch asset upload), and unknown/thrown values — into a single human-readable message used for user-facing notifications. The produced message SHALL be safe and user-friendly: it MUST NOT expose raw backend response bodies, stack traces, or internal implementation details. Where an error's raw text is not suitable for display, the normalizer SHALL substitute a friendly, status-aware message.

#### Scenario: ApiError is normalized

- **WHEN** an action fails with an `ApiError` carrying a status and message
- **THEN** the notification shows a safe, user-friendly message derived from that error rather than a generic placeholder
- **AND** the message does not expose a raw backend response body, stack trace, or internal details

#### Scenario: Non-ApiError failure is normalized

- **WHEN** an action fails with a plain `Error` or a non-error thrown value
- **THEN** the normalizer still produces a readable message and the app displays a notification without throwing

### Requirement: Error notifications are perceivable to assistive technology

Error notifications SHALL be announced to assistive technology so screen-reader users perceive them. The toast container SHALL use an appropriate live region (e.g. `aria-live="assertive"` / `role="alert"`) so newly shown error messages are announced without requiring focus.

#### Scenario: Error toast is announced to screen readers

- **WHEN** an error notification appears
- **THEN** it is rendered within a live region that causes its message to be announced to assistive technology

### Requirement: Background enhancements degrade silently

Optional, non-user-initiated background fetches SHALL degrade silently without raising a user-facing error notification. These include the AI-enabled capability probe, the known-tags autocomplete load, the profile family-info load, and the background session validation check. The token-refresh / 401 handling path SHALL NOT raise a user-facing error notification. Silent degradation MAY still log to the developer console for debugging.

#### Scenario: AI-enabled probe fails

- **WHEN** the background probe that checks whether AI tagging is enabled fails
- **THEN** the app treats AI tagging as unavailable
- **AND** no error notification is shown to the user

#### Scenario: Known-tags autocomplete load fails

- **WHEN** the background load of known tags for autocomplete fails
- **THEN** autocomplete falls back to an empty list
- **AND** no error notification is shown to the user

#### Scenario: Session refresh handling does not toast

- **WHEN** a request receives a 401 and the refresh/redirect flow runs
- **THEN** no error notification is shown for the 401 itself
