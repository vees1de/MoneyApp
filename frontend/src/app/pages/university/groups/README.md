# Page: University Groups

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Список групп и потоков внутреннего обучения.

## Зависимости
### Features
- `corporate-university`
- `calendar-sync`

### Entities
- `course`
- `calendar-event`
- `user`

## Импорты (концептуально)
```ts
import { CorporateUniversityFacade } from '@app/features/corporate-university';
import { CalendarSyncFacade } from '@app/features/calendar-sync';
import type { Course } from '@app/entities/course';
import type { CalendarEvent } from '@app/entities/calendar-event';
import type { User } from '@app/entities/user';
```

## Что происходит на странице
- Показывает группы, тренера и график.
- Отображает ближайшие сессии и статус набора.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
