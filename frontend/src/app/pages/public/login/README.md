# Page: Login

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Вход пользователя в систему и инициализация его персональных настроек.

## Зависимости
### Features
- `auth`
- `profile-settings`

### Entities
- `user`

## Импорты (концептуально)
```ts
import { AuthFacade } from '@app/features/auth';
import { ProfileSettingsFacade } from '@app/features/profile-settings';
import type { User } from '@app/entities/user';
```

## Что происходит на странице
- Показывает форму входа и валидацию полей.
- Вызывает auth-флоу и получает текущего пользователя.
- Применяет locale/theme/timezone после успешного входа и редиректит в ролевой dashboard.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
