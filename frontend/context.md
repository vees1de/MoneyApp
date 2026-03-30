# Project Context (MVP)

## Product Baseline
- Формат: MVP за 2 дня.
- Роли: `ADMINISTRATOR`, `HR_LND`, `MANAGER`, `EMPLOYEE`, `TRAINER`.
- Язык: только RU.
- Тема: одна.
- Канал уведомлений: in-app.
- Экспорт отчетов: Excel.

## Flow
1. Пользователь проходит логин.
2. Backend возвращает `roles[]` и `features[]`.
3. Пользователь попадает на защищенную `home`.
4. `home` показывает ролевые рабочие столы и доступные фичи.
5. Переходы ограничиваются guard-ами по role/feature.

## Routing Decisions
- `""` -> `/home`.
- `home` не публичная (под `authGuard`).
- Публичные страницы: `login`, `forbidden`, `error`, `not-found`.
- Неизвестный URL -> `/dashboard`.
- `/dashboard` сейчас ведет на home-экземпляр (главная защищенная точка входа).

## Layout Decisions
- Защищенная зона работает через `AppShell`.
- В shell есть:
  - `Sidebar` (динамический по `features[]`)
  - кнопка `Вернуться назад`
  - контент через `<router-outlet>`
- На `login` shell не отображается.

## Access Control
- `authGuard`: проверка аутентификации.
- `roleGuard`: проверка доступа по ролям.
- `featureGuard`: проверка доступа по feature enum.
- При отсутствии доступа -> `/forbidden`.

## Current Technical Assumptions
- Auth/user state пока замокан в `AuthStateService`.
- Интеграции (Outlook, workflow, аналитика) подключаются поверх текущего каркаса.
- Все страницы подключены через lazy `loadComponent`.

## Open Questions (Next Iteration)
- Детальная карта feature-per-role (пункты sidebar и виджеты home).
- Приоритет/дефолтный рабочий стол при множественных ролях.
- Контракт реального backend для auth/session/permissions.
- Состав "последних действий" и источник данных.
