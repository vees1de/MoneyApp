# Page: External Requests List

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Реестр заявок на внешние курсы с фильтрами и статусами.

## Зависимости
### Features
- `external-course-requests`
- `approvals`

### Entities
- `request`
- `approval-step`
- `budget-limit`

## Импорты (концептуально)
```ts
import { ExternalCourseRequestsFacade } from '@app/features/external-course-requests';
import { ApprovalsFacade } from '@app/features/approvals';
import type { Request } from '@app/entities/request';
import type { ApprovalStep } from '@app/entities/approval-step';
import type { BudgetLimit } from '@app/entities/budget-limit';
```

## Что происходит на странице
- Показывает список заявок по ролям и фильтрам.
- Показывает текущий шаг согласования и бюджетный контекст.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
