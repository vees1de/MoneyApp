# Page: My Learning

## Route
$(System.Collections.Hashtable.Route)

## Назначение
Список моих назначений с прогрессом и дедлайнами.

## Зависимости
### Features
- `enrollments-progress`
- `notifications`

### Entities
- `enrollment`
- `course`
- `notification`

## Импорты (концептуально)
```ts
import { EnrollmentsProgressFacade } from '@app/features/enrollments-progress';
import { NotificationsFacade } from '@app/features/notifications';
import type { Enrollment } from '@app/entities/enrollment';
import type { Course } from '@app/entities/course';
import type { Notification } from '@app/entities/notification';
```

## Что происходит на странице
- Показывает активные и завершенные назначения.
- Подсвечивает просрочки и риски дедлайнов.
- Позволяет открыть детальную страницу прохождения.

## Состояния UI
- loading: первичная загрузка данных.
- error: ошибка запроса/операции.
- empty: нет данных по текущим фильтрам.
- eady: данные загружены, доступны действия.

## Вопросы для уточнения
- Какие роли имеют доступ к странице?
- Какой минимальный набор данных обязателен для первого релиза?
- Нужны ли особые правила аудита действий на этой странице?
