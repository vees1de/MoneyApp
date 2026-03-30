# Page: Admin Learning Directions

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Управление направлениями и учебными программами.

## Зависимости
### Features
- `admin`
- `corporate-university`

### Entities
- `learning-direction`
- `course`

## Импорты (концептуально)
```ts
import { AdminFacade } from '@app/features/admin';
import { CorporateUniversityFacade } from '@app/features/corporate-university';
import type { LearningDirection } from '@app/entities/learning-direction';
import type { Course } from '@app/entities/course';
```

## Что происходит на странице
- Создает/редактирует направления.
- Привязывает курсы к направлениям.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
