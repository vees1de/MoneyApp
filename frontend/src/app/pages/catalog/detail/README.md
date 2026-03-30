# Page: Catalog Detail

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Детальная карточка курса и выбор сценария записи.

## Зависимости
### Features
- `learning-catalog`
- `enrollments-progress`
- `external-course-requests`

### Entities
- `course`
- `enrollment`
- `request`

## Импорты (концептуально)
```ts
import { LearningCatalogFacade } from '@app/features/learning-catalog';
import { EnrollmentsProgressFacade } from '@app/features/enrollments-progress';
import { ExternalCourseRequestsFacade } from '@app/features/external-course-requests';
import type { Course } from '@app/entities/course';
import type { Enrollment } from '@app/entities/enrollment';
import type { Request } from '@app/entities/request';
```

## Что происходит на странице
- Показывает полное описание курса.
- Для внутреннего курса: назначение/запись.
- Для внешнего курса: переход к созданию заявки.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
