# MoneyApp

Monorepo for the MoneyApp LMS/L&D backend rewrite.

## Structure

- [backend](/Users/vees1de/repos/MoneyApp/backend) - Go API, worker, LMS PostgreSQL migrations, modular monolith domain packages.
- [frontend](/Users/vees1de/repos/MoneyApp/frontend) - legacy finance UI that is no longer API-compatible with the backend rewrite.
- [docker-compose.yml](/Users/vees1de/repos/MoneyApp/docker-compose.yml) - PostgreSQL, migrations, API, worker.
- [scripts/start.sh](/Users/vees1de/repos/MoneyApp/scripts/start.sh) - helper for the same compose flow.

## Backend stack

- Go
- `net/http` + `chi`
- PostgreSQL
- JWT auth + RBAC
- modular monolith
- DB-backed outbox and background jobs

More details: [backend/README.md](/Users/vees1de/repos/MoneyApp/backend/README.md)

## Start

1. Copy `.env.example` to `.env`.
2. Fill at least:
   - `POSTGRES_PASSWORD`
   - `AUTH_JWT_SECRET`
3. Run:

```bash
./scripts/start.sh
```

Or directly:

```bash
docker compose up --build -d
```

Important:
- the old finance schema is gone
- this backend expects a fresh database
- if you previously used the finance MVP, drop the old PostgreSQL volume before rollout
