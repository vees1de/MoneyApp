# Core Layout

## Назначение
Headless wireframe-layout для защищенной части приложения.

## Что сейчас реализовано
- `AppShellComponent`: sidebar + кнопка "Вернуться назад" + `<router-outlet>`.
- `SidebarComponent`: отрисовка пунктов навигации по доступным фичам пользователя.

## Поведение
- На `login` shell не показывается.
- На защищенных роутов используется shell.
- Навигация строится динамически по `features[]`.

## Что уточнить далее
- Финальная информационная архитектура sidebar для каждой роли.
- Нужны ли grouped sections и pinned shortcuts.
