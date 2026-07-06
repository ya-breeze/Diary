## ADDED Requirements

### Requirement: Failed user-initiated actions surface an error to the user

The frontend SHALL display a human-readable error message to the user whenever an action the user explicitly initiated fails. A user-initiated action is one triggered by a direct user gesture, including: uploading images, saving a diary entry, requesting tag suggestions, accepting a suggested tag, dismissing a suggested tag, updating the AI tagging setting, and logging out. Such failures MUST NOT be swallowed into logging only.

#### Scenario: Image upload fails

- **WHEN** the user uploads one or more images and the upload request fails
- **THEN** the app displays an error notification containing a human-readable message
- **AND** the upload progress indicator is cleared

#### Scenario: Saving an entry fails

- **WHEN** the user saves a diary entry and the save request fails
- **THEN** the app displays an error notification containing a human-readable message
- **AND** the user remains in the editor with their content intact

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

The frontend SHALL normalize the different error shapes produced across the app — the `ApiError` thrown by the shared API client, a plain `Error` (e.g. the raw XHR rejection from the batch asset upload), and unknown/thrown values — into a single human-readable message used for user-facing notifications.

#### Scenario: ApiError is normalized

- **WHEN** an action fails with an `ApiError` carrying a status and message
- **THEN** the notification shows a message derived from that error rather than a generic placeholder

#### Scenario: Non-ApiError failure is normalized

- **WHEN** an action fails with a plain `Error` or a non-error thrown value
- **THEN** the normalizer still produces a readable message and the app displays a notification without throwing

### Requirement: Background enhancements degrade silently

Optional, non-user-initiated background fetches SHALL degrade silently without raising a user-facing error notification. These include the AI-enabled capability probe and the known-tags autocomplete load. The token-refresh / 401 handling path SHALL NOT raise a user-facing error notification. Silent degradation MAY still log to the developer console for debugging.

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
