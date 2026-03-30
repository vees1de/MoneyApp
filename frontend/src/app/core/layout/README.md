# Core Layout

## Purpose
Protected shell with header-only navigation for desktop MVP.

## Header
- Back button
- Notifications dropdown (Material menu)
- Profile link

## Navigation model
- No sidebar.
- Main movement through dashboard widgets.
- Primary route: `/dashboard/test-role`.

## Rendering
All protected routes render through `AppShellComponent` + `<router-outlet>`.
