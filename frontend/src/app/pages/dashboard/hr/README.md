# Page: HR Dashboard

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Операционный кабинет HR/L&D по заявкам, бюджетам и SLA.

## Зависимости
### Features
- `external-course-requests`
- `approvals`
- `reports-analytics`

### Entities
- `request`
- `budget-limit`
- `approval-step`

## Импорты (концептуально)
```ts
import { ExternalCourseRequestsFacade } from '@app/features/external-course-requests';
import { ApprovalsFacade } from '@app/features/approvals';
import { ReportsAnalyticsFacade } from '@app/features/reports-analytics';
import type { Request } from '@app/entities/request';
import type { BudgetLimit } from '@app/entities/budget-limit';
import type { ApprovalStep } from '@app/entities/approval-step';
```

## Что происходит на странице
- Показывает поток заявок и узкие места процесса.
- Показывает план/факт бюджета.
- Показывает просроченные этапы согласования.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
