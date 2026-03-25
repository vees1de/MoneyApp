# Simple Deploy

Deployment is reduced to one Docker Compose stack:

- `postgres`
- `migrate`
- `backend`

The backend image builds the Vue frontend and serves the built SPA itself.

## Manual deploy on a server

1. Copy [.env.example](/Users/vees1de/repos/MoneyApp/.env.example) to `.env`.
2. Fill at least:
   - `POSTGRES_PASSWORD`
   - `AUTH_JWT_SECRET`
3. Run on the server:

```bash
docker compose up --build -d
```

## Deploy over SSH

If you want the repo and `.env` pushed to a server automatically, use:

```bash
chmod +x deploy/deploy_ssh.sh
./deploy/deploy_ssh.sh user@server /opt/moneyapp
```

The script uploads the repository, uploads the root `.env`, and runs `docker compose up --build -d` remotely.
