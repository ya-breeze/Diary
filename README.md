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

  - **Dual Mode Interface**
    - View Mode (default): Read-only mode for browsing entries with rendered content
    - Edit Mode: Full editing capabilities with markdown editor
    - Quick toggle between modes (Ctrl+E)
    - Mode preference persistence
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
  - Toggle between edit and preview modes (Ctrl+P)
  - GitHub-flavored markdown styling
  - Search functionality by text and tags
  - Search results page with clickable entries
  - Quick navigation to search from editor

- **UI/UX Improvements**

  - Responsive design for mobile, tablet, and desktop
  - Loading spinners and progress indicators
  - Toast notifications for user feedback
  - Dark/light theme toggle with system preference detection
  - Comprehensive keyboard shortcuts
  - Accessibility improvements (ARIA labels, keyboard navigation)

- **Core Infrastructure**
  - HTTP interceptors for auth and error handling
  - Service layer for API communication
  - TypeScript models for type safety
  - Environment-based configuration
  - OnPush change detection for performance
  - Comprehensive unit test coverage (94.83%)

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

## Keyboard Shortcuts

The application includes comprehensive keyboard shortcuts for efficient navigation:

| Shortcut   | Action                                 |
| ---------- | -------------------------------------- |
| `Ctrl + E` | Toggle between view and edit modes     |
| `Ctrl + S` | Save current diary entry               |
| `Alt + ←`  | Navigate to previous day               |
| `Alt + →`  | Navigate to next day                   |
| `Ctrl + P` | Toggle markdown preview (in edit mode) |
| `Ctrl + F` | Open search                            |
| `/`        | Show keyboard shortcuts help           |
| `Esc`      | Close modals/dialogs                   |

## Theme Support

The application supports both light and dark themes:

- **Auto-detection**: Automatically detects system preference on first load
- **Manual toggle**: Click the theme toggle button in the header
- **Persistence**: Theme preference is saved to localStorage
- **Smooth transitions**: All theme changes are animated

## Project Structure

```
src/app/
├── core/                    # Singleton services, guards, interceptors
│   ├── guards/             # Route guards (auth)
│   ├── interceptors/       # HTTP interceptors (auth, error)
│   └── services/           # Core services (auth, diary, asset, theme, toast, keyboard)
├── shared/                 # Shared models and utilities
│   ├── components/         # Reusable components
│   │   ├── asset-gallery/
│   │   ├── asset-upload/
│   │   ├── asset-preview-modal/
│   │   ├── keyboard-shortcuts-help/
│   │   ├── loading-spinner/
│   │   ├── theme-toggle/
│   │   └── toast-container/
│   └── models/             # TypeScript interfaces
├── auth/                   # Authentication feature
│   └── login/              # Login component
├── diary/                  # Diary feature
│   ├── diary-list/         # Diary list component
│   ├── diary-editor/       # Diary editor component
│   └── diary-search/       # Search component
├── app.ts                  # Root component
├── app.config.ts           # Application configuration
└── app.routes.ts           # Route definitions
```

## Configuration

### API Endpoints

The application uses different configurations for development and production:

**Development** (`src/environments/environment.ts`):

```typescript
export const environment = {
  production: false,
  apiUrl: "http://localhost:8080/v1",
};
```

- Makes direct requests to the backend at `http://localhost:8080`
- Backend must be running on `http://localhost:8080`
- Backend must have proper CORS headers configured

**Production** (`src/environments/environment.prod.ts`):

```typescript
export const environment = {
  production: true,
  apiUrl: "http://truenas.local:8880/v1",
};
```

- Makes direct requests to the backend at `http://truenas.local:8880`

**Required CORS Headers** (for both development and production):

- `Access-Control-Allow-Origin: <frontend-origin>` (specific origin, not `*`)
- `Access-Control-Allow-Credentials: true`
- `Access-Control-Allow-Headers: Content-Type, Authorization`
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`

### Cookie-Based Authentication

The application uses **dual authentication**:

1. **JWT Bearer Token** - For API requests (stored in localStorage)
2. **HTTP-only Cookies** - For media/asset requests (set by backend)

All HTTP requests include `withCredentials: true` to ensure cookies are sent and received properly.

**Important**: The backend must set cookies with appropriate attributes:

- `HttpOnly` - Prevents JavaScript access
- `SameSite=Lax` or `SameSite=None; Secure` - For cross-site requests
- `Path=/` - Available for all routes
- Appropriate `Max-Age` or `Expires`

## API Integration

The application integrates with a backend API defined in `api/openapi.yaml`:

- `POST /v1/authorize` - User authentication
- `GET /v1/user` - Get user profile
- `GET /v1/items` - List diary items (with search support)
- `GET /v1/items/{date}` - Get diary item by date
- `PUT /v1/items/{date}` - Create/update diary item
- `POST /v1/assets` - Upload single asset
- `POST /v1/assets/batch` - Batch upload assets
- `GET /v1/assets/{path}` - Download asset
- `DELETE /v1/assets/{path}` - Delete asset
