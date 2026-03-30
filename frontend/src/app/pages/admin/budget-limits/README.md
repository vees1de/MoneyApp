# Page: Admin Budget Limits

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Управление бюджетными лимитами на обучение.

## Зависимости
### Features
- `admin`
- `external-course-requests`

### Entities
- `budget-limit`
- `department`

## Импорты (концептуально)
```ts
import { AdminFacade } from '@app/features/admin';
import { ExternalCourseRequestsFacade } from '@app/features/external-course-requests';
import type { BudgetLimit } from '@app/entities/budget-limit';
import type { Department } from '@app/entities/department';
```

## Что происходит на странице
- Задает лимиты по подразделениям и периодам.
- Показывает фактическое потребление бюджета.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
