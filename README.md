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

### Frontend (Next.js)

- **Framework**: Next.js 16+ (App Router)
- **Language**: TypeScript 5+
- **State Management**: Zustand
- **Styling**: Tailwind CSS
- **Testing**: Vitest / Playwright
- **Location**: `next-frontend/`

### Layered Architecture

```
┌─────────────────────────────────────────┐
│         Frontend (Next.js)              │
│  ┌─────────────────────────────────┐   │
│  │  Components & Hooks             │   │
│  │  (React, Zustand, Fetch)        │   │
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
| `GB_DATAPATH`          | Directory for storing data (DB and assets)   | `./diary-data`          |
| `GB_JWTSECRET`           | Secret key for JWT token signing             | Auto-generated          |
| `GB_ISSUER`              | JWT issuer identifier                        | `diary-backend`         |
| `GB_ALLOWEDORIGINS`      | Comma-separated list of allowed CORS origins | `http://localhost:3000` |
| `GB_MAXPERFILESIZEMB`    | Maximum size per uploaded file (MB)          | `25`                    |
| `GB_MAXBATCHFILES`       | Maximum number of files per batch upload     | `10`                    |
| `GB_MAXBATCHTOTALSIZEMB` | Maximum total size per batch upload (MB)     | `100`                   |
| `GEMINI_API_KEY`         | Google Gemini API key enabling AI tag suggestion. When unset, AI tagging is fully disabled and the app behaves as before. | Unset (feature off) |

#### AI Tag Suggestion

When `GEMINI_API_KEY` is set, families can opt in to AI-assisted tag suggestion
(Profile → AI tagging). With it enabled, the entry editor offers a "Suggest tags"
button and debounced auto-suggestions drawn from the family's existing tag
vocabulary. Suggestions are never applied automatically — they appear as chips the
user accepts. The feature requires both the server key and the per-family toggle.
