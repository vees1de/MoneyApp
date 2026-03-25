# MoneyApp Backend

Go backend for the Personal Life OS finance MVP.

## Stack

- Go 1.26
- `net/http` + `chi`
- PostgreSQL
- Redis
- Kafka
- `pgx` stdlib driver
- `slog`
- `validator`
- JWT access tokens + refresh sessions in DB

## Run

1. Start infrastructure locally. The simplest option is from the repo root:

```bash
docker compose up -d postgres redis kafka kafka-ui
```

2. Apply SQL files from [migrations](/Users/vees1de/repos/MoneyApp/backend/migrations).

3. Export local env vars. For local `go run`, use `localhost` in `DATABASE_DSN`:

```bash
export APP_ENV=development
export HTTP_ADDR=:8080
export FRONTEND_DIST_DIR='../frontend/dist'
export DATABASE_DSN='postgres://postgres:postgres@localhost:5432/moneyapp?sslmode=disable'
export REDIS_ENABLED='true'
export REDIS_ADDR='localhost:6379'
export REDIS_PASSWORD='redis'
export REDIS_DB='0'
export REDIS_DASHBOARD_TTL='30s'
export KAFKA_ENABLED='true'
export KAFKA_BROKERS='localhost:9094'
export KAFKA_CLIENT_ID='moneyapp-backend'
export KAFKA_AUDIT_TOPIC='moneyapp.audit'
export KAFKA_WRITE_TIMEOUT='5s'
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

For host deployments without Docker, Redis and Kafka can be disabled:

```bash
export REDIS_ENABLED='false'
export KAFKA_ENABLED='false'
```

With these flags off, the app uses a no-op cache and no-op event publisher. PostgreSQL remains required.

Server health checks:

- `GET /healthz`
- `GET /readyz`

`/readyz` checks PostgreSQL and also Redis/Kafka when they are enabled.

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

- Redis is used as a short-lived cache for `GET /api/v1/dashboard/finance`.
- Kafka publishes audit/domain events for critical user actions into `KAFKA_AUDIT_TOPIC`.
- When disabled, Redis cache writes are bypassed and Kafka event publishing becomes a no-op.
