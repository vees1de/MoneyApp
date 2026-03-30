# Project Context (Widget-Based MVP)

## Key Navigation Decisions
- `"" -> /dashboard`
- `/dashboard -> /dashboard/test-role`
- Main exploration goes through dashboard widgets (no sidebar)
- Protected unknown routes redirect to `/dashboard`

## Dashboard Concept
- `dashboard/test-role` is a showcase route with all available widgets
- role-specific dashboards remain as separate routes (`/dashboard/hr`, etc.)
- widgets are hardcoded per role page for now

## Header for protected routes
- notifications dropdown
- profile link
- back button

## Widget Set (current)
- Team overview
- Upcoming events
- Jira summary
- Course requests (unsolved approvals)
- Current learning (`enrollments/my` target)
- Recommended courses (mock)
- Quick actions
- My requests
- Work activity
- Skillgraph excluded for now

## Data strategy (current stage)
- layout and flow first
- real backend binding later
- block errors are shown as text (`Data did not load :(`)
- skeleton states planned next step

## Backend-first notes
- permissions come from `/api/v1/auth/me` as `user.permissions[]`
- roles come from `/api/v1/auth/me` as `user.roles[]`
- departments are currently available only through profile/related entities
