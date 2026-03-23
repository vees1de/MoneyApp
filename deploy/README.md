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
