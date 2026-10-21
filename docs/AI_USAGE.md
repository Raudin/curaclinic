# AI Usage in CuraClinic

These notes cover where I used AI on this project, where it helped, where it got things wrong, and which decisions I made on my own judgment.

---

## 1. What did you use AI for across the project?

AI was an assistant and reviewer. I owned the design and the final decisions.

| Area | How AI was used |
|---|---|
| **System design** | I designed the system first: the entities (users, doctors, patients, admins, appointments), the RBAC model and the booking flow. Then I used AI to **validate that design** and pressure-test it. It flagged edge cases (double booking under concurrency, refresh-token reuse, what happens when Redis loses its data), and I used those to **improve the design**: for example, keeping PostgreSQL exclusion constraints as the final guard against double booking behind the Redis holds. |
| **API implementation** | I built the Go backend. AI helped me **review it against Go and REST API best practices**: package layout, error handling, interface boundaries and HTTP status codes. AI also helped me **quickly connect a React client to the finished backend**: a typed API client, the auth context and token refresh, and wiring the patient, doctor and admin pages to the real endpoints. |
| **Deployment** | AI helped with the Docker and Compose details for Dokploy: multi-stage images, nginx proxying `/api` on the same origin, and startup ordering through health checks (migrate → seed → api → web). |
| **CI/CD** | I used AI to **diagnose and fix pipeline failures quickly**: formatting and lint failures blocking CI, Go version alignment for `govulncheck`, and the sqlc drift check. I read and checked each fix before merging it. |

---

## 2. One example where an AI suggestion improved your work

**Bringing the Go backend in line with Go and API best practices.**

My first version worked, but an AI review against idiomatic Go and API conventions led to several concrete improvements:

- **One error model.** Every failure became a typed `domain.Error` with a stable code (`validation_failed`, `slot_unavailable`, `hold_expired`, …). A single function in `internal/httpx` maps those codes to HTTP status codes (400/401/403/404/409/429/500). Before that, handlers chose status codes ad hoc.
- **No leaked internals.** Internal causes are logged with the request ID, and the client gets a safe generic message.
- **Thin handlers, testable services.** Business rules moved out of Gin handlers into services that depend on small interfaces, so booking rules, holds and RBAC can be unit-tested with fakes.
- **Correct status codes and validation messages.** Validation errors return `400` with field-level details by JSON path. Database constraint violations come back as `409` instead of `500`.

The API is now more predictable for the frontend and easier to test.

---

## 3. One example where AI output was wrong or incomplete, and how you caught it

**AI suggested a polymorphic relationship for users.**

For the identity model, the AI proposed a polymorphic association: a `users` table with `userable_type` / `userable_id` columns pointing at either a doctor, patient or admin record.

I caught this while reviewing the schema against how the app actually uses the data:

- A polymorphic `userable_id` **cannot have a real foreign key**. PostgreSQL cannot enforce referential integrity on a column that points at different tables depending on another column, so integrity would have moved into application code.
- It assumes **one user = one type**. Here one person can legitimately hold several account types (a doctor can also be a patient), which a single `userable_type` cannot represent cleanly.
- Every query would need conditional joins on the type column, which makes sqlc-generated, type-safe queries harder to write.

So I rejected it and used **one-to-one extension tables** keyed by `user_id` (`patients`, `doctors`, `admins`), each with a real foreign key to `users`. The schema is in [`images/database-erd.png`](images/database-erd.png).

---

## 4. Two decisions you made without AI, and why you trusted your own judgment

### Decision 1: No polymorphic relationships

As described above, I chose plain one-to-one extension tables (`patients`, `doctors`, `admins` → `users`) instead of a polymorphic `userable` relationship.

**Why I trusted my judgment:** I know the domain and how the data will be queried. Polymorphism adds flexibility this project doesn't need, and it gives up database-enforced foreign keys. For a healthcare scheduling system, data integrity matters more than abstraction, and the simpler model is easier to reason about, migrate and query.

### Decision 2: PostgreSQL, not Redis, as the final authority on bookings

Redis slot holds give a smooth booking experience: a slot is reserved while the patient fills in the form. I decided early that **Redis would never be the source of truth**. The real guarantee is an `EXCLUDE USING gist` constraint in PostgreSQL that makes overlapping confirmed appointments for the same doctor or patient impossible, plus a transaction around every booking, reschedule and cancellation.

**Why I trusted my judgment:** a cache can be flushed, restarted or lose data, and a race can always slip between "check availability" and "insert". Correctness that depends on application timing breaks eventually. Putting the invariant in the database means double booking is impossible no matter what happens to Redis or the application layer, and a cancelled appointment frees its slot automatically.
