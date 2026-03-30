# Page: Settings Security

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Безопасность профиля: сессии, пароль, подтверждения.

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
- Показывает активные сессии и параметры безопасности.
- Позволяет завершить чужие сессии и обновить секреты доступа.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
