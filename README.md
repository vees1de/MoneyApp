# MoneyApp

Personal Life OS finance MVP in monorepo format.

## Structure

- [backend](/Users/vees1de/repos/MoneyApp/backend) - Go HTTP API, PostgreSQL migrations, auth, finance core, savings, weekly review, dashboard.
- [frontend](/Users/vees1de/repos/MoneyApp/frontend) - Vue 3 + Vite client.
- [docker-compose.yml](/Users/vees1de/repos/MoneyApp/docker-compose.yml) - one-stack app startup with PostgreSQL, migrations, and backend.
- [scripts/start.sh](/Users/vees1de/repos/MoneyApp/scripts/start.sh) - helper that runs the same compose flow.

## Backend stack

- Go
- `net/http` + `chi`
- PostgreSQL
- JWT auth
- modular monolith architecture

More details: [backend/README.md](/Users/vees1de/repos/MoneyApp/backend/README.md)

## Frontend stack

- Vue 3
- Vite
- Pinia
- Vue Router
- Vitest + Playwright

## Start

1. Copy `.env.example` to `.env`.
2. Fill at least:
   - `POSTGRES_PASSWORD`
   - `AUTH_JWT_SECRET`
3. Run:

```bash
./scripts/start.sh
```

The same flow without the helper:

```bash
docker compose up --build -d
```

The frontend is built into the backend image and can also be copied to `/opt/moneyapp/frontend/dist` for host nginx via `./scripts/build_frontend_dist.sh`. The backend is published only on `127.0.0.1:${APP_PORT}` so it can sit behind host nginx safely. By default the app is available on `http://localhost:8080`.

## Production deploy

The deploy flow is the same compose stack. Manual and SSH-assisted variants are documented in [deploy/README.md](/Users/vees1de/repos/MoneyApp/deploy/README.md).
