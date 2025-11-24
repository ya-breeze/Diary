# Diary

A modern, privacy-focused personal diary application with a powerful web interface and secure backend. Write, store, and manage your daily thoughts and experiences with rich markdown support, asset management, and seamless synchronization.

## 🌟 Features

### Core Functionality

- 🔒 **Secure Authentication** - JWT-based authentication with bcrypt password hashing
- 📝 **Rich Text Editing** - Full markdown support with preview
- 📅 **Date Navigation** - Easy navigation between diary entries by date
- 🏷️ **Tag Management** - Organize entries with customizable tags
- 🔍 **Search** - Full-text search across entries and tags
- 📱 **Responsive Design** - Works seamlessly on mobile, tablet, and desktop

### Asset Management

- 📎 **File Uploads** - Drag-and-drop or multi-select file uploads
- 🖼️ **Media Gallery** - Visual gallery with thumbnail previews
- 🎬 **Media Preview** - Built-in image and video preview modal
- 📋 **Markdown Links** - One-click copy of markdown-formatted asset links
- 🗑️ **Asset Deletion** - Safe deletion with confirmation

### User Experience

- 🌓 **Theme Support** - Light/dark mode with system preference detection
- ⌨️ **Keyboard Shortcuts** - Comprehensive shortcuts for power users
- 💾 **Auto-save Protection** - Warns before losing unsaved changes
- 🔔 **Toast Notifications** - Clear feedback for all actions
- ♿ **Accessibility** - ARIA labels and keyboard navigation support

### Technical Features

- 🐳 **Docker Support** - Easy containerized deployment
- 🔄 **Mobile Sync** - Change tracking for mobile app synchronization
- 📊 **OpenAPI Spec** - Well-documented REST API
- 💾 **SQLite Database** - Simple, file-based storage
- 🧪 **Comprehensive Tests** - High test coverage for both frontend and backend

## 🏗️ Architecture

The application consists of two main components:

### Backend (Go)

- **Language**: Go 1.24+
- **Framework**: Gorilla Mux (routing), GORM (ORM)
- **Database**: SQLite
- **API**: OpenAPI 3.0 specification-driven development
- **Testing**: Ginkgo/Gomega
- **Location**: `backend/`

### Frontend (Angular)

- **Framework**: Angular 20+
- **Language**: TypeScript 5.8 (strict mode)
- **State Management**: Angular Signals + RxJS 7.8
- **Architecture**: Standalone components (no NgModules)
- **Testing**: Jasmine + Karma
- **Location**: `frontend/`

### Layered Architecture

```
┌─────────────────────────────────────────┐
│         Frontend (Angular)              │
│  ┌─────────────────────────────────┐   │
│  │  Components & Services          │   │
│  │  (Signals, RxJS, HTTP)          │   │
│  └─────────────────────────────────┘   │
└─────────────────┬───────────────────────┘
                  │ HTTP/REST
                  │ (JWT Auth)
┌─────────────────▼───────────────────────┐
│         Backend (Go)                    │
│  ┌─────────────────────────────────┐   │
│  │  API Layer (OpenAPI)            │   │
│  ├─────────────────────────────────┤   │
│  │  Service Layer (Business Logic) │   │
│  ├─────────────────────────────────┤   │
│  │  Data Layer (GORM)              │   │
│  └─────────────────────────────────┘   │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│         SQLite Database                 │
│  ┌─────────────────────────────────┐   │
│  │  Diary Items, Users, Assets     │   │
│  │  Change Tracking for Sync       │   │
│  └─────────────────────────────────┘   │
└─────────────────────────────────────────┘
```

## 🚀 Quick Start

### Prerequisites

- **Go** 1.24 or higher
- **Node.js** 18+ and npm
- **Docker** (optional, for containerized deployment)
- **SQLite** (usually pre-installed on most systems)

### Installation

```bash
# Clone the repository
git clone <repository-url>
cd Diary

# Install dependencies
make install
```

### Running the Application

#### Option 1: Using Make (Recommended)

```bash
# Build and run both backend and frontend
make build
make run
```

#### Option 2: Using Docker

```bash
# Build and run with Docker Compose
make compose
# or
docker-compose up --build
```

#### Option 3: Manual Setup

**Backend:**

```bash
cd backend
go build -o bin/diary cmd/main.go
GB_USERS=test@test.com:JDJhJDEwJC9sVWJpTlBYVlZvcU9ZNUxIZmhqYi4vUnRuVkJNaEw4MTQ2VUdFSXRDeE9Ib0ZoVkRLR3pl \
GB_DBPATH=../diary.db \
GB_ASSETPATH=../diary-assets \
./bin/diary server
```

**Frontend:**

```bash
cd frontend
npm install
npm start
```

### Default Credentials

- **Email**: test@test.com
- **Password**: test

### Access the Application

- **Frontend**: http://localhost:4200
- **Backend API**: http://localhost:8080
- **API Documentation**: See `api/openapi.yaml`

## 📚 Documentation

- **[Backend Guidelines](.augment/rules/backend-guidelines.md)** - Go development standards and patterns
- **[Frontend Guidelines](.augment/rules/frontend-guidelines.md)** - Angular development standards and patterns
- **[Backend README](backend/README.md)** - Backend-specific documentation
- **[Frontend README](frontend/README.md)** - Frontend-specific documentation
- **[OpenAPI Specification](api/openapi.yaml)** - Complete API documentation

## 🛠️ Development

### Development Workflow

```bash
# Check dependencies
make check-deps

# Install all dependencies
make install

# Run tests
make test

# Run linters
make lint

# Validate OpenAPI spec
make validate

# Run all checks (build + test + validate + lint)
make all
```

### Backend Development

```bash
cd backend

# Build
go build -o bin/diary cmd/main.go

# Run tests
go tool github.com/onsi/ginkgo/v2/ginkgo -r

# Watch tests
make watch

# Lint
go tool github.com/golangci/golangci-lint/cmd/golangci-lint run

# Format code
go tool mvdan.cc/gofumpt -l -w .

# Generate code from OpenAPI spec (when api/openapi.yaml changes)
HOST_PWD=$(pwd) make generate
```

### Frontend Development

```bash
cd frontend

# Start dev server
npm start
# or
ng serve

# Build for production
npm run build

# Run tests
npm test

# Run tests in watch mode
npm test -- --watch

# Generate coverage report
npm test -- --code-coverage

# Lint
npm run lint

# Format code
npx prettier --write "src/**/*.{ts,html,css,scss,json}"
```

### Code Quality Standards

#### Backend (Go)

- ✅ OpenAPI-first development - always update `api/openapi.yaml` first
- ✅ Use GORM for database operations
- ✅ Wrap related operations in transactions
- ✅ Use structured logging with `slog`
- ✅ Write tests using Ginkgo/Gomega
- ✅ Run `make lint` before committing
- ✅ Never edit generated code in `pkg/generated/`

#### Frontend (Angular)

- ✅ Use standalone components (no NgModules)
- ✅ Use signals for reactive state management
- ✅ Use reactive forms pattern
- ✅ Clean up subscriptions in `ngOnDestroy`
- ✅ Write unit tests for components and services
- ✅ Follow accessibility best practices
- ✅ Use TypeScript strict mode

### Pre-Commit Checklist

Before committing code, ensure:

- [ ] Code builds successfully (`make build`)
- [ ] All tests pass (`make test`)
- [ ] Linting passes (`make lint`)
- [ ] OpenAPI spec is valid (`make validate`)
- [ ] No console errors or warnings
- [ ] Documentation is updated if needed
- [ ] Commit message is descriptive

## ⌨️ Keyboard Shortcuts

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

## ⚙️ Configuration

### Environment Variables

#### Backend Configuration

| Variable                 | Description                                  | Default                 |
| ------------------------ | -------------------------------------------- | ----------------------- |
| `GB_USERS`               | User credentials (email:bcrypt_hash)         | Required                |
| `GB_DBPATH`              | SQLite database file path                    | `./diary.db`            |
| `GB_ASSETPATH`           | Directory for storing uploaded assets        | `./diary-assets`        |
| `GB_JWTSECRET`           | Secret key for JWT token signing             | Auto-generated          |
| `GB_ISSUER`              | JWT issuer identifier                        | `diary-backend`         |
| `GB_ALLOWEDORIGINS`      | Comma-separated list of allowed CORS origins | `http://localhost:3000` |
| `GB_MAXPERFILESIZEMB`    | Maximum size per uploaded file (MB)          | `25`                    |
| `GB_MAXBATCHFILES`       | Maximum number of files per batch upload     | `10`                    |
| `GB_MAXBATCHTOTALSIZEMB` | Maximum total size per batch upload (MB)     | `100`                   |
| `GB_DISABLEIMPORTERS`    | Disable automatic data importers             | `false`                 |

#### Frontend Configuration

Edit `frontend/src/environments/environment.ts` for development:

```typescript
export const environment = {
  production: false,
  apiUrl: "http://localhost:8080/v1",
};
```

Edit `frontend/src/environments/environment.prod.ts` for production:

```typescript
export const environment = {
  production: true,
  apiUrl: "http://your-production-server:8080/v1",
};
```

### CORS Configuration

The backend must be configured with appropriate CORS headers:

- `Access-Control-Allow-Origin`: Frontend origin (specific, not `*`)
- `Access-Control-Allow-Credentials`: `true`
- `Access-Control-Allow-Headers`: `Content-Type, Authorization`
- `Access-Control-Allow-Methods`: `GET, POST, PUT, DELETE, OPTIONS`

Set `GB_ALLOWEDORIGINS` to include your frontend URL (e.g., `http://localhost:4200`).

## 🔒 Security

### Authentication

- **JWT Tokens**: Used for API authentication, stored in localStorage
- **HTTP-only Cookies**: Used for media/asset requests, set by backend
- **Password Hashing**: bcrypt with default cost factor
- **Token Expiration**: 24 hours (configurable)

### Best Practices

- ✅ Never store passwords in plain text
- ✅ Use HTTPS in production
- ✅ Set secure cookie attributes (`HttpOnly`, `SameSite`, `Secure`)
- ✅ Validate all user inputs
- ✅ Sanitize file paths to prevent path traversal
- ✅ Use prepared statements (GORM handles this)
- ✅ Keep dependencies up to date

## 🧪 Testing

### Backend Tests

```bash
cd backend

# Run all tests
go tool github.com/onsi/ginkgo/v2/ginkgo -r

# Run tests with coverage
go tool github.com/onsi/ginkgo/v2/ginkgo -r --cover

# Watch mode
ginkgo watch -r

# Run specific test
go tool github.com/onsi/ginkgo/v2/ginkgo -r --focus="ItemsAPI"
```

### Frontend Tests

```bash
cd frontend

# Run all tests
npm test

# Run tests in watch mode
npm test -- --watch

# Generate coverage report
npm test -- --code-coverage

# Run specific test
npm test -- --include='**/auth.service.spec.ts'
```

### Test Coverage

- **Backend**: Comprehensive integration and E2E tests using Ginkgo/Gomega
- **Frontend**: 94.83% unit test coverage with Jasmine/Karma

## 📦 Deployment

### Docker Deployment

```bash
# Build Docker image
docker build -t diary .

# Run with Docker Compose
docker-compose up -d

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

### Manual Deployment

1. **Build the application**:

   ```bash
   make build
   ```

2. **Set environment variables**:

   ```bash
   export GB_USERS="user@example.com:bcrypt_hash"
   export GB_DBPATH="/path/to/diary.db"
   export GB_ASSETPATH="/path/to/assets"
   export GB_ALLOWEDORIGINS="https://your-frontend-domain.com"
   ```

3. **Run the backend**:

   ```bash
   ./backend/bin/diary server
   ```

4. **Serve the frontend**:
   - Deploy `frontend/dist/Diary.FE/` to a web server (nginx, Apache, etc.)
   - Configure the web server to serve `index.html` for all routes (SPA routing)

### Production Checklist

- [ ] Use HTTPS for all connections
- [ ] Set strong `GB_JWTSECRET`
- [ ] Configure proper CORS origins
- [ ] Set secure cookie attributes
- [ ] Enable rate limiting
- [ ] Set up database backups
- [ ] Configure log rotation
- [ ] Monitor application health
- [ ] Set up error tracking

## 🤝 Contributing

### Development Guidelines

Please read the comprehensive development guidelines before contributing:

- **[Backend Guidelines](.augment/rules/backend-guidelines.md)** - Go development standards
- **[Frontend Guidelines](.augment/rules/frontend-guidelines.md)** - Angular development standards

### Contribution Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Follow the development guidelines
4. Write tests for new functionality
5. Ensure all tests pass (`make test`)
6. Ensure linting passes (`make lint`)
7. Commit your changes (`git commit -m 'Add amazing feature'`)
8. Push to the branch (`git push origin feature/amazing-feature`)
9. Open a Pull Request

## 📄 License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Angular](https://angular.dev)
- Backend powered by [Go](https://golang.org)
- Database: [SQLite](https://www.sqlite.org)
- API specification: [OpenAPI 3.0](https://www.openapis.org)
- Testing: [Ginkgo](https://onsi.github.io/ginkgo/), [Gomega](https://onsi.github.io/gomega/), [Jasmine](https://jasmine.github.io/), [Karma](https://karma-runner.github.io/)

## 📞 Support

For issues, questions, or contributions, please:

1. Check existing [issues](../../issues)
2. Review the [documentation](#-documentation)
3. Create a new issue with detailed information

---

**Made with ❤️ for privacy-conscious diary keeping**
