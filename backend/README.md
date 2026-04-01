# MoneyApp Backend

Go backend for the LMS/L&D modular monolith.

## Runtime

- Go 1.26
- `net/http` + `chi`
- PostgreSQL
- JWT access tokens + DB-backed refresh sessions
- DB-backed outbox + background jobs
- `cmd/api` for HTTP
- `cmd/worker` for jobs, notifications, integrations, exports

## Domain modules

- `identity`
- `org`
- `admin`
- `catalog`
- `learning`
- `testing`
- `certificates`
- `external_training`
- `outlook`
- `notifications`
- `university`
- `analytics`
- `audit`

## Database

The old finance schema was removed. The backend now expects a fresh PostgreSQL database and applies the LMS baseline migrations from [migrations](/Users/vees1de/repos/MoneyApp/backend/migrations).

Important rollout note:
- this is a full product pivot
- old finance data is not migrated
- use a fresh database or drop the old volume before startup

## Local run

1. Start PostgreSQL:

```bash
docker compose up -d postgres
```

2. Apply migrations:

```bash
docker compose run --rm migrate
```

3. Start the API:

```bash
go run ./cmd/api
```

4. Start the worker in another shell:

```bash
go run ./cmd/worker
```

Or start the whole stack:

```bash
docker compose up --build -d backend worker
```

## API

Main API base path:
- `/api/v1`

Key groups:
- `/auth`
- `/admin/users`
- `/courses`
- `/assignments`
- `/enrollments`
- `/tests`
- `/certificates`
- `/external-requests`
- `/approval-workflows`
- `/budget-limits`
- `/integrations/outlook`
- `/integrations/yougile`
- `/integrations/github`
- `/notifications`
- `/programs`
- `/analytics`
- `/audit-logs`

Health:
- `GET /healthz`
- `GET /readyz`

Docs:
- `GET /swagger`
- `GET /swagger.json`
- `GET /openapi.yaml`

Frontend integration docs:
- [FRONTEND_API_GUIDE.md](/Users/vees1de/repos/MoneyApp/backend/FRONTEND_API_GUIDE.md)
- Source of truth for generated clients: `GET /openapi.yaml`

## Notes

- Numeric values are returned as strings.
- Timestamps are RFC3339.
- The current frontend in this repository still targets the legacy finance product and is not API-compatible with this backend rewrite.
