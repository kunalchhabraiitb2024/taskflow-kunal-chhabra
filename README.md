# TaskFlow

A task management system with authentication, project management, and task tracking. Full-stack app with a Docker Compose workflow suitable for local review or demos.

---

## 1. Overview

**TaskFlow** lets teams create projects, add tasks to them, and assign tasks to members. Features:

- Register / login with JWT authentication
- Create and manage projects
- Create, edit, assign, filter, and delete tasks
- Optimistic UI for status changes
- Responsive design with dark mode

**Stack**:

| Layer | Technology |
|---|---|
| Backend | Go 1.25 (`go.mod`) · Chi router · pgx v5 (no ORM) |
| Database | PostgreSQL 16 |
| Migrations | golang-migrate (auto-run on startup) |
| Auth | JWT (HS256, 24h expiry) · bcrypt (configurable cost, default 12) |
| Frontend | React 18 · TypeScript · Vite 6 |
| UI | shadcn/ui · Tailwind CSS |
| Infrastructure | Docker Compose · multi-stage Dockerfiles |

---

## 2. Architecture Decisions

### Backend: Handler → Service → Repository

Three distinct layers with clean interfaces:

- **Handlers** deal only with HTTP: parse requests, call services, write responses.
- **Services** hold business logic and authorization checks (e.g., "only the project owner can delete").
- **Repositories** hold SQL. No ORM — raw `pgx` queries give precise control over JOINs, CTEs, and enum types.

This makes each layer independently testable and avoids the "god handler" anti-pattern.

### No ORM

`pgx` with raw SQL rather than GORM/ent. The schema uses PostgreSQL enum types (`task_status`, `task_priority`) which most ORMs handle awkwardly. Direct SQL also makes query intent obvious to reviewers.

### JWT in Authorization header only

No cookies. Token stored in `localStorage` and attached via Axios request interceptor. Simple and stateless — appropriate for this scope.

### Migrations auto-run on container start

`golang-migrate` is called from `main.go` before the HTTP listener starts. Zero manual steps. Both `up` and `down` migrations exist for every change.

### Seed runs from Go, not a shell script

`seed/seed.sql` is executed via `pgconn.Exec` (simple query protocol, supports multi-statement SQL) on startup. All `INSERT`s use `ON CONFLICT DO NOTHING` — idempotent and safe to re-run.

### `created_by` on tasks

Not in the original schema but required to enforce the delete rule: *"project owner or task creator can delete."* Added as a non-nullable FK to `users`.

### Intentional omissions

- **No refresh tokens** — 24h access tokens are sufficient for this scope.
- **No WebSocket** — real-time updates would require significant additional infrastructure.
- **Assignee name not resolved in task responses** — tasks return `assignee_id` (UUID). The frontend could fetch user details separately; for this scope it's acceptable.
- **No email verification** — out of scope.

---

## 3. Running Locally

Prerequisites: **Docker** and **Docker Compose** (nothing else needed).

From the repository root (the directory that contains `docker-compose.yml`):

```bash
cp .env.example .env
# Edit .env if needed — defaults match docker-compose and seed data.
docker compose up
```

If you cloned from GitHub, `cd` into the cloned folder first, then run the commands above.

**Node / npm (optional, for local frontend dev without Docker):** `package.json` lives under **`frontend/`**, not the repo root. Use:

```bash
cd frontend
npm install
npm run dev
```

Running `npm install` from the repository root will fail with `ENOENT` / missing `package.json` — always `cd frontend` first (or use `npm install --prefix frontend` from the root).

- Frontend: http://localhost:3000
- API: http://localhost:8080
- Health check: http://localhost:8080/health

The first `docker compose up` will:
1. Start PostgreSQL
2. Build the Go binary (multi-stage)
3. Run all migrations automatically
4. Insert seed data (test user, project, 3 tasks)
5. Build and serve the React app via Nginx

---

## 4. Running Migrations

Migrations run **automatically on container start**. No manual steps required.

If you need to run them manually (e.g., during local development without Docker):

```bash
# Install golang-migrate CLI
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run up
migrate -path backend/migrations -database "$DATABASE_URL" up

# Run down (rollback one step)
migrate -path backend/migrations -database "$DATABASE_URL" down 1
```

---

## 5. Test Credentials

Seed data is automatically inserted on first startup:

```
Email:    test@example.com
Password: password123
```

The seed includes:
- 1 user (above credentials)
- 1 project: "Website Redesign"
- 3 tasks with statuses: `done`, `in_progress`, `todo`

---

## 6. API Reference

All protected endpoints require `Authorization: Bearer <token>`.

All error responses use JSON:
```json
{ "error": "not found" }
{ "error": "validation failed", "fields": { "email": "is required" } }
```

### Auth

| Method | Endpoint | Description |
|---|---|---|
| POST | `/auth/register` | Register. Body: `{ name, email, password }`. Returns `{ token, user }`. |
| POST | `/auth/login` | Login. Body: `{ email, password }`. Returns `{ token, user }`. |

### Projects

| Method | Endpoint | Description |
|---|---|---|
| GET | `/projects` | List user's projects. Supports `?page=&limit=`. |
| POST | `/projects` | Create project. Body: `{ name, description? }`. |
| GET | `/projects/:id` | Get project + all its tasks. |
| PATCH | `/projects/:id` | Update name/description (owner only). |
| DELETE | `/projects/:id` | Delete project + tasks (owner only). Returns `204`. |

### Tasks

| Method | Endpoint | Description |
|---|---|---|
| GET | `/projects/:id/tasks` | List tasks. Supports `?status=todo\|in_progress\|done`, `?assignee=uuid`, `?page=&limit=`. |
| POST | `/projects/:id/tasks` | Create task. Body: `{ title, description?, status?, priority?, assignee_id?, due_date? }`. |
| PATCH | `/tasks/:id` | Update any task field (partial). |
| DELETE | `/tasks/:id` | Delete task (project owner or task creator only). Returns `204`. |
| GET | `/projects/:id/stats` | Task counts by status and assignee. |

### Paginated response shape

```json
{
  "data": [ /* items */ ],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 47,
    "total_pages": 3
  }
}
```

### Example: Register

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Jane Doe","email":"jane@example.com","password":"secret123"}'
```

```json
{
  "token": "<jwt>",
  "user": { "id": "uuid", "name": "Jane Doe", "email": "jane@example.com" }
}
```

### Example: Create a task

```bash
curl -X POST http://localhost:8080/projects/<project-id>/tasks \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"title":"Build landing page","priority":"high","due_date":"2026-05-01"}'
```

---

## 7. Running Integration Tests

Integration tests require a running PostgreSQL instance. They are skipped automatically when `TEST_DATABASE_URL` is not set.

```bash
# With docker compose running:
export TEST_DATABASE_URL="postgres://taskflow:taskflow_secret@localhost:5432/taskflow?sslmode=disable"
cd backend
go test ./tests/ -v -count=1
```

4 tests covering:
1. Register + login flow (including wrong-password → 401)
2. Project create and list (including unauthenticated → 401)
3. Task create + filter by status
4. Authorization: non-owner gets 403 on project delete

---

## 8. What You'd Do With More Time

**Already in this repo (was “future work”, now shipped):** Server-Sent Events for cross-tab task sync, **Board** view with `@dnd-kit` drag-between-columns, **projects** list pagination, and **stats** badges + `/projects/:id/stats`. See **§10** for how to try them.

**Harder real-time / scale:** The SSE broker is **in-process** (fine for one API replica). Multiple backend instances would need **Redis pub/sub**, **NATS**, or **Postgres `LISTEN/NOTIFY`** so every instance broadcasts task changes. **WebSockets** are an alternative if you need bidirectional messages or different client semantics.

**Kanban / tasks at scale:** **Ordering inside a column** is not persisted (no `sort_order` column). Adding one (plus migration and PATCH) would support “reorder within Done”. **Task pagination** on the project page still loads all tasks with the project (`GET /projects/:id`); for hundreds of tasks you’d paginate **`GET /projects/:id/tasks`** in the UI (infinite scroll or page controls).

**Proper team management**: A project is visible if the user **owns** it or has a **task assigned**. Production apps usually add **`project_members`**, roles (viewer/editor/owner), and invites.

**Assignee name resolution**: Tasks return `assignee_id` only. Embed assignee display name via a SQL `JOIN` to `users`, or resolve in a small user cache on the client.

**Auth & security**: **Refresh tokens** (short access + `httpOnly` refresh cookie) and **rate limiting** on `/auth/login` (and registration) for brute-force protection.

**Polish**: Full **WCAG** / screen-reader pass; **loading skeletons** instead of spinners; replace browser **`confirm()`** deletes with accessible confirmation dialogs.

**Data model nit**: Due dates are `DATE` in SQL but round-tripped through `time.Time` / ISO strings — a dedicated date-only type end-to-end would avoid timezone edge cases.

**Shortcuts still in place**:
- `confirm()` for task/project delete — fast to ship, weaker UX than a modal
- List/Board **List** view still uses the same delete pattern; Board cards use the grip handle for status moves only

---

## 9. Requirements alignment (backend, frontend, infra, README, evaluation)

This section maps **what reviewers usually expect** from a full-stack submission to **what this repository provides**. It mirrors the stack and scope described above.

### Backend

| Expectation | Where it is met |
|-------------|-----------------|
| REST API, JSON request/response | **§6** API reference; Chi routes in `backend/internal/router` |
| Authentication (register/login) and protected routes | **§6** `/auth/*`; JWT middleware on project/task routes |
| Password hashing (not plaintext) | bcrypt via `BCRYPT_COST` / **§2** |
| PostgreSQL persistence, sensible schema | Migrations in `backend/migrations/`; enums for task status/priority |
| Schema versioning / migrations | golang-migrate **up** + **down** files; auto-run on startup **§4** |
| Layered structure (handlers vs business vs data) | **§2** Handler → Service → Repository |
| Automated tests (API or integration) | **§7** — `backend/tests/` when `TEST_DATABASE_URL` is set |
| Containerized or documented run | **§3** — backend `Dockerfile`, env from `.env` / Compose |

### Frontend

| Expectation | Where it is met |
|-------------|-----------------|
| SPA with routing | React Router — `frontend/src/App.tsx` |
| Login / register and session handling | JWT in `localStorage`, Axios client **§2** |
| Project and task CRUD aligned with API | Pages under `frontend/src/pages/`, API under `frontend/src/api/` |
| Task filters / assignment as supported by API | Filters + forms (status, assignee, etc.) per **§6** |
| Responsive, usable UI | Tailwind + shadcn; **§1** dark mode |
| Optimistic feedback where appropriate | Task status updates **§1**; see `TaskCard` (optimistic + revert on error) |
| Production build | `npm run build`; multi-stage **frontend** Docker image **§3** |

### Infrastructure

| Expectation | Where it is met |
|-------------|-----------------|
| One-command local stack | `docker compose up` from repo root **§3** |
| Services: app DB, API, static/UI | `docker-compose.yml` — `db`, `backend`, `frontend` |
| Non-secret configuration template | **`.env.example`** (copy to `.env`); **`.env` gitignored** |
| Health / readiness for API | `GET /health` **§3** |
| Postgres health gating backend | `depends_on` + `condition: service_healthy` in Compose |

### README / documentation

| Expectation | Where it is met |
|-------------|-----------------|
| What the product does | **§1** Overview |
| How to run locally | **§3** Prerequisites, `cp .env.example`, `docker compose up`, URLs |
| How migrations work | **§4** |
| How to try the app quickly | **§5** seed credentials + seed contents |
| API surface for manual or automated checks | **§6** |
| How to run tests | **§7** |
| Honest limits and tradeoffs | **§8** |

### Before you submit (self-evaluation)

1. **Fresh clone:** `cp .env.example .env` → `docker compose up --build` → open **http://localhost:3000**, log in with **§5** credentials, create a project and a task.
2. **Secrets:** Confirm **`.env` is not committed** (`git status` should not list it).
3. **Tests (optional but strong):** With Compose DB up, set `TEST_DATABASE_URL` as in **§7** and run `go test ./tests/ -v -count=1` from `backend/`.
4. **Lint (optional):** From `frontend/`, `npm run lint`.

---

## 10. Optional “bonus” rubric (tests, pagination, DnD, real-time, dark mode, stats)

Typical +5-style bonuses and where to see them in this repo:

| Bonus | Implemented? | How to verify |
|-------|----------------|---------------|
| **Integration / API tests** | Yes | **§7** — `backend/tests/` with `TEST_DATABASE_URL` set |
| **Pagination** | Yes (projects list) | **Projects** page: **Previous / Next** when you have more projects than the page size (8 per page) |
| **Drag-and-drop (status columns)** | Yes | Open a project → **Board** view (default) → drag tasks by the **grip** icon between **To do / In progress / Done** |
| **Real-time updates (SSE)** | Yes | With the app open on a project, change tasks in another tab or another browser — the project view refreshes from **`GET /projects/:id/events`** (SSE with `Authorization` header via `@microsoft/fetch-event-source`) |
| **Dark mode (persistent)** | Yes | Navbar **sun/moon** toggle — theme stored in **`localStorage`** (`theme` key) |
| **Stats endpoint** | Yes | **Project detail** shows **Stats (API)** badges (counts by status); same data from **`GET /projects/:id/stats`** |

**Note:** Task **ordering** within a column is not persisted (no `sort_order` in the DB); dragging **between** columns updates **`status`** via the API. WebSockets are not used; **SSE** satisfies “real-time or SSE” style requirements.

---

If your course provides a separate rubric PDF, use **§9–10** as a cross-check: every row should still be true for your branch.
