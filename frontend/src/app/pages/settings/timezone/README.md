# Page: Settings Timezone

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Настройка часового пояса пользователя.

## Зависимости
### Features
- `profile-settings`
- `calendar-sync`

### Entities
- `user`
- `calendar-event`

## Импорты (концептуально)
```ts
import { ProfileSettingsFacade } from '@app/features/profile-settings';
import { CalendarSyncFacade } from '@app/features/calendar-sync';
import type { User } from '@app/entities/user';
import type { CalendarEvent } from '@app/entities/calendar-event';
```

## Что происходит на странице
- Показывает текущий timezone.
- Пересчитывает локальное отображение учебных событий.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
