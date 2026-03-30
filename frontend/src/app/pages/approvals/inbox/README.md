# Page: Approvals Inbox

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Очередь задач согласования для руководителя/HR.

## Зависимости
### Features
- `approvals`
- `notifications`

### Entities
- `approval-step`
- `request`
- `notification`

## Импорты (концептуально)
```ts
import { ApprovalsFacade } from '@app/features/approvals';
import { NotificationsFacade } from '@app/features/notifications';
import type { ApprovalStep } from '@app/entities/approval-step';
import type { Request } from '@app/entities/request';
import type { Notification } from '@app/entities/notification';
```

## Что происходит на странице
- Показывает задачи к согласованию с приоритетами.
- Дает быстрые действия approve/reject/request_changes.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
