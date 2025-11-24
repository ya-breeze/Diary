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
