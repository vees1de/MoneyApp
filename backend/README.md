# MoneyApp Backend

Go backend for the Personal Life OS finance MVP.

## Stack

- Go 1.26
- `net/http` + `chi`
- PostgreSQL
- `pgx` stdlib driver
- `slog`
- `validator`
- JWT access tokens + refresh sessions in DB

## Run

1. Start PostgreSQL:

```bash
docker compose up -d postgres
```

2. Apply SQL files from [migrations](/Users/vees1de/repos/MoneyApp/backend/migrations).

3. Export local env vars:

```bash
export APP_ENV=development
export HTTP_ADDR=:8080
export FRONTEND_DIST_DIR='../frontend/dist'
export POSTGRES_HOST='localhost'
export POSTGRES_PORT='5432'
export POSTGRES_DB='moneyapp'
export POSTGRES_USER='postgres'
export POSTGRES_PASSWORD='postgres'
export AUTH_JWT_SECRET='local-dev-secret'
export AUTH_JWT_ISSUER='moneyapp'
export AUTH_ACCESS_TOKEN_TTL='15m'
export AUTH_REFRESH_TOKEN_TTL='720h'
export AUTH_ALLOW_INSECURE_DEV_AUTH='true'
```

4. Start the API:

```bash
go run ./cmd/api
```

Optional frontend integration:

```bash
cd ../frontend
npm install
npm run build
cd ../backend
export FRONTEND_DIST_DIR='../frontend/dist'
go run ./cmd/api
```

With `FRONTEND_DIST_DIR` set, the backend serves the built SPA and falls back to `index.html` for non-API routes.

Server health checks:

- `GET /healthz`
- `GET /readyz`

`/readyz` checks PostgreSQL.

Main API base path:

- `/api/v1`

Swagger:

- `GET /swagger`
- `GET /swagger.json`
- raw spec: `GET /openapi.yaml`

Swagger generation:

```bash
go generate ./internal/docs
```

Or for the whole backend:

```bash
go generate ./...
```

## Quick API test

1. Open `http://localhost:8080/swagger`
2. Call `POST /api/v1/auth/telegram`
3. Use a local dev payload:

```json
{
  "provider_user_id": "tg_10001",
  "username": "veeside",
  "first_name": "Vee",
  "last_name": "Side",
  "auth_date": 1711111111,
  "hash": "dev-mode"
}
```

4. Copy `tokens.access_token` from the response
5. Click `Authorize` in Swagger UI and paste:

```text
Bearer <access_token>
```

6. Test protected endpoints like:

- `GET /api/v1/accounts`
- `POST /api/v1/finance/transactions`
- `GET /api/v1/dashboard/finance`

## Structure

- `internal/core` contains auth, users, sessions, audit, links, and health.
- `internal/modules/finance` contains accounts, categories, transactions, and summary.
- `internal/modules/review`, `internal/modules/savings`, and `internal/modules/dashboard` contain product workflows.
- `internal/platform` contains infrastructure helpers.
- `internal/app` is the composition root.

## Infra usage

- The backend only depends on PostgreSQL.
- The frontend build can be embedded into the backend container or served from `FRONTEND_DIST_DIR`.
