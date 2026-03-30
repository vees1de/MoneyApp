# Core Auth

## Canonical Source
Backend is the source of truth.

## Bootstrap Contract
1. `POST /api/v1/auth/login`
2. `GET /api/v1/auth/me`

`auth/me` payload drives:
- `user.roles[]`
- `user.permissions[]`
- `user.employee_profile`

## Access Model
- UI visibility and route access are based on `permissions` from backend.
- Roles are used for role-specific dashboards and labels.
- Unknown/missing permissions result in `/forbidden`.

## Implemented in frontend
- `AuthStateService`
- `authGuard`
- `roleGuard`
- `permissionGuard` (in `feature.guard.ts`)
- `PERMISSIONS` constants mapped to seeded codes.

## Notes
- JWT + refresh flow should be handled in HTTP interceptor (next migration step).
- Logout should be best-effort API call + local session cleanup.
