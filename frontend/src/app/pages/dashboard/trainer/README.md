# Page: Trainer Dashboard

## Route
$(System.Collections.Hashtable.Route)

## Назначение
ЛК внутреннего тренера: группы, расписание, обратная связь.

## Зависимости
### Features
- `corporate-university`
- `calendar-sync`
- `notifications`

### Entities
- `course`
- `calendar-event`
- `notification`

## Импорты (концептуально)
```ts
import { CorporateUniversityFacade } from '@app/features/corporate-university';
import { CalendarSyncFacade } from '@app/features/calendar-sync';
import { NotificationsFacade } from '@app/features/notifications';
import type { Course } from '@app/entities/course';
import type { CalendarEvent } from '@app/entities/calendar-event';
import type { Notification } from '@app/entities/notification';
```

## Что происходит на странице
- Показывает ближайшие занятия и группы.
- Показывает задачи по проверке/фидбеку.
- Показывает изменения расписания.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
