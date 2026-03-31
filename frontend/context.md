# Project Context (Widget-Based MVP)

## Key Navigation Decisions
- `"" -> /dashboard`
- `/dashboard` is role-aware redirect (`admin|hr|manager|trainer|employee`)
- `/login` is guest-only (authorized user is redirected to `/dashboard`)
- Main exploration goes through dashboard widgets (no sidebar)
- Protected unknown routes redirect to `/dashboard`

## Role Rule
- One user has exactly one role (backend returns array, frontend uses first role as primary).

## Header for protected routes
- dark top bar
- brand (`ИМПУЛЬС`) routes to `/dashboard`
- back button removed from shell
- custom notifications panel with mark-as-read action
- profile link

## Dashboard Strategy
- Real role dashboards are now the primary entry:
  - `/dashboard/hr` -> widgets `2,4,7`
  - `/dashboard/manager` -> widgets `1,2,3,4,5,6,7`
  - `/dashboard/employee` -> widgets `1,2,3,5,6,7`
  - `/dashboard/trainer` -> widgets `1,2,3,4,5,6,7`
- `/dashboard/test-role` is a playground page with all widgets, without role switcher.
- `dashboard/hr` rebuilt to two-column layout:
  - main column: quick actions, requests board, process board
  - side column: upcoming calendar events

## Widget Data Integration Status
- Connected to backend now:
  - team overview (`auth/me` + manager dashboard)
  - upcoming events (`GET /calendar/events/upcoming`)
  - jira summary (`GET /jira/board-summary`)
  - course requests:
    - employee: `GET /external-requests?scope=my&status=manager_approval&status=hr_approval`
    - fallback if filters endpoint unavailable: `GET /external-requests/my` + frontend status filter
    - approver roles: `GET /external-requests/pending-approvals`
  - current learning / process (`GET /enrollments/my` + `GET /courses`)
  - recommendations (`GET /recommendations/courses`)
  - quick actions counters (`GET /learning-plan/my` + `GET /external-requests/my`)
  - my requests (`GET /external-requests/my`)
  - work activity (`GET /notifications`)

## External Request Workflow (MVP)
- Implemented domain rules in `src/app/core/domain/external-request.workflow.ts`.
- Status labels + allowed actions per role are centralized.
- Implemented pages:
  - `/external-requests` (list with filters)
  - `/external-requests/new` (create)
  - `/external-requests/:requestId` (detail + actions)
  - `/approvals/inbox` (pending approvals)
  - `/approvals/:requestId` (decision)
- Budget check in MVP is manual HR decision (no hard-stop automation).

## Learning Execution (MVP)
- `/my-learning` loads real enrollments.
- `/learning/:enrollmentId` supports start/progress/complete actions.

## Catalog Page
- `/catalog` loads real courses from `GET /api/v1/courses`.
- Filters sent as query params: `status`, `source_type`, `level`, `limit`, `offset`.

## Auth implementation status
- login page with email/password form is implemented
- explicit field validation messages are shown under each input
- auth API service + session storage + bootstrap + auth interceptor are implemented
- auth persistence on reload is implemented via localStorage (tokens + user snapshot)
- global `401` handler now performs refresh-token retry; if refresh fails -> clear session + redirect `/login`
- redirect after successful login is implemented
- logout/tests deferred by request

## API layer status
- API base URL unified: `https://bims.su/api`
- Added domain services under `src/app/core/api` aligned to backend reference
- Added typed contracts (`src/app/core/api/contracts.ts`) for dashboard/widget payloads

## Design baseline
- corporate clean style
- cold "diamond" palette on white base
- subtle gradients and minimal shadows
- no dark theme planned
- new radii standard: `24px` and `40px`
- app font: `Inter_18pt-Regular.ttf` from `/public` applied globally

## Styling architecture
- design tokens: `src/styles/tokens.scss`
- material overrides: `src/styles/material-overrides.scss`
- global entry: `src/styles.scss`
- protected shell restyled with horizontal nav + icon actions in `src/app/core/layout/app-shell`
