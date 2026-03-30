# Page: Employee Dashboard

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Личный кабинет сотрудника с обучением, дедлайнами и заявками.

## Зависимости
### Features
- `enrollments-progress`
- `external-course-requests`
- `notifications`

### Entities
- `user`
- `enrollment`
- `request`
- `notification`

## Импорты (концептуально)
```ts
import { EnrollmentsProgressFacade } from '@app/features/enrollments-progress';
import { ExternalCourseRequestsFacade } from '@app/features/external-course-requests';
import { NotificationsFacade } from '@app/features/notifications';
import type { User } from '@app/entities/user';
import type { Enrollment } from '@app/entities/enrollment';
import type { Request } from '@app/entities/request';
import type { Notification } from '@app/entities/notification';
```

## Что происходит на странице
- Показывает мои курсы и ближайшие дедлайны.
- Показывает статусы заявок на внешние курсы.
- Отображает критичные напоминания и быстрые действия.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
