# Page: Calendar Overview

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Единый календарь учебных активностей и дедлайнов.

## Зависимости
### Features
- `calendar-sync`
- `enrollments-progress`

### Entities
- `calendar-event`
- `enrollment`
- `course`

## Импорты (концептуально)
```ts
import { CalendarSyncFacade } from '@app/features/calendar-sync';
import { EnrollmentsProgressFacade } from '@app/features/enrollments-progress';
import type { CalendarEvent } from '@app/entities/calendar-event';
import type { Enrollment } from '@app/entities/enrollment';
import type { Course } from '@app/entities/course';
```

## Что происходит на странице
- Показывает события обучения по дням/неделям.
- Подсвечивает дедлайны и пересечения.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
