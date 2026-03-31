# Page: External Requests List

## Route
`/external-requests`

## Что реализовано
- Реальный список заявок через `GET /api/v1/external-requests`.
- Скоуп выбирается по роли:
  - `employee|trainer` -> `scope=my`
  - `manager` -> `scope=team`
  - `hr|admin` -> `scope=all`
- Фильтры: `status[]`, `assignee`.
- Для employee фильтр "мои на согласовании" использует статусы `manager_approval`, `hr_approval`.
- Fallback для employee: если фильтруемый endpoint недоступен, используется `/external-requests/my` с фронтовым фильтром статусов.
