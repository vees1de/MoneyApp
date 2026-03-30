# Page: Catalog List

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Каталог курсов с поиском, фильтрами и сортировкой.

## Зависимости
### Features
- `learning-catalog`
- `profile-settings`

### Entities
- `course`
- `learning-direction`
- `user`

## Импорты (концептуально)
```ts
import { LearningCatalogFacade } from '@app/features/learning-catalog';
import { ProfileSettingsFacade } from '@app/features/profile-settings';
import type { Course } from '@app/entities/course';
import type { LearningDirection } from '@app/entities/learning-direction';
import type { User } from '@app/entities/user';
```

## Что происходит на странице
- Загружает список курсов по фильтрам.
- Отображает карточки и быстрые действия.
- Учитывает локаль/часовой пояс в отображении дат.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
