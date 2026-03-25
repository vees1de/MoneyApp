# Simple Deploy

Deployment is reduced to one Docker Compose stack:

- `postgres`
- `migrate`
- `backend`

For host nginx, the deploy flow copies the built frontend directly to `/opt/moneyapp/frontend/dist`.

## Manual deploy on a server

1. Copy [.env.example](/Users/vees1de/repos/MoneyApp/.env.example) to `.env`.
2. Fill at least:
   - `POSTGRES_PASSWORD`
   - `AUTH_JWT_SECRET`
3. Build static frontend files for nginx:

```bash
./scripts/build_frontend_dist.sh
```

4. Run on the server:

```bash
docker compose up --build -d
```

This matches a host nginx config that:

- serves `/opt/moneyapp/frontend/dist`
- proxies `/api/`, `/healthz`, `/readyz`, `/openapi.yaml`, and `/swagger` to `127.0.0.1:${APP_PORT}`

## Deploy over SSH

If you want the repo and `.env` pushed to a server automatically, use:

```bash
chmod +x deploy/deploy_ssh.sh
./deploy/deploy_ssh.sh user@server /opt/moneyapp
```

The script uploads the repository, uploads the root `.env`, copies the built frontend to `/opt/moneyapp/frontend/dist` on the server for nginx, and runs `docker compose up --build -d` remotely.

## GitHub Actions CI/CD

Workflow: [.github/workflows/ci-cd.yml](/Users/vees1de/repos/MoneyApp/.github/workflows/ci-cd.yml)

What it does:

- on every push and `pull_request`: runs backend tests, frontend build, and Docker image build
- on push to `main`: runs the same CI checks and then deploys
- on `workflow_dispatch`: lets you trigger deploy manually

Required GitHub secrets for deploy:

- `DEPLOY_HOST`
- `DEPLOY_USER`
- `DEPLOY_SSH_KEY`

Recommended optional SSH secrets:

- `DEPLOY_PORT`
- `DEPLOY_PATH`
- `DEPLOY_KNOWN_HOSTS`

Application config can be provided in one of two ways:

- recommended: `DEPLOY_ENV` containing the full contents of the root `.env`
- minimal fallback: `POSTGRES_PASSWORD` and `AUTH_JWT_SECRET`

Optional fallback secrets for customizing the generated `.env`:

- `APP_PORT`
- `POSTGRES_DB`
- `POSTGRES_USER`
- `AUTH_JWT_ISSUER`
- `AUTH_ACCESS_TOKEN_TTL`
- `AUTH_REFRESH_TOKEN_TTL`
- `AUTH_ALLOW_INSECURE_DEV_AUTH`
- `DEFAULT_BASE_CURRENCY`
- `DEFAULT_TIMEZONE`
- `DEFAULT_WEEKLY_REVIEW_HOUR`
- `TELEGRAM_BOT_TOKEN`
- `YANDEX_CLIENT_ID`
