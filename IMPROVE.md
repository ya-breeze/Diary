# Diary Codebase Improvement Analysis

This document provides a comprehensive analysis of the Diary codebase with constructive criticism and actionable improvement suggestions, organized by priority.

---

## 🔴 Critical Priority Issues

### 1. Hardcoded Session Key (Security Vulnerability)

**Files:**
- `backend/pkg/server/api/api_auth_controller.go:17`
- `backend/pkg/server/webapp/webapp.go:24`

**Problem:**
```go
var store = sessions.NewCookieStore([]byte("SESSION_KEY"))
```

The session store uses a hardcoded, weak secret key. This is a severe security vulnerability that allows:
- Session forgery attacks
- Session hijacking
- Predictable session tokens

**Suggestion:**
1. Add a `GB_SESSION_SECRET` environment variable to `config.go`
2. Generate a cryptographically secure random key (minimum 32 bytes)
3. Fail startup if the secret is not configured in production

```go
// In config.go
SessionSecret string `mapstructure:"sessionSecret"`

// In auth_controller.go
var store = sessions.NewCookieStore([]byte(cfg.SessionSecret))
```

---

### 2. ✅ Insecure Cookie Configuration (FIXED)

**File:** `backend/pkg/server/api/api_auth_controller.go:35-38`

**Problem:**
```go
session.Options.Secure = false
session.Options.HttpOnly = true
session.Options.SameSite = http.SameSiteLaxMode
```

Setting `Secure = false` allows cookies to be transmitted over unencrypted HTTP connections, exposing session tokens to network sniffing attacks.

**Resolution:**
1. ✅ Added `CookieSecure` field to Config struct with default value `true`
2. ✅ Updated both API auth controller and webapp router to use `cfg.CookieSecure`
3. ✅ Added `GB_COOKIE_SECURE` environment variable configuration
4. ✅ Updated documentation (.env.example, docker-compose.yml, Makefile)
5. ✅ All tests updated and passing

---

### 3. No Rate Limiting on Authentication Endpoints

**File:** `backend/pkg/server/middlewares.go`

**Problem:**
The authentication endpoint (`/v1/authorize`) has no rate limiting, making it vulnerable to:
- Brute force password attacks
- Credential stuffing attacks
- Denial of service

**Suggestion:**
1. Implement rate limiting middleware using a library like `golang.org/x/time/rate`
2. Apply stricter limits to authentication endpoints (e.g., 5 attempts per minute per IP)
3. Implement exponential backoff after failed attempts

```go
// Example rate limiter middleware
func RateLimitMiddleware(limiter *rate.Limiter) mux.MiddlewareFunc {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

---

### 4. XSS Vulnerability via bypassSecurityTrustHtml

**File:** `frontend/src/app/diary/diary-editor/diary-editor.component.ts:~450`

**Problem:**
```typescript
this.sanitizer.bypassSecurityTrustHtml(html)
```

Using `bypassSecurityTrustHtml` bypasses Angular's built-in XSS protection. If the markdown content contains malicious scripts, they will be executed.

**Suggestion:**
1. Use a proper markdown sanitization library (e.g., DOMPurify)
2. Configure the markdown parser to strip dangerous HTML
3. Whitelist only safe HTML tags and attributes

```typescript
import DOMPurify from 'dompurify';

const sanitizedHtml = DOMPurify.sanitize(html, {
  ALLOWED_TAGS: ['p', 'br', 'strong', 'em', 'ul', 'ol', 'li', 'a', 'img', 'code', 'pre'],
  ALLOWED_ATTR: ['href', 'src', 'alt', 'class']
});
```

---

## 🟠 High Priority Issues

### 5. Missing Database Indexes

**Files:**
- `backend/pkg/database/models/item.go`
- `backend/pkg/database/models/user.go`

**Problem:**
The `Item` model lacks indexes on frequently queried fields. Only `ItemChange` has proper indexes defined.

```go
type Item struct {
    UserID string `gorm:"primaryKey"`
    Date   string `gorm:"primaryKey"`
    // No indexes on Title, Body for search operations
}
```

**Suggestion:**
Add indexes for search optimization:

```go
type Item struct {
    UserID string `gorm:"primaryKey;index:idx_user_date"`
    Date   string `gorm:"primaryKey;index:idx_user_date;index:idx_date"`
    Title  string `gorm:"index:idx_title"`
    Body   string
    Tags   StringList `gorm:"type:json"`
}
```

---

### 6. No Input Validation on Date Fields

**File:** `backend/pkg/server/api/api_items_service.go`

**Problem:**
Date strings are accepted without format validation. Invalid dates could cause data inconsistencies.

**Suggestion:**
Add date validation:

```go
func validateDateFormat(date string) error {
    _, err := time.Parse("2006-01-02", date)
    if err != nil {
        return fmt.Errorf("invalid date format, expected YYYY-MM-DD: %w", err)
    }
    return nil
}
```

---

### 7. Storage Interface Bloat

**File:** `backend/pkg/database/storage.go:32-60`

**Problem:**
```go
//nolint:interfacebloat
type Storage interface {
    // 20+ methods in a single interface
}
```

The `Storage` interface violates the Interface Segregation Principle with too many methods.

**Suggestion:**
Split into focused interfaces:

```go
type UserStorage interface {
    CreateUser(username, hashedPassword string) (*models.User, error)
    GetUserID(username string) (string, error)
    GetUserByID(userID string) (*models.User, error)
}

type ItemStorage interface {
    GetItem(userID, date string) (*models.Item, error)
    PutItem(userID string, item *models.Item) error
    DeleteItem(userID, itemID string) error
    SearchItems(userID string, params SearchParams) ([]models.Item, error)
}

type ChangeStorage interface {
    GetChangesSince(userID string, sinceID int32, limit int) ([]models.ItemChange, error)
}

type Storage interface {
    UserStorage
    ItemStorage
    ChangeStorage
    Transaction(fn func(tx Storage) error) error
}
```

---

### 8. Missing Pagination on List/Search Endpoints

**File:** `api/openapi.yaml` - `/items` and `/items/search` endpoints

**Problem:**
The search endpoint returns all matching items without pagination, which could cause performance issues with large datasets.

**Suggestion:**
Add pagination parameters to the OpenAPI spec:

```yaml
parameters:
  - name: limit
    in: query
    schema:
      type: integer
      default: 50
      maximum: 100
  - name: offset
    in: query
    schema:
      type: integer
      default: 0
```

---

### 9. Large Component File (Maintainability)

**File:** `frontend/src/app/diary/diary-editor/diary-editor.component.ts` (698 lines)

**Problem:**
The diary editor component is too large and handles multiple responsibilities:
- Markdown editing
- Preview rendering
- Asset management
- Keyboard shortcuts
- Date navigation
- Auto-save

**Suggestion:**
Extract into smaller, focused components and services:

```
diary-editor/
├── diary-editor.component.ts (orchestration only)
├── markdown-editor/
│   └── markdown-editor.component.ts
├── markdown-preview/
│   └── markdown-preview.component.ts
├── services/
│   ├── auto-save.service.ts
│   └── markdown-renderer.service.ts
```

---

## 🟡 Medium Priority Issues

### 10. Inconsistent Naming Convention

**Files:**
- `backend/cmd/commands/root.go:15` - Uses "GeekBudget" but project is "Diary"
- Various config prefixes use `GB_` instead of `DIARY_`

**Problem:**
```go
rootCmd = &cobra.Command{
    Use:   "diary",
    Short: "GeekBudget is a personal finance manager",
```

The codebase has remnants of a previous project name, causing confusion.

**Suggestion:**
1. Update all references from "GeekBudget" to "Diary"
2. Consider renaming environment variable prefix from `GB_` to `DIARY_`
3. Update documentation accordingly

---

### 11. Commented-Out Dead Code

**File:** `backend/pkg/database/models/item.go:10-18`

**Problem:**
```go
// AssetIDs StringList `gorm:"type:json"`

// func (u Item) FromDB() goserver.Item {
//     return goserver.Item{
//         Email:     u.Login,
//         StartDate: u.StartDate,
//     }
// }
```

Dead code clutters the codebase and causes confusion about intended functionality.

**Suggestion:**
Remove commented-out code. Use version control to track historical changes instead.

---

### 12. Unused Field in Auth Service

**File:** `frontend/src/app/core/services/auth.service.ts`

**Problem:**
The `authCheckInProgress` field is declared but never properly utilized for preventing concurrent authentication checks.

**Suggestion:**
Either implement proper concurrent request prevention or remove the unused field:

```typescript
private authCheckInProgress = false;

validateAuthentication(): Observable<boolean> {
  if (this.authCheckInProgress) {
    return this.currentUser$.pipe(
      map(user => !!user),
      take(1)
    );
  }

  this.authCheckInProgress = true;
  return this.loadUserProfile().pipe(
    finalize(() => this.authCheckInProgress = false)
  );
}
```

---

### 13. Missing Error Boundaries in Frontend

**File:** `frontend/src/app/app.component.ts`

**Problem:**
No global error boundary exists to catch and handle unexpected errors gracefully.

**Suggestion:**
Implement an Angular ErrorHandler:

```typescript
@Injectable()
export class GlobalErrorHandler implements ErrorHandler {
  constructor(private toastService: ToastService) {}

  handleError(error: Error): void {
    console.error('Unhandled error:', error);
    this.toastService.error('An unexpected error occurred. Please try again.');
  }
}

// In app.config.ts
providers: [
  { provide: ErrorHandler, useClass: GlobalErrorHandler }
]
```

---

### 14. Inconsistent API Response Patterns

**File:** `api/openapi.yaml`

**Problem:**
Some endpoints return raw data while others wrap responses in objects. For example:
- `GET /items` returns `DiaryItemsListResponse` with `items` array
- `PUT /items` returns `ItemsResponse` directly

**Suggestion:**
Standardize all responses to use a consistent wrapper pattern:

```yaml
ApiResponse:
  type: object
  properties:
    success:
      type: boolean
    data:
      type: object
    error:
      type: string
```

---

### 15. Missing OnPush Change Detection

**Files:** All Angular components

**Problem:**
Components don't use `OnPush` change detection strategy, which could impact performance.

**Suggestion:**
Add `changeDetection: ChangeDetectionStrategy.OnPush` to components that use signals:

```typescript
@Component({
  selector: 'app-diary-editor',
  standalone: true,
  changeDetection: ChangeDetectionStrategy.OnPush,
  // ...
})
```

---

## 🟢 Low Priority Issues

### 16. Magic Numbers and Strings

**Files:** Various

**Problem:**
Hardcoded values scattered throughout the codebase:
- `backend/pkg/auth/jwt.go`: Token expiration `time.Hour * 24`
- `frontend/src/app/core/services/auth.service.ts`: `diary_auth_token`

**Suggestion:**
Extract to constants:

```go
// constants.go
const (
    DefaultTokenExpiration = 24 * time.Hour
    MaxLoginAttempts = 5
    SessionCookieName = "diary_session"
)
```

```typescript
// constants.ts
export const AUTH_TOKEN_KEY = 'diary_auth_token';
export const THEME_KEY = 'diary_theme';
```

---

### 17. Missing TypeScript Strict Null Checks Usage

**File:** `frontend/src/app/shared/models/diary-item.model.ts`

**Problem:**
```typescript
export interface DiaryItem {
  tags?: string[];
  previousDate?: string | null;
  nextDate?: string | null;
}
```

Mixing `undefined` (via `?`) and `null` creates inconsistent null handling.

**Suggestion:**
Choose one approach consistently:

```typescript
export interface DiaryItem {
  tags: string[];  // Always present, empty array if none
  previousDate: string | null;  // Explicitly null when not available
  nextDate: string | null;
}
```

---

### 18. Limited Test Coverage

**Problem:**
- Frontend has service tests but limited component tests
- Backend tests focus on integration/flow tests, missing unit tests for utilities
- No tests for error interceptor, auth guard

**Suggestion:**
1. Add component tests for all Angular components
2. Add unit tests for backend utilities (`pkg/auth/`, `pkg/utils/`)
3. Add tests for edge cases and error conditions
4. Set up code coverage thresholds (e.g., 80% minimum)

---

### 19. Missing API Versioning Strategy

**File:** `api/openapi.yaml`

**Problem:**
API uses `/v1` prefix but there's no documented versioning strategy or deprecation policy.

**Suggestion:**
Document versioning strategy in README:
- How breaking changes are handled
- Deprecation timeline
- Version sunset policy

---

### 20. Docker Compose Security

**File:** `docker-compose.yml:13`

**Problem:**
```yaml
GB_USERS: ${GB_USERS:-test@test.com:JDJhJDEwJC9sVWJpTlBYVlZvcU9ZNUxIZmhqYi4vUnRuVkJNaEw4MTQ2VUdFSXRDeE9Ib0ZoVkRLR3pl}
```

Default credentials in docker-compose could be accidentally used in production.

**Suggestion:**
1. Remove default credentials
2. Require explicit configuration
3. Add validation that fails startup if using default credentials

---

## Summary

| Priority | Count | Key Areas |
|----------|-------|-----------|
| 🔴 Critical | 4 | Security vulnerabilities (session key, cookies, rate limiting, XSS) |
| 🟠 High | 5 | Database optimization, input validation, architecture, maintainability |
| 🟡 Medium | 6 | Code quality, consistency, error handling |
| 🟢 Low | 5 | Best practices, documentation, testing |

### Recommended Action Plan

1. **Immediate (Week 1):** Address all critical security issues
2. **Short-term (Week 2-3):** Implement high priority improvements
3. **Medium-term (Month 1-2):** Address medium priority issues
4. **Ongoing:** Continuously improve low priority items
