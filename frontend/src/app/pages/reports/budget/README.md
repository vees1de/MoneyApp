# Page: Reports Budget

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Отчет по бюджету обучения и отклонениям.

## Зависимости
### Features
- `reports-analytics`
- `external-course-requests`

### Entities
- `budget-limit`
- `request`
- `department`

## Импорты (концептуально)
```ts
import { ReportsAnalyticsFacade } from '@app/features/reports-analytics';
import { ExternalCourseRequestsFacade } from '@app/features/external-course-requests';
import type { BudgetLimit } from '@app/entities/budget-limit';
import type { Request } from '@app/entities/request';
import type { Department } from '@app/entities/department';
```

## Что происходит на странице
- Показывает план/факт бюджета.
- Показывает перерасход и источники отклонений.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
