# CuraClinic

CuraClinic is a React/Vite patient portal and a Go/Gin clinic API in one repository. The backend now implements authentication, RBAC with per-user overrides, doctor schedules, Redis slot holds, dynamic availability, appointments with an audit trail, and admin tooling.

## Project layout

- `frontend/` — React application and Vite configuration (still uses demo data)
- `backend/` — Gin API, PostgreSQL migrations, Redis stores and backend docs
- `.github/workflows/backend.yml` — backend CI (tests, sqlc check, image build/push, Dokploy trigger)

## Backend quick start

Requirements: Go 1.25+, PostgreSQL 18 binaries. Redis is optional locally (`REDIS_ADDR=memory`).

```powershell
cd backend
powershell -ExecutionPolicy Bypass -File .\scripts\dev-db.ps1 start   # local Postgres on 127.0.0.1:55432
Copy-Item .env.example .env
go run ./cmd/migrate up
go run ./cmd/seed --demo
go run ./cmd/api
```

`GET /api/health` reports liveness and `GET /api/ready` reports PostgreSQL/Redis readiness (Redis shows `disabled` in local memory mode). Demo logins and every endpoint are documented in [`backend/README.md`](backend/README.md). The local admin/doctor/patient passwords live in the gitignored `backend/.env`.

With Docker (real Redis included):

```powershell
cd backend
Copy-Item .env.example .env
docker compose up --build
```

## Frontend

Copy `frontend/.env.example` to `frontend/.env.local` and run:

```powershell
cd frontend
pnpm install
pnpm dev
```

The frontend reads `VITE_API_BASE_URL` (default `http://localhost:8080/api`) through `frontend/src/lib/api.ts`. Copy `frontend/.env.example` to `frontend/.env.local` if you need to override it. Keep credentials and production secrets in the deployment environment, never in `VITE_*` variables because frontend variables are public in the built bundle.

## Authentication

- `/login` signs in every role and redirects by RBAC: admin → `/admin`, doctor → `/daily-agenda`, patient → `/`. Front-desk staff use an admin-scoped custom role and land in the admin area.
- `/signup` creates a patient account and signs in immediately.
- Access and refresh tokens are stored in localStorage; refresh tokens rotate server-side and the API client uses a single-flight refresh to avoid token-family reuse.
- Sidebars and routes are permission-aware. Patients see the patient portal, doctors see agenda/schedule, and clinic staff see the admin area with only the sections their permissions allow.

## Tests

```powershell
cd backend
go test ./...
# with the local dev cluster:
$env:TEST_DATABASE_URL="postgres://postgres@127.0.0.1:55432/curaclinic?sslmode=disable"
go test -race ./...
```

## Deployment

The backend ships a multi-stage `Dockerfile` (API + migration + seed binaries, non-root user, `/api/ready` health check). GitHub Actions builds and pushes to GHCR on `main` and can trigger a Dokploy deploy webhook. Configure the environment variables listed in `backend/README.md`, run `/app/migrate up` as the pre-deploy command, and keep PostgreSQL/Redis on persistent volumes.
