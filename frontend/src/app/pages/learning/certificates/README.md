# Page: Certificates

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Сертификаты сотрудника и статусы их верификации.

## Зависимости
### Features
- `enrollments-progress`
- `external-course-requests`

### Entities
- `certificate`
- `course`
- `request`

## Импорты (концептуально)
```ts
import { EnrollmentsProgressFacade } from '@app/features/enrollments-progress';
import { ExternalCourseRequestsFacade } from '@app/features/external-course-requests';
import type { Certificate } from '@app/entities/certificate';
import type { Course } from '@app/entities/course';
import type { Request } from '@app/entities/request';
```

## Что происходит на странице
- Показывает список сертификатов и связанный курс.
- Позволяет загрузить/обновить сертификат для внешнего курса.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
