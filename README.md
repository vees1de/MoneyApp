# MoneyApp

Personal Life OS finance MVP in monorepo format.

## Structure

- [backend](/Users/vees1de/repos/MoneyApp/backend) - Go HTTP API, PostgreSQL migrations, auth, finance core, savings, weekly review, dashboard.
- [frontend](/Users/vees1de/repos/MoneyApp/frontend) - Vue 3 + Vite client.
- [docker-compose.yml](/Users/vees1de/repos/MoneyApp/docker-compose.yml) - local infrastructure and backend container wiring.

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

## Local start

1. Copy `.env.example` to `.env` in the repo root and adjust values if needed.
2. Start infrastructure:

```bash
docker compose up -d postgres redis kafka kafka-ui
```

3. Run backend:

```bash
cd backend
go run ./cmd/api
```

4. Run frontend:

```bash
cd frontend
npm install
npm run dev
```

## Git

Repository initialized with `main` as the default branch.
