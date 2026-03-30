# Page: Admin Users

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Управление пользователями и их привязкой к структуре.

## Зависимости
### Features
- `admin`
- `auth`

### Entities
- `user`
- `department`

## Импорты (концептуально)
```ts
import { AdminFacade } from '@app/features/admin';
import { AuthFacade } from '@app/features/auth';
import type { User } from '@app/entities/user';
import type { Department } from '@app/entities/department';
```

## Что происходит на странице
- Поиск и редактирование карточек пользователей.
- Изменение роли и организационных атрибутов.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
