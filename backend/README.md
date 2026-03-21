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

1. Start PostgreSQL locally. The simplest option is from the repo root:

```bash
docker compose up -d postgres
```

2. Apply SQL files from [migrations](/Users/vees1de/repos/MoneyApp/backend/migrations).

3. Export local env vars. For local `go run`, use `localhost` in `DATABASE_DSN`:

```bash
export APP_ENV=development
export HTTP_ADDR=:8080
export DATABASE_DSN='postgres://postgres:postgres@localhost:5432/moneyapp?sslmode=disable'
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

Server health checks:

- `GET /healthz`
- `GET /readyz`

Main API base path:

- `/api/v1`

Swagger:

- `GET /swagger`
- raw spec: `GET /openapi.yaml`

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
