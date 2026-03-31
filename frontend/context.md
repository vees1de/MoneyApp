# Project Context (Widget-Based MVP)

## Key Navigation Decisions
- `"" -> /dashboard`
- `/dashboard` is role-aware redirect (`admin|hr|manager|trainer|employee`)
- `/login` is guest-only (authorized user is redirected to `/dashboard`)
- Main exploration goes through dashboard widgets (no sidebar)
- Protected unknown routes redirect to `/dashboard`

## Role Rule
- One user has exactly one role (backend still returns array, frontend uses first role as primary).

## Header for protected routes
- dark top bar
- back button is hidden on dashboard routes (`/dashboard*`) and shown on inner pages
- custom notifications panel with mark-as-read action
- profile link

## Dashboard Strategy
- Real role dashboards are now the primary entry:
  - `/dashboard/hr` -> widgets `2,4,7`
  - `/dashboard/manager` -> widgets `1,2,3,4,5,6,7`
  - `/dashboard/employee` -> widgets `1,2,3,5,6,7`
  - `/dashboard/trainer` -> widgets `1,2,3,4,5,6,7`
- `/dashboard/test-role` is a playground page with all widgets, without role switcher.

## Widget Numbering
1. Team overview
2. Upcoming events
3. Jira summary
4. Course requests
5. Current learning
6. Recommendations
7. Quick actions

## Widget Data Integration Status
- Connected to backend now:
  - team overview (`auth/me` + manager dashboard)
  - upcoming events (`GET /calendar/events/upcoming`)
  - jira summary (`GET /jira/board-summary`)
  - course requests:
    - employee: `GET /external-requests?scope=my&status=manager_approval&status=hr_approval`
    - fallback if filters endpoint unavailable: `GET /external-requests/my` + frontend status filter
    - approver roles: `GET /external-requests/pending-approvals`
  - current learning (`GET /learning-plan/my`)
  - recommendations (`GET /recommendations/courses`)
  - quick actions counters (`GET /learning-plan/my` + `GET /external-requests/my`)
  - my requests (`GET /external-requests/my`)
  - work activity (`GET /notifications`)

## Auth implementation status
- login page with email/password form is implemented
- explicit field validation messages are shown under each input
- auth API service + session storage + bootstrap + auth interceptor are implemented
- global `401` handler: clear session + clear user + redirect `/login`
- redirect after successful login is implemented
- refresh/logout/tests deferred by request

## API layer status
- API base URL unified: `https://bims.su/api`
- Added domain services under `src/app/core/api` aligned to backend reference
- Added typed contracts (`src/app/core/api/contracts.ts`) for dashboard/widget payloads

## Design baseline
- corporate clean style
- cold "diamond" palette on white base
- subtle gradients and minimal shadows
- no dark theme planned

## Styling architecture
- design tokens: `src/styles/tokens.scss`
- material overrides: `src/styles/material-overrides.scss`
- global entry: `src/styles.scss`
