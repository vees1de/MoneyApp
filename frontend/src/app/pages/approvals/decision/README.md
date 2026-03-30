# Page: Approval Decision

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Принятие решения по конкретной заявке.

## Зависимости
### Features
- `approvals`
- `external-course-requests`

### Entities
- `approval-step`
- `request`
- `budget-limit`

## Импорты (концептуально)
```ts
import { ApprovalsFacade } from '@app/features/approvals';
import { ExternalCourseRequestsFacade } from '@app/features/external-course-requests';
import type { ApprovalStep } from '@app/entities/approval-step';
import type { Request } from '@app/entities/request';
import type { BudgetLimit } from '@app/entities/budget-limit';
```

## Что происходит на странице
- Показывает контекст заявки и бюджетные лимиты.
- Фиксирует решение и комментарий согласующего.
- Обновляет следующий шаг маршрута.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
