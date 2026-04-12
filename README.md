# TaskFlow

## 1. Overview
TaskFlow is a robust, full-stack project management application designed to help teams organize projects, assign tasks, and track their progress efficiently. 

**Tech Stack Used:**
* **Backend:** Go (Golang) using `go-chi` for lightweight routing.
* **Database:** PostgreSQL.
* **Database Driver:** `pgx` (pure Go Postgres driver, no bulky ORMs).
* **Frontend:** React (TypeScript) built with Vite.
* **Infrastructure:** Docker and Docker Compose for seamless containerized deployments.

## 2. Architecture Decisions
* **Go + `chi` over bloated frameworks:** I chose standard library-focused Go with `chi` to maintain absolute control over the HTTP layer. It is incredibly fast, explicit, and avoids the "magic" that heavier frameworks like Gin or Fiber introduce.
* **Raw SQL (`pgx`) instead of an ORM (like GORM):** To ensure maximum performance and give myself full control over database queries (like complex joins and aggregations for Project Stats), I opted against using an ORM. Writing raw SQL via `pgx` prevents N+1 query issues and makes the database interactions perfectly transparent.
* **Domain Structure (Handlers -> Repository):** I structured the backend into handlers (HTTP logic) and repositories (Database logic). 
  * *Tradeoff/Omission:* I intentionally left out a middle "Service" layer initially to avoid over-engineering. For an application of this current scope, routing directly from handler to repository is highly efficient. As business rules become more complex, a service interface layer could be introduced later.
* **Strict Compile-Time Typing:** Custom string types (e.g., `TaskStatus`) are enforced deeply within the Go codebase to prevent raw string bugs.

## 3. Running Locally
To get the application up and running on your local machine using Docker:

```bash
git clone https://github.com/chandratsekharttt-CS/task-flow-chandra-sekhar.git
cd task-flow-chandra-sekhar
cp .env.example .env
docker compose up --build
```
* **Frontend Application:** http://localhost:3000
* **Backend API:** http://localhost:8080

### For Co-Developers (Native Setup Without Docker)
If you are a co-developer looking to actually write code, edit handlers, and run the server natively on your machine without waiting for Docker to rebuild, you will need Go installed. 

Once your Postgres database is running, open a new terminal:
```bash
cd backend
go mod tidy          # Cleans and resolves all package dependencies for your IDE
go run ./cmd/server/main.go
```

## 4. Running Migrations
Database migrations **run automatically on startup** via the `golang-migrate` package embedded directly into the Go backend's `main.go`. When `docker compose up` is executed, the API service applies any pending `.sql` files automatically before starting the web server.

If you ever need to run them manually from your host machine (assuming you have `golang-migrate` installed):
```bash
migrate -path backend/migrations -database "postgresql://taskflow:taskflow_secret@localhost:5432/taskflow?sslmode=disable" up
```

## 5. Test Credentials
The database automatically seeds itself on its very first run. You can log in immediately to test the functionality without needing to register a new account:

* **Email:** `test@example.com`
* **Password:** `password123`

## 6. API Reference

> **Note:** A complete React frontend application has been built alongside this API. Most of the endpoints listed below can be easily tested and interacted with directly through the frontend UI (available at `http://localhost:3000` when running Docker) without needing a tool like Postman.

### Authentication
* `POST /api/auth/register` - Register a new user.
* `POST /api/auth/login` - Authenticate and receive a JWT.
* `GET /api/auth/me` - Get current authenticated user details.

### Projects
* `GET /api/projects` - List all projects owned by the user.
* `POST /api/projects` - Create a new project.
* `GET /api/projects/:id` - Get specific project details and all associated tasks.
* `PATCH /api/projects/:id` - Update project details.
* `DELETE /api/projects/:id` - Delete a project and cascade delete its tasks.
* `GET /api/projects/:id/stats` - Fetch aggregate statistics (task counts by status/assignee).

### Tasks
* `GET /api/projects/:id/tasks?page=1&limit=20` - Get paginated tasks.
* `GET /api/tasks/me` - Get all tasks assigned to the current user globally.
* `POST /api/projects/:id/tasks` - Add a task to a project.
* `PATCH /api/tasks/:id` - Update task details (status, assignee, etc).

**Example Request:** `POST /api/projects/:id/tasks`
```json
{
  "title": "Setup CI/CD Pipeline",
  "status": "todo",
  "priority": "high",
  "assignee_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

## 7. What You'd Do With More Time
If given more time to expand the scope of this project, I would focus on several key product features and architectural improvements:

### Product & Security Enhancements
* **Email OTP Verification:** To ensure data integrity and prevent spam, I would implement a One-Time Password (OTP) system sent via email during the registration flow to verify that the provided email address is valid before fully activating the account.
* **Admin Approval Workflow:** Since this application is tailored for organizational team management, I would add an admin permission layer to the registration process. Instead of allowing any random person to create a fully active account, new sign-ups would be placed in a "pending" state until a verified company Admin approves them.
* **Centralized Global Dashboard:** I would build a comprehensive "Home" dashboard upon login that aggregates data across the entire organization. This would give users instant visibility into upcoming cross-team projects, a timeline of open tasks, and a clear breakdown of who is currently working on what, significantly improving high-level project visibility.

### Architectural Improvements
* **Service Layer & Interface Driven Testing:** I would extract the direct repository calls out of the handlers and put them into a formal "Service" layer defined by Interfaces. This would allow me to easily mock the database logic and write comprehensive unit tests for the API controllers.
* **Advanced Permissions (RBAC):** Currently, permissions are locked to "Project Owners" or "Assignees". Alongside the Admin approval workflow mentioned above, I would implement a full Role-Based Access Control (RBAC) system with a `project_members` join table to allow distinct "Admin", "Editor", and "Viewer" roles on specific projects.
* **Redis Caching:** The `/stats` endpoint recalculates task counts directly against the database every time it is hit. With more time, especially if powering a Centralized Global Dashboard, I would wrap this in a Redis caching layer to heavily reduce the database load on high-traffic queries.
* **API Documentation:** Automatically generate Swagger/OpenAPI documentation directly from the Go struct tags using `swaggo` instead of manually maintaining endpoint lists.
