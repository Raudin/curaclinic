# CuraClinic Backend

Go 1.27 + Gin API for clinic scheduling: JWT authentication, database-backed RBAC with per-user overrides, recurring doctor schedules, Redis slot holds, dynamic availability and appointment history.

## Stack

| Concern | Choice |
|---|---|
| HTTP | Gin |
| Database | PostgreSQL 18 + pgx v5 + sqlc |
| Cache / holds | Redis 7 + go-redis v9 |
| Auth | JWT access tokens + rotating opaque refresh tokens |
| Authorization | Roles + permissions + per-user allow/deny overrides (custom middleware) |
| Validation | Gin binding + go-playground/validator + service-layer rules |
| Migrations | goose v3 with embedded SQL |
| Logging | zap (structured) |
| Config | viper + environment variables |
| Tests | testify, miniredis, real-Postgres integration tests |

## Quick start

### Option A — Docker Compose (Postgres + Redis + API)

```powershell
cd backend
Copy-Item .env.example .env
docker compose up --build
docker compose run --rm migrate up   # if the migrate service was skipped
docker compose run --rm --entrypoint /app/seed api --demo
```

Then:

```powershell
curl http://localhost:8080/api/health
curl http://localhost:8080/api/ready
```

### Option B — Local Go process (no Redis required)

Requirements: Go 1.27+, PostgreSQL 18 binaries (the project-local cluster script uses them).

```powershell
cd backend
powershell -ExecutionPolicy Bypass -File .\scripts\dev-db.ps1 start   # local Postgres on 127.0.0.1:55432
Copy-Item .env.example .env      # .env.example defaults to the local cluster + REDIS_ADDR=memory
go mod download
make migrate
make seed
make run
```

The local `.env` uses `REDIS_ADDR=memory`, which enables the development-only in-memory hold store; permissions are read from PostgreSQL and rate limiting is disabled. Production and Docker still require a real Redis instance. The dev database is trust-auth and listens on loopback only under `%LOCALAPPDATA%\CuraClinic\pgdata`. Stop it with `dev-db.ps1 stop`; remove it with `dev-db.ps1 reset -Force`.

If you prefer an existing PostgreSQL service, point `DATABASE_URL` at that database instead (the migration/seed commands are unchanged).

Seeded demo logins (when `--demo` is used):

| Role | Email | Password |
|---|---|---|
| Admin | `admin@curaclinic.local` | `BOOTSTRAP_ADMIN_PASSWORD` in `backend/.env` |
| Doctor | `naomi.kamau@curaclinic.local` | `DEMO_PASSWORD` in `backend/.env` |
| Patient | `patient@curaclinic.local` | `DEMO_PASSWORD` in `backend/.env` |

New patients can also self-register from `/signup` in the frontend.

## Quality gates

Run these from `backend/`. The `Makefile` is the single source of truth and the same commands run in CI.

| Target | What it does |
|---|---|
| `make fmt` | Formats `cmd/`, `internal/` and `migrations/` with gofmt |
| `make vet` | Runs `go vet ./...` |
| `make lint` | Runs `golangci-lint` with `backend/.golangci.yml` |
| `make test` | Race-enabled tests (`-shuffle=on`); needs a C compiler for `-race` |
| `make test-integration` | Same tests against `TEST_DATABASE_URL` (repository/router integration suites) |
| `make sqlc` | Regenerates `internal/repository/postgres/dbgen` from `queries/` |
| `make build` | Builds every package |
| `make run` | Starts the API |
| `make migrate` | Applies pending migrations |
| `make seed` | Seeds demo data |

Install the linter once with `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0`. Database integration tests are enabled by setting `TEST_DATABASE_URL`; without it they skip and `make test` only runs the unit suites. The targets require GNU make, and `make test` additionally requires a C compiler for the race detector (CI runs both on Ubuntu); on Windows use Git Bash/MSYS2 or WSL.

## Configuration

All settings are environment variables; `.env` is optional and ignored by Git.

| Variable | Default | Notes |
|---|---|---|
| `APP_ENV` | `development` | `production` switches zap to JSON |
| `API_PORT` | `8080` | |
| `GIN_MODE` | `debug` | use `release` in production |
| `LOG_LEVEL` | `info` | |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:5173` | CSV allow-list |
| `TRUSTED_PROXIES` | empty | CSV of proxy CIDRs allowed to set `X-Forwarded-For`; set this to your Dokploy/Traefik network |
| `DATABASE_URL` | — | required |
| `REDIS_ADDR` | `localhost:6379` | required in production; set `memory` locally to use in-memory holds |
| `REDIS_PASSWORD`, `REDIS_DB`, `REDIS_TLS` | empty / `0` / `false` | |
| `JWT_SECRET` | — | required, **at least 32 characters** |
| `JWT_ISSUER`, `JWT_AUDIENCE` | `curaclinic-api`, `curaclinic-web` | |
| `JWT_ACCESS_TTL` | `15m` | |
| `JWT_REFRESH_TTL` | `720h` | |
| `JWT_SESSION_IDLE_TTL` | `30m` | signing out after this much inactivity; resets on every token refresh |
| `BCRYPT_COST` | `12` | 10–14 |
| `CLINIC_TIMEZONE` | `Africa/Nairobi` | IANA zone used as the default |
| `BOOKING_LEAD_TIME` | `30m` | minimum notice for bookings and reschedules |
| `BOOKING_HORIZON_DAYS` | `90` | how far ahead availability is exposed |
| `HOLD_TTL` | `5m` | Redis slot hold lifetime |
| `RBAC_CACHE_TTL` | `60s` | permission cache TTL |
| `RATE_LIMIT_GLOBAL_PER_MIN` | `240` | per client IP |
| `RATE_LIMIT_AUTH_PER_15MIN` | `20` | login/refresh per client IP |
| `BOOTSTRAP_ADMIN_*` | — | used by `cmd/seed` |
| `DEMO_PASSWORD` | `DemoPass-12345` | used by `cmd/seed --demo` |

## Commands

```powershell
go run ./cmd/api                 # start the API
go run ./cmd/migrate up          # apply migrations
go run ./cmd/migrate status      # show applied migrations
go run ./cmd/migrate down        # roll back one migration
go run ./cmd/seed                # create the bootstrap admin
go run ./cmd/seed --demo         # bootstrap admin + demo doctors/patient/hours
.\scripts\dev-db.ps1 start       # project-local PostgreSQL on port 55432
.\scripts\dev-db.ps1 stop        # stop the local cluster
go test ./...                    # unit tests
$env:TEST_DATABASE_URL="postgres://postgres@127.0.0.1:5432/curaclinic_test?sslmode=disable"
go test -race ./...              # unit + Postgres integration tests
```

### sqlc

Queries live in `internal/repository/postgres/queries`, generated code in `internal/repository/postgres/dbgen` (committed).

```powershell
go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1 generate
```

## Architecture

```
cmd/api            HTTP entrypoint + graceful shutdown + wiring
cmd/migrate        goose runner over embedded migrations
cmd/seed           bootstrap admin and optional demo data
internal/config    viper configuration + fail-fast validation
internal/router    middleware order and route table
internal/handlers  thin Gin adapters, DTO mapping, path/query parsing
internal/service   business rules (availability, holds, appointments, RBAC, auth)
internal/repository/postgres  pgx repos + sqlc queries + transactions + audit
internal/cache     Redis holds, permission cache
internal/middleware request id, recovery, access log, CORS, rate limit, auth, RBAC
internal/domain    entities, error codes, permission catalog, audit actions
internal/httpx     JSON envelopes and error mapping
migrations         goose SQL (embedded)
```

**Middleware order:** request ID → panic recovery → structured access log → CORS → Redis rate limit → auth → permission guard → handler. Audit entries are written inside the same database transaction as the mutation.

## Database

Tables: `users`, `patients`, `doctors`, `admins`, `auth_sessions`, `roles`, `role_scopes`, `permissions`, `role_permissions`, `doctor_roles`, `admin_roles`, `doctor_permission_overrides`, `admin_permission_overrides`, `doctor_working_hours`, `schedule_blocks`, `appointments`, `appointment_events`, `audit_logs`.

Patient access comes from a linked patient record and has no RBAC grants. Doctor and admin roles and direct permission overrides are attached to their respective records. Shared identity fields such as date of birth and avatar are stored on `users`; doctor-specific clinical and scheduling details are stored on `doctors`.

Integrity guarantees:

- Partial unique indexes allow soft-deleted grants to be re-created.
- `EXCLUDE USING gist` constraints prevent overlapping **confirmed** appointments per doctor and per patient.
- `schedule_blocks` prevents overlapping active blocks per doctor.
- `btree_gist` is enabled in the first migration.
- Cancelling an appointment frees the slot automatically (the exclusion predicate only covers `status = 'confirmed'`).

### RBAC resolution

`deny` override → `allow` override → role grant → deny.

Effective permissions are resolved in SQL (`ListEffectivePermissionCodes`) and cached in Redis as `rbac:perms:{version}:{userID}`. Any role/permission/override change increments `rbac:version`, invalidating all cached snapshots. Auth middleware always re-reads the user from PostgreSQL, so deactivation takes effect immediately; permissions come from the versioned cache.

Guards: the last active admin cannot be demoted or deactivated, and system roles cannot be deleted.

## Redis slot holds

| Key | Purpose |
|---|---|
| `hold:{doctorID}:{YYYY-MM-DD}:{HHMM}` | hold payload (token, patient, expiry) with TTL |
| `holds:{doctorID}:{YYYY-MM-DD}` | ZSET index of hold keys scored by expiry |
| `hold:token:{uuid}` | maps an opaque hold token to its hold key |

Creation is a single Lua script: `SET NX PX` + `ZADD` + token mapping. Release is a Lua script that verifies ownership, deletes the hold, removes the index member and deletes the token mapping. Availability prunes expired index members lazily and verifies each member still exists.

If Redis loses data, the database exclusion constraints still prevent double booking.

## Booking rules (enforced in services, not only handlers)

1. Slot must be in the future and at least `BOOKING_LEAD_TIME` away.
2. Slot must fall inside the doctor's working hours for that weekday and align to the slot grid.
3. Slot must not overlap a schedule block.
4. Slot must not overlap a confirmed appointment.
5. Patient self-booking requires an active hold owned by the caller.
6. Reschedules validate the same rules, take a short-lived hold, and update the appointment in a transaction with an `appointment_events` record.
7. Cancellation requires a reason, blocks double cancellation, and records an event.

Staff with `appointments.create_manual` can book without a hold; database constraints remain the final authority.



## API summary

Base path: `/api`. Responses use `{ "data": ... }`; lists add `pagination`; errors use `{ "error": { code, message, details, request_id } }`.

| Method | Path | Permission |
|---|---|---|
| GET | `/health` | public |
| GET | `/ready` | public (DB + Redis) |
| POST | `/auth/login`, `/auth/register`, `/auth/refresh`, `/auth/logout` | public (rate limited) |
| GET | `/auth/me` | authenticated |
| POST | `/auth/change-password` | authenticated |
| GET | `/me/profile` | authenticated |
| PATCH | `/me/profile` | `profiles.update_own` |
| GET | `/me/schedule` | `doctor.schedule.view_own` |
| PUT | `/me/working-hours` | `doctor.schedule.define_working_hours_own` or `admin.schedule.define_working_hours_any` |
| GET/POST | `/me/schedule-blocks` | `doctor.schedule.view_own` / `doctor.schedule.block_time_own` |
| DELETE | `/me/schedule-blocks/:id` | `doctor.schedule.remove_block_own` |
| GET | `/doctors`, `/doctors/:id` | patient `doctors.view`, `doctor.patients.view` or `admin.users.view` |
| POST | `/doctors` | `admin.users.create` |
| PUT | `/doctors/:id/profile` | `doctor.profile.update_own` or `admin.users.update_profiles` |
| GET/PUT | `/doctors/:id/working-hours` | doctor own schedule permissions or `admin.schedule.define_working_hours_any` |
| GET/POST | `/doctors/:id/schedule-blocks` | doctor own schedule permissions or `admin.schedule.define_working_hours_any` |
| DELETE | `/schedule-blocks/:id` | `doctor.schedule.remove_block_own` or `admin.schedule.define_working_hours_any` |
| GET | `/doctors/:id/availability?date=YYYY-MM-DD` | patient booking, doctor booking or doctor schedule permissions |
| POST | `/holds` | patient `appointments.create_self` |
| DELETE | `/holds/:token` | authenticated (owner only) |
| GET | `/appointments`, `/appointments/:id`, `/appointments/:id/events` | `appointments.view_own` or `appointments.view_any` |
| POST | `/appointments` | patient `appointments.create_self` (hold required) |
| POST | `/appointments/manual` | `doctor.appointments.create_for_patient` or `admin.appointments.create_any` |
| PATCH | `/appointments/:id/cancel` | patient/doctor own permission or `admin.appointments.cancel_any` |
| PATCH | `/appointments/:id/reschedule` | patient/doctor own permission or `admin.appointments.reschedule_any` |
| GET/POST | `/users` | `admin.users.view` / `admin.users.create` |
| PATCH | `/users/:id/profile`, `/users/:id/status` | `admin.users.update_profiles` / `admin.users.create` |
| PUT | `/users/:id/roles` | `admin.users.assign_roles` |
| GET | `/users/:id/permissions` | `admin.users.view` |
| PUT/DELETE | `/users/:id/permissions[/:permissionId]` | `admin.permissions.manage` |
| GET | `/roles`, `/permissions` | `admin.roles.manage` / `admin.permissions.manage` |
| POST/PUT/DELETE | `/roles[/:id]` | `admin.roles.manage` |
| GET | `/audit-logs` | `admin.audit.view` |

Error codes: `validation_failed` (400), `unauthorized` (401), `forbidden` (403), `not_found` (404), `conflict` / `slot_unavailable` / `hold_expired` / `hold_not_owned` (409), `rate_limited` (429), `internal_error` (500).

List endpoints accept pagination (`limit`, `offset`). `GET /appointments` additionally supports `scope=upcoming|past|cancelled|all`, `from`, `to`, `order=asc|desc`, `doctor_id`, `patient_id` and `as=patient|doctor|admin` for users linked to multiple account records. `GET /users` supports `role` and `search`. Doctors are automatically scoped to their own diary; `admin.appointments.view_any` is required to see the whole clinic diary.

## Tests

- Unit tests cover slot generation/filtering, all booking rules, hold lifecycle (miniredis), cancel/reschedule permissions, RBAC last-admin protection and permission cache invalidation, error envelopes, auth and rate-limit middleware.
- Integration tests (enabled by `TEST_DATABASE_URL`) cover deny-wins permission resolution, the appointment exclusion constraints, cancel-frees-slot behaviour, role re-assignment and a full HTTP flow: login → availability → hold → book → duplicate conflict → list → events → cancel → slot released.

Docker is not required for unit tests. `make test` skips database integration tests when `TEST_DATABASE_URL` is unset; `make test-integration` runs them against the database in `TEST_DATABASE_URL`. `-race` requires cgo and a C compiler; CI runs it on Ubuntu.

## Deployment (Docker + GitHub Actions + Dokploy)

- `Dockerfile` builds `api`, `migrate` and `seed` binaries and runs as an unprivileged user with an `/api/ready` health check.
- `.github/workflows/backend.yml` verifies module checksums, formatting, vet, golangci-lint and govulncheck, runs race + coverage tests against Postgres/Redis services with a coverage floor, verifies generated sqlc code, pushes an image to GHCR on `main`, then optionally calls `DOKPLOY_WEBHOOK_URL`.
- In Dokploy: create the app from the image (or repo), add PostgreSQL and Redis services, set the environment variables above, and run `/app/migrate up` as the pre-deploy command. Set `JWT_SECRET` to a fresh 48-byte random value and use managed persistent volumes for Postgres and Redis.

## Notes and limitations

- Access tokens remain valid until their short TTL expires after logout; refresh tokens are revoked immediately. Lower `JWT_ACCESS_TTL` if stricter logout semantics are required. Sessions also expire after `JWT_SESSION_IDLE_TTL` (default 30 minutes) without a token refresh;
the frontend mirrors this with an inactivity timer (`VITE_SESSION_IDLE_MINUTES`) so the UI signs out immediately.
- The frontend still uses demo data; endpoints are ready to be wired through `frontend/src/lib/api.ts`.
- Email/SMS reminders, reports and background workers are not implemented yet.
- Rate limiting is per IP at the edge; Redis failures fail open for rate limiting and fall back to PostgreSQL for authorization.

