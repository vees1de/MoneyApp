# Personal Life OS / Finance Core frontend

Mobile-first Vue 3 + Pinia frontend for the MVP Finance Core of Personal Life OS.

Current state:
- foundation app shell is implemented
- auth, onboarding, dashboard, accounts, transactions, categories, savings, review, and settings pages exist
- frontend is wired to the Go backend API and uses dev-auth for local login
- Vite dev server proxies `/api`, `/healthz`, `/readyz`, `/openapi.yaml`, and `/swagger` to the backend

## Structure

```text
src/
  app/        # bootstrap, layouts, router, global stores
  pages/      # route-level pages
  widgets/    # product sections for dashboard/review/finance
  features/   # user actions and forms
  entities/   # domain types
  shared/     # styles, ui-kit, api base, libs, demo data
```

## Scripts

```sh
npm install
npm run dev
npm run build
npm run test:unit
```

## Next implementation stages

1. Add unit/integration coverage around the API-backed stores.
2. Replace local token storage with a more secure cookie-based refresh flow when backend supports it.
3. Add real PWA runtime caching and offline outbox.
4. Expand dashboard and weekly review UX with richer backend payloads.
