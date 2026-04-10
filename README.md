# TaskFlow

A full-stack task management system built with **Go** (backend), **React + TypeScript** (frontend), and **PostgreSQL** (database).

## 🏗️ Tech Stack

| Layer | Technology | Rationale |
|---|---|---|
| **Backend** | Go 1.22, Chi router | Lightweight, stdlib-compatible, idiomatic REST routing |
| **Database** | PostgreSQL 16 | Robust relational DB with UUID support and enums |
| **Migrations** | golang-migrate | Industry standard, supports up/down SQL files |
| **Auth** | JWT (HS256, 24h expiry) | Stateless authentication, bcrypt password hashing (cost 12) |
| **Frontend** | React 18, TypeScript, Vite | Fast builds, type safety, modern DX |
| **Styling** | Custom CSS design system | Full control, dark/light theme, no heavy dependencies |
| **Containerization** | Docker, Docker Compose | One-command full-stack deployment |

## 🚀 Quick Start

### Prerequisites
- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)

### Run the Full Stack

```bash
# 1. Clone the repository
git clone <repo-url>
cd taskflow-chandu

# 2. Copy environment variables
cp .env.example .env

# 3. Start everything
docker compose up --build
```

That's it! The app will be available at:
- **Frontend**: http://localhost:3000
- **API**: http://localhost:8080
- **Health Check**: http://localhost:8080/health

### Seed Credentials

The database is automatically seeded with test data on first run:

| Field | Value |
|---|---|
| **Email** | `testuser@taskflow.com` |
| **Password** | `password123` |

This creates 1 user, 1 project ("Sample Project"), and 3 tasks with different statuses.

## 📁 Project Structure

```
taskflow-chandu/
├── backend/
│   ├── cmd/server/main.go        # Entry point, router, graceful shutdown
│   ├── internal/
│   │   ├── config/               # Env-based configuration
│   │   ├── database/             # PostgreSQL connection pool (pgx)
│   │   ├── handlers/             # HTTP handlers + integration tests
│   │   ├── middleware/           # JWT auth, CORS, request logging
│   │   ├── models/              # Data models + request/response types
│   │   ├── repository/          # Database queries (no ORM)
│   │   └── validator/           # Structured field validation
│   ├── migrations/              # SQL migration files (up + down)
│   ├── seed/                    # Seed data documentation
│   └── Dockerfile               # Multi-stage build
├── frontend/
│   ├── src/
│   │   ├── api/client.ts        # Axios + JWT interceptor
│   │   ├── components/          # Reusable UI components
│   │   ├── context/             # Auth state management
│   │   ├── pages/               # Route-level pages
│   │   ├── types/               # TypeScript interfaces
│   │   └── index.css            # Design system (dark/light themes)
│   ├── nginx.conf               # Production reverse proxy config
│   └── Dockerfile               # Multi-stage build (Node → Nginx)
├── docker-compose.yml           # Full stack orchestration
├── .env.example                 # Environment variables template
└── README.md
```

## 🔌 API Reference

### Authentication

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| POST | `/api/auth/register` | No | Register (name, email, password) |
| POST | `/api/auth/login` | No | Login, returns JWT |
| GET | `/api/auth/me` | Yes | Get current user info |

### Projects

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| GET | `/api/projects` | Yes | List user's projects (paginated) |
| POST | `/api/projects` | Yes | Create project |
| GET | `/api/projects/:id` | Yes | Get project + tasks |
| PATCH | `/api/projects/:id` | Yes | Update project (owner only) |
| DELETE | `/api/projects/:id` | Yes | Delete project + tasks (owner only) |
| GET | `/api/projects/:id/stats` | Yes | Task counts by status/assignee |

### Tasks

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| GET | `/api/projects/:id/tasks` | Yes | List tasks (?status=&assignee=&page=&limit=) |
| POST | `/api/projects/:id/tasks` | Yes | Create task |
| PATCH | `/api/tasks/:id` | Yes | Update task fields |
| DELETE | `/api/tasks/:id` | Yes | Delete (owner or creator only) |

### Error Responses

```json
// 400 Validation
{ "error": "validation failed", "fields": { "email": "is required" } }

// 401 Unauthenticated
{ "error": "unauthorized" }

// 403 Forbidden
{ "error": "forbidden" }

// 404 Not Found
{ "error": "not found" }
```

## 🎨 Frontend Features

- **Auth**: Login/Register with JWT persistence across page refreshes
- **Projects**: Card grid with create modal
- **Tasks**: Kanban board (Todo / In Progress / Done columns)
- **Optimistic UI**: Task status changes update instantly, revert on error
- **Dark Mode**: Toggle with session persistence
- **Responsive**: Works at 375px (mobile) through 1280px+ (desktop)
- **Error Handling**: Inline validation errors, alert messages, loading spinners
- **Empty States**: Friendly messages when no data exists

## ⚙️ Environment Variables

| Variable | Default | Description |
|---|---|---|
| `POSTGRES_USER` | `taskflow` | PostgreSQL username |
| `POSTGRES_PASSWORD` | `taskflow_secret` | PostgreSQL password |
| `POSTGRES_DB` | `taskflow` | Database name |
| `DATABASE_URL` | (composed) | Full connection string |
| `JWT_SECRET` | (dev default) | **Change in production!** |
| `API_PORT` | `8080` | Backend server port |
| `BCRYPT_COST` | `12` | bcrypt hash cost factor |
| `SEED_DB` | `true` | Auto-seed on first startup |

## 🧪 Running Tests

Integration tests require a running PostgreSQL instance:

```bash
# Start just the database
docker compose up db -d

# Run tests (from backend directory)
cd backend
TEST_DATABASE_URL="postgres://taskflow:taskflow_secret@localhost:5432/taskflow?sslmode=disable" go test ./... -v
```

## 🏛️ Architecture Decisions

1. **No ORM**: All SQL is hand-written in the repository layer. This gives full control over queries and avoids ORM abstraction leaks.
2. **PATCH semantics**: Update endpoints accept partial JSON. Only fields present in the request body are updated.
3. **Embedded migrations**: Migration files are read at runtime from the filesystem. The Docker image includes them alongside the binary.
4. **Custom CSS**: No component library dependency. A CSS custom properties system provides dark/light theming with minimal overhead.
5. **Idempotent seeding**: The seed function checks if data exists before inserting, making container restarts safe.

## 📝 License

MIT
