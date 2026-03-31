# Page: External Request Detail

## Route
`/external-requests/:requestId`

## Что реализовано
- Загрузка заявки по id.
- Ролевые действия по state-machine:
  - employee: `submit` для `draft|revision_requested`
  - manager: `approve/reject/request_revision` для `manager_approval`
  - hr: `approve/reject/request_revision` для `hr_approval`
- Комментарий к решению отправляется в backend.
- Бюджетная проверка в MVP отмечена как ручная (решение HR).
