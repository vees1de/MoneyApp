# Core Auth

## Canonical Source
Backend is the source of truth.

## Bootstrap Contract
1. `POST https://bims.su/api/v1/auth/login`
2. `GET https://bims.su/api/v1/auth/me`

`auth/me` payload drives:
- `user.roles[]`
- `user.permissions[]`
- `user.employee_profile`

## Implemented Services
- `AuthApiService` (`login`, `me`)
- `AuthSessionService` (localStorage tokens + user snapshot)
- `AuthBootstrapService` (restore user from token/snapshot on app start)
- `authHttpInterceptor` (Bearer token + global `401` handling)

## Guards
- `authGuard`: blocks protected routes for anonymous users.
- `guestGuard`: blocks `/login` for authenticated users (redirects to `/dashboard`).
- `roleGuard`: role route protection.
- `permissionGuard`: permission route protection.
- `dashboardRedirectGuard`: resolves `/dashboard` to role dashboard.

## 401 Handling
- clear session
- clear current user
- redirect to `/login`

## Login Flow (current)
1. User submits email/password.
2. Frontend calls `POST /auth/login`.
3. Tokens are saved in localStorage.
4. Frontend calls `GET /auth/me`.
5. `AuthStateService` receives user payload.
6. User snapshot is saved to localStorage for fast restore on reload.
7. Redirect to `/dashboard` (then role redirect guard).

## Reload Persistence
- On app init, if access token exists, frontend hydrates user from localStorage snapshot.
- Then it refreshes user by calling `GET /auth/me`.
- Session is cleared only on `401`; transient backend/network errors do not force logout.

## Deferred
- refresh-token automation
- logout API flow integration in UI
- unit tests
