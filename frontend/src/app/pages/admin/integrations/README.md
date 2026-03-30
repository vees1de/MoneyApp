# Page: Admin Integrations

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Настройка интеграций (Outlook, почта и др.).

## Зависимости
### Features
- `admin`
- `calendar-sync`
- `notifications`

### Entities
- `calendar-event`
- `notification`

## Импорты (концептуально)
```ts
import { AdminFacade } from '@app/features/admin';
import { CalendarSyncFacade } from '@app/features/calendar-sync';
import { NotificationsFacade } from '@app/features/notifications';
import type { CalendarEvent } from '@app/entities/calendar-event';
import type { Notification } from '@app/entities/notification';
```

## Что происходит на странице
- Показывает состояние коннекторов.
- Дает ручные тесты соединений и базовую диагностику.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
