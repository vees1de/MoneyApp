# Page: Notifications Center

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Центр уведомлений пользователя.

## Зависимости
### Features
- `notifications`

### Entities
- `notification`

## Импорты (концептуально)
```ts
import { NotificationsFacade } from '@app/features/notifications';
import type { Notification } from '@app/entities/notification';
```

## Что происходит на странице
- Показывает ленту уведомлений по приоритетам.
- Отмечает уведомления прочитанными и фильтрует по типам.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
