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

## HTTPS

The nginx template config only enables HTTP. After the first deploy and DNS setup, issue TLS on the server, for example:

```bash
sudo certbot --nginx -d your-domain.example.com
```
