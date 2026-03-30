# Simple Deploy

Deployment is reduced to one Docker Compose stack:

- `postgres`
- `migrate`
- `backend`

For host nginx, the deploy flow expects the built frontend to be uploaded to `/root/MoneyApp/frontend/dist`.

## Manual deploy on a server

1. Copy [.env.example](/Users/vees1de/repos/MoneyApp/.env.example) to `.env`.
2. Fill at least:
   - `POSTGRES_PASSWORD`
   - `AUTH_JWT_SECRET`
3. Build static frontend files locally:

```bash
./scripts/build_frontend_dist.sh
```

4. Upload the built frontend from the local machine:

```bash
./deploy/build_frontend_local_to_server.sh
```

By default it uploads to `root@193.187.92.116:/root/MoneyApp/frontend/dist`.

The local build runs directly from the checked out project with `npm`, without Docker.

5. Run on the server:

```bash
docker compose up --build -d
```

This matches a host nginx config that:

- serves `/root/MoneyApp/frontend/dist`
- proxies `/api/`, `/healthz`, `/readyz`, `/openapi.yaml`, and `/swagger` to `127.0.0.1:${APP_PORT}`

## Deploy over SSH

If you want the server to pull the repo itself over git and only upload `.env`, use:

```bash
chmod +x deploy/deploy_ssh.sh
./deploy/deploy_ssh.sh
```

The script preserves the uploaded `frontend/dist`, uploads the root `.env`, and runs the compose deploy remotely.
The repo URL, branch, deploy path, and default SSH target are hardcoded for this project.

To build the static frontend locally and upload it over SSH, use:

```bash
chmod +x deploy/build_frontend_local_to_server.sh
./deploy/build_frontend_local_to_server.sh
```

To build the static frontend directly on the server over SSH, use:

```bash
chmod +x deploy/build_frontend_on_server.sh
./deploy/build_frontend_on_server.sh
```

## GitHub Actions CI/CD

Workflow: [.github/workflows/ci-cd.yml](/Users/vees1de/repos/MoneyApp/.github/workflows/ci-cd.yml)

What it does:

- on push to `main`: runs deploy only
- on `workflow_dispatch`: lets you trigger deploy manually

Required GitHub secrets for deploy:

- `DEPLOY_HOST`
- `DEPLOY_USER`
- `DEPLOY_SSH_KEY`

Recommended optional deploy secrets:

- `DEPLOY_KNOWN_HOSTS`

If the repository is private, make sure the server can authenticate to `https://github.com/vees1de/MoneyApp.git` via deploy key, SSH agent, or embedded HTTPS credentials.

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
- `YANDEX_CLIENT_SECRET`
- `VITE_TELEGRAM_BOT_USERNAME`
- `VITE_YANDEX_CLIENT_ID`
- `VITE_YANDEX_REDIRECT_URI`
