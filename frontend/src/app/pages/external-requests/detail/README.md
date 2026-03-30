# Page: External Request Detail

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Детальная карточка заявки с timeline и связанными объектами.

## Зависимости
### Features
- `external-course-requests`
- `approvals`
- `calendar-sync`

### Entities
- `request`
- `approval-step`
- `calendar-event`
- `certificate`

## Импорты (концептуально)
```ts
import { ExternalCourseRequestsFacade } from '@app/features/external-course-requests';
import { ApprovalsFacade } from '@app/features/approvals';
import { CalendarSyncFacade } from '@app/features/calendar-sync';
import type { Request } from '@app/entities/request';
import type { ApprovalStep } from '@app/entities/approval-step';
import type { CalendarEvent } from '@app/entities/calendar-event';
import type { Certificate } from '@app/entities/certificate';
```

## Что происходит на странице
- Показывает историю статусов и решений.
- Показывает календарные резервы и дедлайны.
- Показывает сертификат по завершению обучения.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
