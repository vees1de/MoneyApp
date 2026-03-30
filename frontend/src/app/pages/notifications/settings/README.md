# Page: Notification Settings

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Настройка каналов и частоты уведомлений.

## Зависимости
### Features
- `notifications`
- `profile-settings`

### Entities
- `notification`
- `user`

## Импорты (концептуально)
```ts
import { NotificationsFacade } from '@app/features/notifications';
import { ProfileSettingsFacade } from '@app/features/profile-settings';
import type { Notification } from '@app/entities/notification';
import type { User } from '@app/entities/user';
```

## Что происходит на странице
- Позволяет включать/выключать каналы отправки.
- Настраивает quiet-hours и критичные исключения.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
