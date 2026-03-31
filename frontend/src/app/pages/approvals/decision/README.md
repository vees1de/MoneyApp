# Page: Approvals Decision

## Route
`/approvals/:requestId`

## Что реализовано
- Загрузка заявки по id.
- Действия согласующего: `approve`, `reject`, `request_revision`.
- Доступность действий определяется state-machine по роли и статусу.
- Комментарий к решению отправляется в backend.
