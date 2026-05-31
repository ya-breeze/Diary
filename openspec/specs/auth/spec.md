# Feature: Authentication

## Requirement: Login with email and password
A user logs in with their email and password. On success they are redirected to the diary.

### Scenario: Successful login
- **WHEN** the user submits a valid email and correct password
- **THEN** the server sets an access cookie (JWT, 15-minute TTL) and a refresh cookie (1-year TTL)
- **AND** the client stores `isAuthenticated: true` in localStorage
- **AND** the user is redirected to `/diary`

### Scenario: Wrong password
- **WHEN** the user submits a valid email with an incorrect password
- **THEN** a 401 is returned and an error message is shown on the login page
- **AND** no cookies are set

### Scenario: Unknown email
- **WHEN** the user submits an email that does not exist
- **THEN** a 401 is returned (same error as wrong password — no distinction to prevent enumeration)
- **AND** no cookies are set

### Scenario: Empty email field
- **WHEN** the user submits the form with an empty email
- **THEN** a client-side validation error is shown ("Please enter a valid email")
- **AND** no request is sent to the server

### Scenario: Invalid email format
- **WHEN** the user submits a string that is not a valid email address
- **THEN** a client-side validation error is shown ("Please enter a valid email")
- **AND** no request is sent to the server

### Scenario: Empty password field
- **WHEN** the user submits the form with an empty password
- **THEN** a client-side validation error is shown ("Password is required")
- **AND** no request is sent to the server

---

## Requirement: Session persistence across page reloads
Auth state survives a browser refresh.

### Scenario: Authenticated user reloads the page
- **GIVEN** the user is logged in (`isAuthenticated: true` in localStorage)
- **WHEN** the user reloads the page
- **THEN** they remain on their current page without being redirected to `/login`

### Scenario: Session validation on reload
- **GIVEN** `isAuthenticated: true` is in localStorage but the access cookie has expired
- **WHEN** the page loads and `validateSession` is called
- **THEN** the client calls `GET /v1/user`; if it fails, `isAuthenticated` is set to `false` and the user is treated as logged out

---

## Requirement: Automatic token refresh
The access token is short-lived; the refresh token is used to issue a new one silently.

### Scenario: Access token refreshed successfully
- **GIVEN** the access cookie has expired but the refresh cookie is still valid
- **WHEN** an API call returns 401
- **THEN** the client calls `POST /auth/refresh`
- **AND** the server rotates the refresh token and sets new access and refresh cookies
- **AND** the original request is retried transparently

### Scenario: Refresh token reuse detected (compromised)
- **GIVEN** an already-consumed refresh token is presented
- **WHEN** `POST /auth/refresh` is called
- **THEN** the server revokes all sessions for that user and returns 401
- **AND** auth cookies are cleared
- **AND** the user is redirected to `/login`

### Scenario: Refresh token expired or invalid
- **GIVEN** the refresh cookie is expired or missing
- **WHEN** `POST /auth/refresh` is called
- **THEN** the server returns 401
- **AND** the user is redirected to `/login`

---

## Requirement: Logout
The user can log out from the profile page.

### Scenario: Successful logout
- **WHEN** the user clicks "Log out"
- **THEN** `POST /v1/logout` is called
- **AND** the server blacklists the current access token and revokes the refresh token
- **AND** both cookies are cleared by the server
- **AND** client auth state is reset (`user: null`, `isAuthenticated: false`)
- **AND** the user is redirected to `/login`

### Scenario: Logout when server is unreachable
- **GIVEN** the logout API call fails (network error)
- **WHEN** the user clicks "Log out"
- **THEN** client auth state is still reset and the user is still redirected to `/login`
- **AND** the error is logged to the console (not shown to the user)

---

## Requirement: Protected routes
All routes except `/login` and `/api/*` require authentication.

### Scenario: Unauthenticated user visits a protected route
- **GIVEN** the user has no valid session (`isAuthenticated: false`)
- **WHEN** they navigate to any protected page (e.g. `/diary`, `/search`, `/profile`)
- **THEN** they are redirected to `/login`
