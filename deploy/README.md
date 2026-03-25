# Production Deploy

This deployment setup assumes:

- Ubuntu or Debian server
- Docker with the Compose plugin installed
- host-level `nginx`
- SSH access to the server

## What it does

- builds one production image that contains the Go backend and the built Vue frontend
- starts PostgreSQL, Redis, Kafka, and the app with `docker compose`
- applies SQL migrations from [backend/migrations](/Users/vees1de/repos/MoneyApp/backend/migrations)
- configures `nginx` as a reverse proxy to the app bound on `127.0.0.1:${APP_PORT}`

## Host nginx without Docker

If you deploy on a server without Docker, use [deploy/nginx/moneyapp.host.conf.template](/Users/vees1de/repos/MoneyApp/deploy/nginx/moneyapp.host.conf.template).

This variant assumes:

- `nginx` serves the built frontend from a directory like `/opt/moneyapp/frontend/dist`
- the Go backend listens on `127.0.0.1:${APP_PORT}`
- PostgreSQL is local or private-network only

The host template serves:

- `/` and SPA routes from the frontend build
- `/api/`, `/healthz`, `/readyz`, `/openapi.yaml`, and `/swagger` through the backend

Do not expose PostgreSQL through regular `nginx` `location` blocks. `Postgres` is not HTTP, so a site config is the wrong layer for it. If remote DB access is needed, use one of these instead:

- bind PostgreSQL to `127.0.0.1` and connect through an SSH tunnel
- allow access only from private IPs or VPN
- if you absolutely need `nginx` as a TCP proxy, use the `stream` module with strict IP allowlists, not the HTTP `server` block in `sites-available`

## First-time setup

1. Copy `deploy/.env.prod.example` to `deploy/.env.prod`.
2. Fill at least:
   - `APP_DOMAIN`
   - `APP_PORT`
   - `POSTGRES_PASSWORD`
   - `REDIS_PASSWORD`
   - `AUTH_JWT_SECRET`
3. Run:

```bash
chmod +x deploy/deploy_ssh.sh
./deploy/deploy_ssh.sh user@server /opt/moneyapp
```

## Host deploy without Docker

1. Install packages on the server:

```bash
sudo apt update
sudo apt install -y nginx postgresql postgresql-client
```

2. Copy `deploy/.env.host.example` to `deploy/.env.host` and fill it.

Important keys:

- `APP_DOMAIN`
- `SSL_PRIMARY_DOMAIN`
- `APP_PORT`
- `POSTGRES_DB`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`
- `AUTH_JWT_SECRET`

By default this host path disables Redis and Kafka.

3. Run:

```bash
chmod +x deploy/deploy_host_ssh.sh
./deploy/deploy_host_ssh.sh user@server /opt/moneyapp
```

The script will:

- build `frontend/dist` locally
- cross-compile the Go backend for Linux
- upload frontend, backend binary, and SQL migrations
- create/update the PostgreSQL role and database
- apply migrations via `psql`
- install/update the `systemd` unit
- optionally install an SSL nginx config from [deploy/nginx/moneyapp.host.ssl.conf.template](/Users/vees1de/repos/MoneyApp/deploy/nginx/moneyapp.host.ssl.conf.template)

## GitHub Actions auto deploy

Workflow: [deploy-host.yml](/Users/vees1de/repos/MoneyApp/.github/workflows/deploy-host.yml)

Trigger:

- every push to `main`
- manual start from `workflow_dispatch`

The simplest setup is to keep the entire host env file in one GitHub secret:

- `DEPLOY_ENV_HOST`

Use the contents of [deploy/.env.host.example](/Users/vees1de/repos/MoneyApp/deploy/.env.host.example#L1) as the template for that secret.

You still need these separate GitHub secrets:

- `DEPLOY_HOST`
- `DEPLOY_USER`
- `DEPLOY_SSH_KEY`

Optional SSH/runtime secrets:

- `DEPLOY_PORT`
- `DEPLOY_PATH`
- `DEPLOY_KNOWN_HOSTS`

Fallback mode without `DEPLOY_ENV_HOST`:

- the workflow can also build `deploy/.env.host` from individual secrets like `APP_DOMAIN`, `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `AUTH_JWT_SECRET`, and the other keys referenced inside the workflow

Notes:

- `DEPLOY_KNOWN_HOSTS` is recommended. You can get it with `ssh-keyscan -H your-host`.
- if `POSTGRES_USER` or `POSTGRES_PASSWORD` contains URL-sensitive characters, set `DATABASE_DSN` explicitly
- the workflow writes `deploy/.env.host`, configures SSH, and then calls [deploy_host_ssh.sh](/Users/vees1de/repos/MoneyApp/deploy/deploy_host_ssh.sh)

## HTTPS

The nginx template config only enables HTTP. After the first deploy and DNS setup, issue TLS on the server, for example:

```bash
sudo certbot --nginx -d your-domain.example.com
```
