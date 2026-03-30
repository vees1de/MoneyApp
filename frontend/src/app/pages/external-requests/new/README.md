# Page: External Request Create

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Создание новой заявки на внешний курс.

## Зависимости
### Features
- `external-course-requests`
- `calendar-sync`

### Entities
- `request`
- `budget-limit`
- `calendar-event`

## Импорты (концептуально)
```ts
import { ExternalCourseRequestsFacade } from '@app/features/external-course-requests';
import { CalendarSyncFacade } from '@app/features/calendar-sync';
import type { Request } from '@app/entities/request';
import type { BudgetLimit } from '@app/entities/budget-limit';
import type { CalendarEvent } from '@app/entities/calendar-event';
```

## Что происходит на странице
- Собирает данные о курсе, стоимости и сроках.
- Выполняет предварительную проверку конфликта расписания.
- Отправляет заявку в workflow согласования.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
