# Widgets Layer

Purpose: standalone dashboard widgets for role pages.

Rules:
- each widget is a standalone Angular component
- each widget wraps content with `WidgetShellComponent`
- title can link to list route
- cards inside widget link to entity details
- loading/error/empty states are supported in shell
