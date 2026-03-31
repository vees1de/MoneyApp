# Page: Approvals Inbox

## Route
`/approvals/inbox`

## Что реализовано
- Входящие задачи согласования через `GET /api/v1/external-requests/pending-approvals`.
- Карточки задач с переходом в `/approvals/:requestId`.
- Отображение текущего шага согласования и статуса заявки.
