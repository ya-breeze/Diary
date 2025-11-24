---
type: "always_apply"
---

# Backend Development Guidelines (Go)

This document defines the development standards, patterns, and best practices for the Diary backend application written in Go.

## Project Overview

- **Language**: Go 1.24+
- **Framework**: Gorilla Mux (HTTP routing), GORM (ORM)
- **Database**: SQLite
- **API**: OpenAPI 3.0 specification-driven development
- **Testing**: Ginkgo/Gomega
- **Authentication**: JWT with bcrypt password hashing

## Architecture Principles

### Layered Architecture

The backend follows a strict layered architecture:

1. **API Layer** (`pkg/server/api/`) - HTTP handlers and request/response processing
2. **Service Layer** (`pkg/server/api/*_service.go`) - Business logic implementation
3. **Data Layer** (`pkg/database/`) - Database operations using GORM
4. **Model Layer** (`pkg/database/models/`) - Data structures and domain models
5. **Web Layer** (`pkg/server/webapp/`) - Server-side rendered web interface

### Package Organization

```
backend/
├── cmd/                    # Application entry points
│   ├── main.go            # Main application
│   └── commands/          # CLI commands
├── pkg/
│   ├── auth/              # Authentication utilities (JWT, bcrypt)
│   ├── config/            # Configuration management
│   ├── database/          # Database layer
│   │   └── models/        # GORM models
│   ├── generated/         # OpenAPI-generated code (DO NOT EDIT)
│   │   └── goserver/      # Generated server stubs
│   ├── server/            # HTTP server implementation
│   │   ├── api/           # API service implementations
│   │   └── webapp/        # Web application routes
│   └── utils/             # Shared utilities
└── test/                  # Integration and E2E tests
    └── flows/             # Test scenarios
```

## Code Standards

### OpenAPI-First Development

- **ALWAYS** define API changes in `api/openapi.yaml` first
- Run `HOST_PWD=$(pwd) make generate` to regenerate server code after OpenAPI changes
- Never manually edit files in `pkg/generated/` - they are auto-generated
- Implement service interfaces defined in generated code

### Service Implementation Pattern

All API services must implement the generated interface:

```go
type ItemsAPIServiceImpl struct {
    logger *slog.Logger
    db     database.Storage
}

func NewItemsAPIService(logger *slog.Logger, db database.Storage) goserver.ItemsAPIService {
    return &ItemsAPIServiceImpl{
        logger: logger,
        db:     db,
    }
}
```

### Authentication & Authorization

- Use JWT tokens for API authentication
- Extract user ID from context using `common.UserIDKey`
- Always validate user context in service methods:

```go
userID, ok := ctx.Value(common.UserIDKey).(string)
if !ok {
    s.logger.Error("User ID not found in context")
    return goserver.Response(401, nil), nil
}
```

- Use bcrypt for password hashing (never store plain passwords)
- Session cookies for web interface, JWT for API

### Database Operations

#### GORM Models

- Define models in `pkg/database/models/`
- Use proper GORM tags for schema definition
- Implement `FromDB()` methods to convert to API models
- Use custom types (e.g., `StringList`) for JSON fields

#### Transactions

- **CRITICAL**: Wrap related database operations in transactions
- Use transactions for operations involving multiple tables
- Example pattern:

```go
err := s.db.Transaction(func(tx database.Storage) error {
    // Perform multiple operations
    if err := tx.SaveItem(item); err != nil {
        return err
    }
    if err := tx.CreateChangeRecord(change); err != nil {
        return err
    }
    return nil
})
```

#### Change Tracking

- Always create change records for data modifications
- Use auto-incrementing IDs for efficient synchronization
- Store complete item snapshots in change records
- Include operation type (created, updated, deleted)

### Error Handling

- Use structured logging with `slog`
- Log errors with context (user ID, item ID, etc.)
- Return appropriate HTTP status codes
- Never expose internal errors to clients

```go
if err != nil {
    s.logger.Error("Failed to save item", "userID", userID, "error", err)
    return goserver.Response(500, nil), nil
}
```

### Type Safety

- Always validate type conversions
- Use bounds checking for integer conversions (uint ↔ int32)
- Add `#nosec` annotations for intentionally safe operations
- Enable strict type checking in tests

## Testing Requirements

### Test Structure

- Use Ginkgo/Gomega for all tests
- Place integration tests in `test/flows/`
- Follow the test pyramid: unit → integration → E2E

### Test Categories

1. **Unit Tests**: Models, utilities, business logic
2. **Integration Tests**: Database operations, API handlers
3. **End-to-End Tests**: Complete user workflows
4. **Concurrency Tests**: Atomic operations, race conditions

### Test Patterns

```go
var _ = Describe("Feature", func() {
    Context("when condition", func() {
        It("should behavior", func() {
            // Arrange
            // Act
            // Assert
            Expect(result).To(Equal(expected))
        })
    })
})
```

### Test Requirements

- Test all API endpoints
- Test error conditions and edge cases
- Test authentication and authorization
- Test database transactions and rollbacks
- Test concurrent operations where applicable

## Development Workflow

### Pre-Development Checklist

1. ✅ Verify existing code builds: `make build`
2. ✅ Verify tests pass: `make test`
3. ✅ Review OpenAPI specification for API contracts

### Implementation Process

1. **Design Phase**

   - Update `api/openapi.yaml` for API changes
   - Design database schema and models
   - Plan transaction boundaries

2. **Implementation Phase**

   - Run `HOST_PWD=$(pwd) make generate` if OpenAPI changed
   - Implement service interfaces in `pkg/server/api/`
   - Implement database operations in `pkg/database/`
   - Add proper error handling and logging

3. **Testing Phase**

   - Write unit tests for new models/logic
   - Write integration tests for database operations
   - Write E2E tests for complete workflows
   - Test edge cases and error conditions

4. **Quality Assurance**
   - Run `make build` - must succeed
   - Run `make test` - all tests must pass
   - Run `make lint` - fix all linting issues
   - Format code: `go tool mvdan.cc/gofumpt -l -w .`

### Build Commands

```bash
# Build application (NEVER CANCEL - wait 3+ minutes on first run)
make build

# Run tests (NEVER CANCEL - wait 2+ minutes)
make test

# Lint code (NEVER CANCEL - wait 8+ minutes on first run)
make lint

# Format code
go tool mvdan.cc/gofumpt -l -w .

# Validate OpenAPI spec
HOST_PWD=$(pwd) make validate

# Regenerate code from OpenAPI (only when api/openapi.yaml changes)
HOST_PWD=$(pwd) make generate
```

### Running the Application

```bash
# Always build first
make build

# Run with default configuration
make run

# Manual run with custom configuration
GB_USERS=test@test.com:JDJhJDEwJC9sVWJpTlBYVlZvcU9ZNUxIZmhqYi4vUnRuVkJNaEw4MTQ2VUdFSXRDeE9Ib0ZoVkRLR3pl \
GB_DISABLEIMPORTERS=true \
GB_DBPATH=$(pwd)/diary.db \
GB_ASSETPATH=$(pwd)/diary-assets \
./bin/diary server
```

## Security Best Practices

### Password Management

- **NEVER** store passwords in plain text
- Use bcrypt with `bcrypt.DefaultCost` for hashing
- Validate password strength on user creation
- Use constant-time comparison for password verification

### JWT Tokens

- Set appropriate expiration times (default: 24 hours)
- Use strong secrets (minimum 32 characters)
- Validate issuer and subject claims
- Use HMAC-SHA256 signing method

### Input Validation

- Validate all user inputs
- Sanitize file paths for asset operations
- Prevent path traversal attacks
- Validate file types for uploads

### CORS Configuration

- Configure allowed origins explicitly (no wildcards with credentials)
- Set `Access-Control-Allow-Credentials: true`
- Validate origin headers
- Use secure cookie settings

## API Design Patterns

### Pagination

- Use `hasMore` and `nextId` pattern for efficient pagination
- Implement cursor-based pagination for large datasets
- Include total count when appropriate

```go
type ListResponse struct {
    Items   []Item `json:"items"`
    HasMore bool   `json:"hasMore"`
    NextID  *int32 `json:"nextId,omitempty"`
}
```

### Response Patterns

- Use `goserver.Response(statusCode, body)` for all responses
- Return appropriate HTTP status codes:
  - 200: Success
  - 201: Created
  - 400: Bad request
  - 401: Unauthorized
  - 404: Not found
  - 500: Internal server error

### File Handling

- Use `os.File` for file responses
- Framework handles file closing automatically
- Set appropriate content types
- Validate file access permissions

## Mobile Synchronization

### Change Tracking Design

- Use auto-incrementing IDs for change ordering
- Store complete item snapshots (not deltas)
- Include metadata: timestamp, operation type, user ID
- Index on user ID and timestamp for efficient queries

### Sync API Patterns

- Implement incremental sync with `sinceId` parameter
- Return changes in chronological order
- Include `hasMore` flag for pagination
- Support filtering by operation type

## Common Patterns

### Middleware

- Authentication middleware validates JWT tokens
- Extracts user ID and adds to request context
- Skips auth for public endpoints (`/`, `/web/*`, `/v1/authorize`)

### Custom Routers

- Implement `goserver.Router` interface
- Define routes in `Routes()` method
- Use for endpoints not in OpenAPI spec (e.g., batch uploads)

### Asset Management

- Store assets in user-specific directories
- Use UUIDs for asset filenames
- Validate file access by user ID
- Support batch uploads for efficiency

## Performance Considerations

### Database Optimization

- Add indexes on frequently queried fields
- Use composite indexes for multi-column queries
- Optimize queries with `EXPLAIN QUERY PLAN`
- Use prepared statements (GORM handles this)

### Caching Strategy

- Cache user authentication data
- Use in-memory caching for frequently accessed data
- Invalidate cache on data modifications

## Logging Standards

### Structured Logging

- Use `slog` for all logging
- Include context in log messages
- Use appropriate log levels:
  - `Info`: Normal operations
  - `Warn`: Unexpected but handled situations
  - `Error`: Errors requiring attention

### Log Context

```go
s.logger.Info("Operation completed",
    "userID", userID,
    "itemID", itemID,
    "duration", duration)
```

## Configuration Management

### Environment Variables

- Use Viper for configuration management
- Support environment variables with `GB_` prefix
- Provide sensible defaults
- Document all configuration options

### Required Configuration

- `GB_USERS`: User credentials (email:bcrypt_hash)
- `GB_DBPATH`: SQLite database path
- `GB_ASSETPATH`: Asset storage directory
- `GB_JWTSECRET`: JWT signing secret
- `GB_ISSUER`: JWT issuer identifier

## Code Quality Standards

### Linting

- **ALWAYS** run `make lint` before committing
- Fix all linting issues - no exceptions
- Use `#nosec` annotations sparingly and with justification
- Follow golangci-lint recommendations

### Code Formatting

- Use `gofumpt` for consistent formatting
- Run formatter before committing
- Configure IDE to format on save

### Code Review Checklist

- [ ] OpenAPI spec updated (if API changed)
- [ ] Tests written and passing
- [ ] Linting issues resolved
- [ ] Error handling implemented
- [ ] Logging added for important operations
- [ ] Transactions used for multi-step operations
- [ ] Security considerations addressed
- [ ] Documentation updated

## Common Pitfalls to Avoid

### ❌ Don't

- Manually edit generated code in `pkg/generated/`
- Store passwords in plain text
- Skip transaction boundaries for related operations
- Ignore linting errors
- Cancel long-running build/test commands
- Use `SELECT *` in queries
- Expose internal error details to clients
- Skip input validation

### ✅ Do

- Use OpenAPI-first development
- Wrap related operations in transactions
- Validate all user inputs
- Use structured logging with context
- Write comprehensive tests
- Handle errors gracefully
- Use type-safe conversions
- Follow the established patterns

## Troubleshooting

### Build Failures

- If generated packages are missing: `git restore pkg/generated/`
- If OpenAPI validation fails: Check `api/openapi.yaml` syntax
- If dependencies are missing: `go mod download`

### Test Failures

- Check database state between tests
- Verify test isolation
- Check for race conditions in concurrent tests
- Review test setup and teardown

### Runtime Issues

- Check environment variables are set correctly
- Verify database file permissions
- Check asset directory permissions
- Review logs for error context

## Additional Resources

- OpenAPI Specification: `api/openapi.yaml`
- Existing Guidelines: `1/development-guidelines.md`
- Copilot Instructions: `1/copilot-instructions.md`
- Backend README: `backend/README.md`
