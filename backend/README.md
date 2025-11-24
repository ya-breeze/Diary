# Diary.BE

A web-based personal diary application with a focus on privacy and simplicity.
Write, store, and manage your daily thoughts and experiences securely.

## Features

- 🔒 Secure authentication with JWT tokens
- 📝 Markdown support for rich text formatting
- 🌐 Web-based interface using Bootstrap
- 📱 Responsive design for mobile and desktop
- 💾 SQLite database for easy deployment
- 🐳 Docker support for containerized deployment

## Development

### 📋 Before Starting Any Task

1. **Review Guidelines**: Read `docs/development-guidelines.md` for best practices

### 🔧 Development Workflow

```bash
# During development
make test    # Run tests
make lint    # Check code quality

# Before committing
make all     # Run full validation
```

## Technical Overview

The application is built using:

- Go (backend)
- Bootstrap (frontend)
- SQLite (database)
- OpenAPI/Swagger (API documentation)

### Architecture

The application follows a layered architecture:

1. **Web Layer** - Handles HTTP requests and user interface (webapp/)
2. **API Layer** - RESTful API defined using OpenAPI spec (api/openapi.yaml)
3. **Service Layer** - Business logic implementation
4. **Data Layer** - Database operations using GORM

### Security

- Password hashing using bcrypt
- Session management with secure cookies
- JWT-based API authentication

## Getting Started

### Prerequisites

- Go 1.24+
- SQLite
- Docker (optional)

### Quick Start

1. Using Docker:

```bash
docker-compose up
```

2. Using Make:

```bash
make run
```

Default credentials:

- Username: test@example.com
- Password: test

## Development

- `make build` - Build the application
- `make test` - Run tests
- `make lint` - Run linters
- `make validate` - Validate code

## Configuration

Environment variables:

- `GB_USERS` - User credentials
- `GB_DBPATH` - Database path
- `GB_ASSETPATH` - Assets path
- `GB_ALLOWEDORIGINS` - Comma-separated list of allowed CORS origins (default: `http://localhost:3000`)
- `GB_MAXPERFILESIZEMB` - Max size per uploaded file in MB (default 25)
- `GB_MAXBATCHFILES` - Max number of files per batch (default 10)
- `GB_MAXBATCHTOTALSIZEMB` - Max total size per batch in MB (default 100)

## Batch Asset Uploads

- API endpoint: `POST /v1/assets/batch`

  - multipart/form-data with repeated field `assets`
  - Response JSON example:
    ```json
    {
      "files": [
        {
          "originalName": "photo.jpg",
          "savedName": "<uuid>.jpg",
          "size": 12345,
          "contentType": "image/jpeg"
        }
      ],
      "count": 1
    }
    ```

- Single upload remains available at `POST /v1/assets` with field `asset`

- Web UI: the Edit page file picker supports multi-select; when multiple files are chosen, it automatically calls the batch endpoint. A progress bar and errors are shown inline.
