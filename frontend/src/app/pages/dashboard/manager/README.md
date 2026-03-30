# Page: Manager Dashboard

## Route
$(System.Collections.Hashtable.Route)

## Назначение
ЛК руководителя с очередью согласований и прогрессом команды.

## Зависимости
### Features
- `approvals`
- `reports-analytics`
- `enrollments-progress`

### Entities
- `user`
- `approval-step`
- `enrollment`

## Импорты (концептуально)
```ts
import { ApprovalsFacade } from '@app/features/approvals';
import { ReportsAnalyticsFacade } from '@app/features/reports-analytics';
import { EnrollmentsProgressFacade } from '@app/features/enrollments-progress';
import type { User } from '@app/entities/user';
import type { ApprovalStep } from '@app/entities/approval-step';
import type { Enrollment } from '@app/entities/enrollment';
```

## Что происходит на странице
- Показывает задачи согласования.
- Показывает completion/overdue по подразделению.
- Дает переходы к деталям команды и заявок.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
