# Personal Life OS / Finance Core frontend

Mobile-first Vue 3 + Pinia frontend for the MVP Finance Core of Personal Life OS.

Current state:
- foundation app shell is implemented
- auth, onboarding, dashboard, accounts, transactions, categories, savings, review, and settings pages exist
- data is currently backed by local demo state so the product flow can be exercised before backend wiring

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

1. Replace local demo state with typed HTTP client + query layer.
2. Wire Telegram/Yandex auth to the Go backend.
3. Add real PWA setup, runtime caching, and offline outbox.
4. Expand dashboard and weekly review with backend-driven summaries.
