# Personal Diary Frontend

A modern Angular 20 application for managing personal diary entries with authentication, date navigation, and asset management.

## Features

### Implemented ✅

- **Authentication System**

  - JWT-based authentication
  - Login/logout functionality
  - Protected routes with auth guards
  - Automatic token management

- **Diary Management**

  - View diary entries by date
  - Navigate between dates (previous/next)
  - Date picker for quick navigation
  - Display entry title, body, and tags
  - Create and edit diary entries
  - Rich text editor with markdown support
  - Tags management (add/remove)
  - Unsaved changes detection
  - Auto-save prompts

- **Asset Management**

  - Drag-and-drop file upload
  - Multiple file selection and batch upload
  - Upload progress tracking
  - Asset gallery with thumbnail previews
  - Image and video preview modal
  - Copy markdown links for assets
  - Asset deletion with confirmation
  - Responsive grid layout

- **Enhanced Features**

  - Markdown preview with live rendering
  - Toggle between edit and preview modes
  - GitHub-flavored markdown styling
  - Search functionality by text and tags
  - Search results page with clickable entries
  - Quick navigation to search from editor

- **Core Infrastructure**
  - HTTP interceptors for auth and error handling
  - Service layer for API communication
  - TypeScript models for type safety
  - Environment-based configuration

### Planned 🚧

- Markdown editor with live preview
- Search functionality with filters
- Asset upload and management
- Calendar view for navigation
- Dark/light theme
- Keyboard shortcuts

## Tech Stack

- **Angular 20.3.0** - Latest Angular with standalone components
- **TypeScript 5.8** - Strict mode enabled
- **RxJS 7.8** - Reactive programming
- **Angular Signals** - Modern reactive state management

## Getting Started

### Prerequisites

- Node.js 18+ and npm
- Backend API running (see api/openapi.yaml for specification)

### Installation

```bash
# Install dependencies
npm install
# or
make install
```

### Development

```bash
# Start development server
npm start
# or
make dev

# The app will be available at http://localhost:4200/
```

### Build

```bash
# Build for production
npm run build
# or
make build

# Output will be in dist/Diary.FE/
```

### Testing

```bash
# Run unit tests
npm test
# or
make test

# Run tests in watch mode
make test-watch

# Generate coverage report
make coverage
```

## Project Structure

```
src/app/
├── core/                    # Singleton services, guards, interceptors
│   ├── guards/             # Route guards (auth)
│   ├── interceptors/       # HTTP interceptors (auth, error)
│   └── services/           # Core services (auth, diary, asset)
├── shared/                 # Shared models and utilities
│   └── models/             # TypeScript interfaces
├── auth/                   # Authentication feature
│   └── login/              # Login component
├── diary/                  # Diary feature
│   ├── diary-list/         # Diary list component
│   └── diary-editor/       # Diary editor component
├── app.ts                  # Root component
├── app.config.ts           # Application configuration
└── app.routes.ts           # Route definitions
```

## Configuration

### API Endpoints

Update the API URL in environment files:

**Development** (`src/environments/environment.ts`):

```typescript
export const environment = {
  production: false,
  apiUrl: "http://localhost:8080/v1",
};
```

**Production** (`src/environments/environment.prod.ts`):

```typescript
export const environment = {
  production: true,
  apiUrl: "/v1",
};
```

## API Integration

The application integrates with a backend API defined in `api/openapi.yaml`:

- `POST /v1/authorize` - User authentication
- `GET /v1/user` - Get user profile
- `GET /v1/items` - List diary items
- `GET /v1/items/{date}` - Get diary item by date
- `PUT /v1/items/{date}` - Create/update diary item
- `POST /v1/assets` - Upload asset
- `POST /v1/assets/batch` - Batch upload assets

## Development Commands

```bash
make help          # Show all available commands
make dev           # Start development server
make build         # Build for production
make test          # Run tests
make lint          # Run linter
make lint-fix      # Fix linting issues
make format        # Format code with Prettier
make clean         # Clean build artifacts
```

## Authentication Flow

1. User navigates to the app
2. Auth guard checks for valid JWT token
3. If no token, redirect to `/login`
4. User enters credentials
5. On successful login:
   - Token is stored in localStorage
   - User profile is loaded
   - Redirect to `/diary`
6. All API requests include JWT token in Authorization header
7. On 401 error, user is logged out and redirected to login

## Contributing

See `IMPLEMENTATION.md` for detailed implementation plan and `PROGRESS.md` for current status.

## License

Private project
