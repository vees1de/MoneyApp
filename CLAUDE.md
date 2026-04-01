# MoneyApp

Corporate learning & development platform (LMS) for managing employee training, course enrollment, HR analytics, and budget tracking.

## Tech Stack

### Frontend
- **Framework:** Angular 21.2 (standalone components, signals)
- **UI:** Angular Material 21.2 + Angular CDK
- **State:** Signal-based (`signal()`, `computed()`) — no NgRx
- **Forms:** Reactive Forms (`FormBuilder`)
- **Styling:** SCSS with design tokens (`/frontend/src/styles/tokens.scss`)
- **Language:** TypeScript 5.9

### Backend
- **Language:** Go 1.23
- **Router:** chi v5
- **Database:** PostgreSQL (via `database/sql`)
- **Auth:** JWT (access + refresh tokens)
- **Middleware:** RBAC via `middleware.RBAC("permission.code")`
- **Validation:** go-playground/validator

## Project Structure

```
frontend/
  src/app/
    core/           # Auth, API services, layout, config
    entities/       # Data models (course, enrollment, user, etc.)
    features/       # Facades (state management layer)
    pages/          # Route-level page components
    widgets/        # Reusable dashboard widgets
    app.routes.ts   # All route definitions
  src/styles/       # Global SCSS, tokens, material overrides

backend/
  internal/
    app/            # DI container, router, server setup
    core/users/     # User profiles, teams, avatars
    modules/
      org/          # Employee profiles, departments
      identity/     # Auth, registration, sessions
      analytics/    # HR/manager dashboards
      smart_export/ # Dynamic Excel export (excelize)
      ...           # certificates, notifications, etc.
    middleware/     # Auth, RBAC, CORS, logging
    platform/      # DB, HTTP helpers
  migrations/       # SQL migration files
```

## Key Patterns

### Frontend
- **Path aliases:** `@core/*`, `@entities/*`, `@features/*`, `@pages/*`, `@shared/*`
- **API services:** `core/api/*.service.ts` — each domain has a service
- **Facades:** `features/*/` — inject API services, expose signals to pages
- **Pages:** `pages/domain/action/` — `.page.ts`, `.page.html`, `.page.scss`
- **Lazy loading:** All routes use `loadComponent` with dynamic imports
- **Guards:** `authGuard`, `roleGuard(['admin'])`, `permissionGuard([PERMISSIONS.x])`

### Backend
- **Router:** `/api/v1/*` namespace, chi route groups
- **Auth middleware:** `middleware.AuthRequired(jwt)` → `middleware.RBAC("perm")`
- **Handler pattern:** Decode → Validate → Service call → httpx.WriteJSON
- **Container DI:** All services wired in `internal/app/container.go`

## Commands

```bash
# Frontend
cd frontend && npm start          # ng serve (port 4200)
cd frontend && npm run build      # Production build
cd frontend && npm test           # Vitest tests

# Backend
cd backend && go run ./cmd/server # Start API server
cd backend && go test ./...       # Run all tests
```

## Permissions (used in guards & RBAC)

`users.read`, `users.write`, `roles.manage`, `courses.read`, `courses.write`, `courses.assign`, `intakes.manage`, `intakes.apply`, `enrollments.read`, `enrollments.manage`, `external_requests.*`, `certificates.verify`, `programs.manage`, `analytics.read_hr`, `analytics.read_manager`, `notifications.manage`, `settings.manage`, `audit.read`
