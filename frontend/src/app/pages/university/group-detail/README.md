# Page: University Group Detail

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Детали конкретной группы с участниками и расписанием.

## Зависимости
### Features
- `corporate-university`
- `calendar-sync`
- `notifications`

### Entities
- `user`
- `calendar-event`
- `course`
- `notification`

## Импорты (концептуально)
```ts
import { CorporateUniversityFacade } from '@app/features/corporate-university';
import { CalendarSyncFacade } from '@app/features/calendar-sync';
import { NotificationsFacade } from '@app/features/notifications';
import type { User } from '@app/entities/user';
import type { CalendarEvent } from '@app/entities/calendar-event';
import type { Course } from '@app/entities/course';
import type { Notification } from '@app/entities/notification';
```

## Что происходит на странице
- Показывает участников и посещаемость.
- Управляет расписанием с синхронизацией календаря.
- Рассылает уведомления об изменениях.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
