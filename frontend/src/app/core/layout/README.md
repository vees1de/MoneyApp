# Core Layout

## Purpose
Protected shell with header-only navigation for desktop MVP.

## Header
- Dark corporate top bar
- Conditional back button (hidden on `/dashboard/test-role`)
- Custom notifications panel (not Material menu)
- `Прочитать` action per notification item
- Profile link

## Navigation model
- No sidebar.
- Main movement through dashboard widgets.
- Primary route: `/dashboard/test-role`.

## Rendering
All protected routes render through `AppShellComponent` + `<router-outlet>`.
