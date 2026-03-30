# Page: Calendar Conflicts

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Список конфликтов расписания и варианты их решения.

## Зависимости
### Features
- `calendar-sync`
- `external-course-requests`

### Entities
- `calendar-event`
- `request`
- `user`

## Импорты (концептуально)
```ts
import { CalendarSyncFacade } from '@app/features/calendar-sync';
import { ExternalCourseRequestsFacade } from '@app/features/external-course-requests';
import type { CalendarEvent } from '@app/entities/calendar-event';
import type { Request } from '@app/entities/request';
import type { User } from '@app/entities/user';
```

## Что происходит на странице
- Показывает конфликтующие события и уровень критичности.
- Предлагает переназначение/эскалацию.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
